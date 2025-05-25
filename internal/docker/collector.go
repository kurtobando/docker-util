package docker

import (
	"context"
	"encoding/binary"
	"io"
	"log"
	"sync"
	"time"

	"bw_util/internal/database"
	"bw_util/internal/models"

	"github.com/docker/docker/api/types/container"
)

// Collector manages log collection from multiple containers
type Collector struct {
	client     *Client
	repository database.LogRepository
	wg         sync.WaitGroup
}

// NewCollector creates a new log collector
func NewCollector(client *Client, repository database.LogRepository) *Collector {
	return &Collector{
		client:     client,
		repository: repository,
	}
}

// StartCollection starts log collection for all provided containers
func (c *Collector) StartCollection(ctx context.Context, containers []models.Container) error {
	for _, cont := range containers {
		c.wg.Add(1)
		go c.streamAndStoreLogs(ctx, cont.ID, cont.Name)
	}
	return nil
}

// Wait waits for all log streaming goroutines to finish
func (c *Collector) Wait() {
	c.wg.Wait()
}

// streamAndStoreLogs handles fetching, parsing, and storing logs for a single container
// This preserves the exact same logic from the original streamAndStoreLogs function
func (c *Collector) streamAndStoreLogs(ctx context.Context, containerID string, containerName string) {
	defer c.wg.Done() // Decrement counter when goroutine finishes

	// Determine a unique "since" time for this container's log stream to minimize duplicate logs across restarts IF NEEDED.
	// For now, using current time as before, or a very early time to get more history for testing.
	// A more robust approach would be to query the DB for the last timestamp for this containerID.
	sinceTime := time.Now().Add(-30 * time.Minute).Format(time.RFC3339Nano) // Get more history for UI demo

	logOpts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
		Details:    false,
		Since:      sinceTime,
	}

	logReader, err := c.client.GetClient().ContainerLogs(ctx, containerID, logOpts) // Pass cancellable context
	if err != nil {
		if ctx.Err() == context.Canceled {
			log.Printf("Log streaming setup for %s (%s) cancelled during ContainerLogs call.\n", containerName, containerID[:12])
		} else {
			log.Printf("Error getting logs for container %s (%s): %v\n", containerName, containerID[:12], err)
		}
		return
	}
	defer logReader.Close()

	hdr := make([]byte, 8)
	for {
		select {
		case <-ctx.Done(): // Check if shutdown has been signaled
			log.Printf("Shutting down log streaming for %s (%s) due to context cancellation.\n", containerName, containerID[:12])
			return
		default:
			// Non-blocking check, proceed with reading logs
		}

		// Set a deadline for the Read operation to make it interruptible by ctx.Done()
		// This is a common pattern but requires the underlying net.Conn if logReader exposes it.
		// For now, the select above is the primary cancellation check point before a potentially blocking Read.
		// A more robust solution might involve a custom reader that respects context cancellation.

		_, err := logReader.Read(hdr)
		if err != nil {
			select {
			case <-ctx.Done(): // If context was cancelled during read
				log.Printf("Log stream for %s (%s) cancelled during read.\n", containerName, containerID[:12])
				return
			default:
			}
			if err == io.EOF {
				log.Printf("Log stream ended (EOF) for container %s (%s)\n", containerName, containerID[:12])
				return // EOF means the container might have stopped, or logs ended.
			}
			log.Printf("Error reading log stream header for %s (%s): %v\n", containerName, containerID[:12], err)
			return // Return on other errors
		}

		var streamTypeStr string
		switch hdr[0] {
		case 1:
			streamTypeStr = "stdout"
		case 2:
			streamTypeStr = "stderr"
		default:
			payloadSize := binary.BigEndian.Uint32(hdr[4:])
			if payloadSize > 0 {
				content := make([]byte, payloadSize)
				_, readErr := io.ReadFull(logReader, content)
				if readErr != nil {
					log.Printf("Error reading and discarding unknown stream payload for %s: %v\n", containerName, readErr)
					return // Error during discard, exit goroutine for this container
				}
			}
			continue
		}

		payloadSize := binary.BigEndian.Uint32(hdr[4:])
		if payloadSize == 0 {
			continue
		}

		content := make([]byte, payloadSize)
		_, err = io.ReadFull(logReader, content)
		if err != nil {
			select {
			case <-ctx.Done():
				log.Printf("Log stream for %s (%s) cancelled during payload read.\n", containerName, containerID[:12])
				return
			default:
			}
			log.Printf("Error reading log message payload for %s (%s): %v\n", containerName, containerID[:12], err)
			return // Error reading payload, exit goroutine for this container
		}

		logMessage := string(content)
		actualTimestampStr, actualLogMessage := ParseLogEntry(logMessage, containerName)

		if actualLogMessage == "" {
			continue
		}

		// Store the log entry using the repository
		entry := &models.LogEntry{
			ContainerID:   containerID,
			ContainerName: containerName,
			Timestamp:     actualTimestampStr,
			StreamType:    streamTypeStr,
			LogMessage:    actualLogMessage,
		}

		if err := c.repository.Store(ctx, entry); err != nil {
			log.Printf("Failed to store log entry for %s (%s): %v", containerName, containerID[:12], err)
			// Continue processing other logs even if one fails
		}
	}
	// log.Printf("Stopped streaming logs for container %s (%s)\n", containerName, containerID[:12]) // This will be logged when loop exits
}
