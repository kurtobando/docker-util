package main

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

const dbName = "./docker_logs.db"

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", dbName)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
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
		return fmt.Errorf("failed to create logs table: %w", err)
	}
	log.Println("Database initialized and logs table ensured.")
	return nil
}

func main() {
	ctx := context.Background()

	if err := initDB(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	dli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}
	defer dli.Close()

	containers, err := dli.ContainerList(ctx, container.ListOptions{All: false}) // Only running containers
	if err != nil {
		log.Fatalf("Error listing containers: %v", err)
	}

	if len(containers) == 0 {
		log.Println("No running containers found to monitor.")
		return
	}

	log.Printf("Found %d running containers. Starting log streaming...\n", len(containers))

	// Use a WaitGroup to wait for all goroutines to finish if the application were to have a clean shutdown.
	// var wg sync.WaitGroup

	for _, cont := range containers {
		containerName := "unknown"
		if len(cont.Names) > 0 {
			containerName = strings.TrimPrefix(cont.Names[0], "/")
		}
		log.Printf("Starting log stream for: ID=%s, Name=%s, Image=%s\n", cont.ID[:12], containerName, cont.Image)

		// wg.Add(1)
		go streamAndStoreLogs(ctx, dli, cont.ID, containerName /*, &wg */) // Launch a goroutine for each container
	}

	// wg.Wait() // Wait for all goroutines to complete if implementing a clean shutdown
	log.Println("Log collection setup complete. Application will run until manually stopped.")
	select {} // Keep main running
}

// streamAndStoreLogs handles fetching, parsing, and storing logs for a single container.
func streamAndStoreLogs(ctx context.Context, dli *client.Client, containerID string, containerName string /*, wg *sync.WaitGroup */) {
	// defer wg.Done() // Decrement counter when goroutine finishes if using WaitGroup

	logOpts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,     // Continuously stream logs
		Timestamps: true,     // Get timestamps from Docker daemon
		Details:    false,    // Not needed for basic log message, timestamp, stream type
		Since:      time.Now().Format(time.RFC3339Nano), // Get logs since now, effectively only new logs
	}

	logReader, err := dli.ContainerLogs(ctx, containerID, logOpts)
	if err != nil {
		log.Printf("Error getting logs for container %s (%s): %v\n", containerName, containerID[:12], err)
		return
	}
	defer logReader.Close()

	// Docker log streams are multiplexed for stdout and stderr.
	// Each log entry is prefixed with an 8-byte header:
	//   - 1 byte: stream type (0 for stdin, 1 for stdout, 2 for stderr)
	//   - 3 bytes: unused (padding)
	//   - 4 bytes: length of the log message (BigEndian uint32)
	// We need to parse this header to correctly read the log message.

	hdr := make([]byte, 8)
	for {
		// Read the 8-byte header
		_, err := logReader.Read(hdr)
		if err != nil {
			if err == io.EOF {
				log.Printf("Log stream ended for container %s (%s)\n", containerName, containerID[:12])
				break
			}
			log.Printf("Error reading log stream header for %s (%s): %v\n", containerName, containerID[:12], err)
			break // or return, depending on desired retry logic
		}

		var streamTypeStr string
		switch hdr[0] {
		case 1: // stdout
			streamTypeStr = "stdout"
		case 2: // stderr
			streamTypeStr = "stderr"
		default: // stdin or other, skip
			log.Printf("Skipping unknown stream type %d for %s\n", hdr[0], containerName)
			// We still need to read the payload to advance the stream correctly.
			payloadSize := binary.BigEndian.Uint32(hdr[4:])
			if payloadSize > 0 {
				content := make([]byte, payloadSize)
				_, err = io.ReadFull(logReader, content)
				if err != nil {
					log.Printf("Error reading and discarding unknown stream payload for %s: %v\n", containerName, err)
					break
				}
			}
			continue
		}

		// Extract the size of the log message payload
		payloadSize := binary.BigEndian.Uint32(hdr[4:])
		if payloadSize == 0 {
			continue // No actual message
		}

		// Read the log message payload
		content := make([]byte, payloadSize)
		_, err = io.ReadFull(logReader, content)
		if err != nil {
			log.Printf("Error reading log message payload for %s (%s): %v\n", containerName, containerID[:12], err)
			break
		}

		logMessage := string(content)

		// The Docker API provides timestamps if logOpts.Timestamps is true.
		// These timestamps are part of the log message itself, usually at the beginning.
		// Example: "2024-01-27T15:04:05.123456789Z log content here"
		// We need to parse this out if we want to store it separately and more cleanly.
		// For simplicity in this first pass, we'll assume the timestamp is at the start of the string up to the first space.
		// A more robust solution would use regex or a proper log parsing library if formats vary.

		var actualTimestampStr string
		var actualLogMessage string

		firstSpaceIndex := strings.Index(logMessage, " ")
		if firstSpaceIndex > 0 {
			timestampCandidate := logMessage[:firstSpaceIndex]
			// Attempt to parse it to validate it's a timestamp Docker uses (RFC3339Nano)
			_, err := time.Parse(time.RFC3339Nano, strings.TrimSuffix(timestampCandidate, "\n")) // Trim trailing newline if present before parsing
			if err == nil {
				actualTimestampStr = timestampCandidate
				actualLogMessage = strings.TrimSpace(logMessage[firstSpaceIndex+1:])
			} else {
				// Not a recognized timestamp format at the beginning, use current time and full message
				actualTimestampStr = time.Now().Format(time.RFC3339Nano)
				actualLogMessage = strings.TrimSpace(logMessage)
				// log.Printf("Debug: Could not parse timestamp '%s' for %s. Error: %v. Using current time.", timestampCandidate, containerName, err)
			}
		} else {
			actualTimestampStr = time.Now().Format(time.RFC3339Nano)
			actualLogMessage = strings.TrimSpace(logMessage)
			// log.Printf("Debug: No space found in log message for %s. Using current time. Msg: '%s'", containerName, logMessage)
		}

		if actualLogMessage == "" { // Don't store empty log messages
		    continue
		}

		// Store in database
		stmt, err := db.Prepare("INSERT INTO logs(container_id, container_name, timestamp, stream_type, log_message) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			log.Printf("Error preparing SQL statement for %s (%s): %v\n", containerName, containerID[:12], err)
			continue // Or break, depending on how critical this is
		}
		defer stmt.Close() // Close statement in each iteration where it's prepared

		_, err = stmt.Exec(containerID, containerName, actualTimestampStr, streamTypeStr, actualLogMessage)
		if err != nil {
			log.Printf("Error inserting log for %s (%s) into DB: %v. Log: %s", containerName, containerID[:12], err, actualLogMessage)
			// Consider retry or other error handling
		}
		// log.Printf("Logged for %s: [%s] %s", containerName, streamTypeStr, actualLogMessage) // Verbose logging
	}
	log.Printf("Stopped streaming logs for container %s (%s)\n", containerName, containerID[:12])
} 