package main

import (
	"context"
	"database/sql"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// dbName is now a variable that can be set by a flag
var dbName string
var db *sql.DB

func init() {
	// Define a command-line flag for the database path
	// The default value is "./docker_logs.db"
	flag.StringVar(&dbName, "dbpath", "./docker_logs.db", "Path to the SQLite database file.")
}

func initDB() error {
	var err error
	// dbName is now set by the flag, parsed in main()
	db, err = sql.Open("sqlite3", dbName)
	if err != nil {
		return fmt.Errorf("failed to open database at %s: %w", dbName, err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		container_id TEXT NOT NULL,
		container_name TEXT NOT NULL,
		timestamp TEXT NOT NULL,
		stream_type TEXT NOT NULL, 
		log_message TEXT NOT NULL
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close() // Close the DB if table creation fails
		return fmt.Errorf("failed to create logs table in %s: %w", dbName, err)
	}
	log.Printf("Database initialized at %s and logs table ensured.", dbName)
	return nil
}

func main() {
	flag.Parse() // Parse command-line flags

	// Create a context that can be cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancel is called to free resources if main exits early

	// Setup signal handling for graceful shutdown
	// Listen for SIGINT (Ctrl+C) and SIGTERM (kill)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	go func() {
		<-sigChan // Block until a signal is received
		log.Println("Shutdown signal received, initiating graceful shutdown...")
		cancel() // Cancel the context to signal goroutines to stop
	}()

	if err := initDB(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	// Defer db.Close() will be called when main exits. 
	// For graceful shutdown, we will close it explicitly after all goroutines are done.

	dli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}
	// Defer dli.Close() for the same reason as db.Close()

	containers, err := dli.ContainerList(ctx, container.ListOptions{All: false}) // Only running containers
	if err != nil {
		// If context was cancelled during list, it's part of shutdown
		if ctx.Err() == context.Canceled {
			log.Println("Context cancelled during container listing, shutting down.")
		} else {
			log.Fatalf("Error listing containers: %v", err)
		}
		// Perform cleanup before exiting due to error after initDB or client creation
		db.Close()
		if dli != nil { dli.Close() }
		return
	}

	if len(containers) == 0 {
		log.Println("No running containers found to monitor.")
		db.Close()
		dli.Close()
		return
	}

	log.Printf("Found %d running containers. Starting log streaming...\n", len(containers))

	for _, cont := range containers {
		containerName := "unknown"
		if len(cont.Names) > 0 {
			containerName = strings.TrimPrefix(cont.Names[0], "/")
		}
		log.Printf("Starting log stream for: ID=%s, Name=%s, Image=%s\n", cont.ID[:12], containerName, cont.Image)

		wg.Add(1)
		go streamAndStoreLogs(ctx, dli, cont.ID, containerName, &wg) // Pass the cancellable context and WaitGroup
	}

	// Wait for either the context to be cancelled (signal received) or all goroutines to complete naturally (less likely for log streaming)
	<-ctx.Done() // This will block until cancel() is called by the signal handler

	log.Println("Waiting for log streaming goroutines to finish...")
	wg.Wait() // Wait for all streamAndStoreLogs goroutines to complete

	log.Println("All goroutines finished. Closing resources.")
	if err := db.Close(); err != nil {
	    log.Printf("Error closing database: %v\n", err)
	}
	dli.Close()
	log.Println("Application shut down gracefully.")
}

// streamAndStoreLogs handles fetching, parsing, and storing logs for a single container.
func streamAndStoreLogs(ctx context.Context, dli *client.Client, containerID string, containerName string, wg *sync.WaitGroup) {
	defer wg.Done() // Decrement counter when goroutine finishes

	// Determine a unique "since" time for this container's log stream to minimize duplicate logs across restarts IF NEEDED.
	// For now, using current time as before, or a very early time to get more history for testing.
	// A more robust approach would be to query the DB for the last timestamp for this containerID.
	sinceTime := time.Now().Add(-5 * time.Minute).Format(time.RFC3339Nano) // Get last 5 mins for demo, or Now() for only new

	logOpts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
		Details:    false,
		Since:      sinceTime, 
	}

	logReader, err := dli.ContainerLogs(ctx, containerID, logOpts) // Pass cancellable context
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
		case 1: streamTypeStr = "stdout"
		case 2: streamTypeStr = "stderr"
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
		if payloadSize == 0 { continue }

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
		actualTimestampStr, actualLogMessage := parseLogEntry(logMessage, containerName)

		if actualLogMessage == "" { continue }

		// Retry logic for database insertion
		for i := 0; i < 3; i++ { // Try up to 3 times
		    select {
		    case <-ctx.Done():
		        log.Printf("Database insert for %s (%s) cancelled.\n", containerName, containerID[:12])
		        return
		    default:
		    }

			stmt, err := db.PrepareContext(ctx, "INSERT INTO logs(container_id, container_name, timestamp, stream_type, log_message) VALUES(?, ?, ?, ?, ?)")
			if err != nil {
				log.Printf("Error preparing SQL statement for %s (%s) (attempt %d): %v\n", containerName, containerID[:12], i+1, err)
				if i == 2 { break } // Last attempt failed
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond) // Exponential backoff
				continue
			}
			
			_, execErr := stmt.ExecContext(ctx, containerID, containerName, actualTimestampStr, streamTypeStr, actualLogMessage)
			stmt.Close() // Close statement after exec or on error

			if execErr == nil {
				// log.Printf("Logged for %s: [%s] %s", containerName, streamTypeStr, actualLogMessage) // Verbose
				break // Success
			}

			log.Printf("Error inserting log for %s (%s) into DB (attempt %d): %v. Log: %.50s...", containerName, containerID[:12], i+1, execErr, actualLogMessage)
			if i == 2 { break } // Last attempt failed
			
			// Check if the error is context cancellation
			if execErr.Error() == context.Canceled.Error() || execErr.Error() == context.DeadlineExceeded.Error() {
			    log.Printf("DB operation for %s cancelled or timed out.", containerName)
			    return
			}
			time.Sleep(time.Duration(i+1) * 200 * time.Millisecond) // Exponential backoff
		}
	}
	// log.Printf("Stopped streaming logs for container %s (%s)\n", containerName, containerID[:12]) // This will be logged when loop exits
}

