package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/binary"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
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

//go:embed templates
var templatesFS embed.FS

var (
	dbName     string
	serverPort string
	db         *sql.DB
	tmpl       *template.Template
)

// LogEntry matches the structure for displaying logs in the template
type LogEntry struct {
	ContainerID   string
	ContainerName string
	Timestamp     string
	StreamType    string
	LogMessage    string
}

func init() {
	flag.StringVar(&dbName, "dbpath", "./docker_logs.db", "Path to the SQLite database file.")
	flag.StringVar(&serverPort, "port", "8080", "Port for the web UI.")
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", dbName+"?_busy_timeout=5000") // Added busy_timeout
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
		db.Close()
		return fmt.Errorf("failed to create logs table in %s: %w", dbName, err)
	}
	log.Printf("Database initialized at %s and logs table ensured.", dbName)
	return nil
}

func loadTemplates() error {
	var err error
	tmpl, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return fmt.Errorf("error parsing templates: %w", err)
	}
	log.Println("HTML templates loaded.")
	return nil
}

func serveLogsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.QueryContext(r.Context(), "SELECT container_id, container_name, timestamp, stream_type, log_message FROM logs ORDER BY id DESC LIMIT 100")
	if err != nil {
		log.Printf("Error querying logs from DB: %v", err)
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logsToDisplay []LogEntry
	for rows.Next() {
		var entry LogEntry
		if err := rows.Scan(&entry.ContainerID, &entry.ContainerName, &entry.Timestamp, &entry.StreamType, &entry.LogMessage); err != nil {
			log.Printf("Error scanning log row: %v", err)
			// Skip this row and continue
			continue
		}
		// Truncate long container IDs for display
		if len(entry.ContainerID) > 12 {
			entry.ContainerID = entry.ContainerID[:12]
		}
		logsToDisplay = append(logsToDisplay, entry)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating log rows: %v", err)
		// Still try to display what we have, or error out
		// For now, proceed with what was scanned successfully
	}

	// Use a map for data to easily add page title or other elements if needed
	data := map[string]interface{}{
		"Logs":   logsToDisplay,
		"DBPath": dbName, // Example of passing more data
	}

	err = tmpl.ExecuteTemplate(w, "logs.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, initiating graceful shutdown...")
		cancel()
	}()

	if err := initDB(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	if err := loadTemplates(); err != nil {
		log.Fatalf("Error loading HTML templates: %v", err)
	}

	// Start Docker log collection
	dli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error creating Docker client (log collection might not work): %v", err)
		// Decide if this is fatal. For now, web UI might still work with existing DB data.
	} else {
		containers, err := dli.ContainerList(ctx, container.ListOptions{All: false})
		if err != nil {
			if ctx.Err() == context.Canceled {
				log.Println("Context cancelled during container listing.")
			} else {
				log.Printf("Error listing containers (log collection might be impacted): %v", err)
			}
		} else if len(containers) == 0 {
			log.Println("No running containers found to monitor.")
		} else {
			log.Printf("Found %d running containers. Starting log streaming...\n", len(containers))
			for _, cont := range containers {
				containerName := "unknown"
				if len(cont.Names) > 0 {
					containerName = strings.TrimPrefix(cont.Names[0], "/")
				}
				log.Printf("Starting log stream for: ID=%s, Name=%s, Image=%s\n", cont.ID[:12], containerName, cont.Image)
				wg.Add(1)
				go streamAndStoreLogs(ctx, dli, cont.ID, containerName, &wg)
			}
		}
	}

	// Setup HTTP server
	http.HandleFunc("/", serveLogsPage)
	server := &http.Server{Addr: ":" + serverPort}

	log.Printf("Starting web UI on http://localhost:%s", serverPort)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down web server...")
	shutdownCtx, serverShutdownCancel := context.WithTimeout(context.Background(), 5*time.Second) // Give it 5 seconds to shutdown gracefully
	defer serverShutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown error: %v", err)
	}

	log.Println("Waiting for log streaming goroutines to finish...")
	wg.Wait()

	log.Println("All goroutines finished. Closing resources.")
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v\n", err)
		}
	}
	if dli != nil {
		dli.Close()
	}
	log.Println("Application shut down gracefully.")
}

// streamAndStoreLogs handles fetching, parsing, and storing logs for a single container.
func streamAndStoreLogs(ctx context.Context, dli *client.Client, containerID string, containerName string, wg *sync.WaitGroup) {
	defer wg.Done() // Decrement counter when goroutine finishes

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
		actualTimestampStr, actualLogMessage := parseLogEntry(logMessage, containerName)

		if actualLogMessage == "" {
			continue
		}

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
				if i == 2 {
					break
				} // Last attempt failed
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
			if i == 2 {
				break
			} // Last attempt failed

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