// parseLogEntry extracts timestamp and message from a raw log line.
func parseLogEntry(logMessage string, containerName string) (timestampStr string, messageStr string) {
	logMessage = strings.TrimRight(logMessage, "\n\r") // Clean trailing newlines first

	firstSpaceIndex := strings.Index(logMessage, " ")
	if firstSpaceIndex > 0 {
		timestampCandidate := logMessage[:firstSpaceIndex]
		// Docker's own timestamp format with nanoseconds sometimes has 'Z' and sometimes doesn't directly.
		// time.RFC3339Nano expects 'Z' or a numeric offset.
		// We need to be a bit flexible or ensure Docker consistently provides 'Z'.
		// For now, assuming it's mostly compliant or we fall back.
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampCandidate)
		if err == nil {
			return parsedTime.UTC().Format(time.RFC3339Nano), strings.TrimSpace(logMessage[firstSpaceIndex+1:])
		}
		// Fallback for slightly different formats, e.g. missing 'Z' but otherwise ok (common in some loggers)
		parsedTime, err = time.Parse("2006-01-02T15:04:05.000000000", timestampCandidate) // Example without Z
		if err == nil {
		    return parsedTime.UTC().Format(time.RFC3339Nano), strings.TrimSpace(logMessage[firstSpaceIndex+1:])
		}
	}

	// Fallback: If no space or parsing failed, use current time in UTC and the whole message.
	// log.Printf("Debug: Could not parse timestamp from log line for %s. Line: '%s'. Using current time.", containerName, logMessage)
	return time.Now().UTC().Format(time.RFC3339Nano), strings.TrimSpace(logMessage)
} 