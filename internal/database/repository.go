package database

import (
	"context"
	"fmt"
	"log"
	"strings"

	"bw_util/internal/models"
)

// LogRepository defines the interface for log database operations
type LogRepository interface {
	Store(ctx context.Context, entry *models.LogEntry) error
	Search(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error)
	GetContainers(ctx context.Context) ([]string, error)
}

// Repository implements LogRepository
type Repository struct {
	db *Database
}

// NewRepository creates a new repository instance
func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

// Store inserts a log entry into the database
// This preserves the exact same retry logic from the original streamAndStoreLogs function
func (r *Repository) Store(ctx context.Context, entry *models.LogEntry) error {
	// Retry logic for database insertion (from original code)
	for i := 0; i < 3; i++ { // Try up to 3 times
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		stmt, err := r.db.GetDB().PrepareContext(ctx, "INSERT INTO logs(container_id, container_name, timestamp, stream_type, log_message) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			log.Printf("Error preparing SQL statement (attempt %d): %v\n", i+1, err)
			if i == 2 {
				return fmt.Errorf("failed to prepare statement after 3 attempts: %w", err)
			}
			continue
		}

		_, execErr := stmt.ExecContext(ctx, entry.ContainerID, entry.ContainerName, entry.Timestamp, entry.StreamType, entry.LogMessage)
		stmt.Close() // Close statement after exec or on error

		if execErr == nil {
			return nil // Success
		}

		log.Printf("Error inserting log into DB (attempt %d): %v. Log: %.50s...", i+1, execErr, entry.LogMessage)
		if i == 2 {
			return fmt.Errorf("failed to insert log after 3 attempts: %w", execErr)
		}

		// Check if the error is context cancellation
		if execErr.Error() == context.Canceled.Error() || execErr.Error() == context.DeadlineExceeded.Error() {
			return execErr
		}
	}
	return nil
}

// Search retrieves logs based on search parameters
// This preserves the exact same query logic from the original serveLogsPage function
func (r *Repository) Search(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
	// Build SQL query conditionally based on search and container filters
	// Use LIMIT 101 to check if there's a next page (exact same logic as original)
	var query string
	var args []interface{}
	var whereConditions []string

	// Add container filter if specified
	if len(params.SelectedContainers) > 0 {
		placeholders := make([]string, len(params.SelectedContainers))
		for i, container := range params.SelectedContainers {
			placeholders[i] = "?"
			args = append(args, container)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("container_name IN (%s)", strings.Join(placeholders, ", ")))
	}

	// Add search filter if specified
	if params.SearchQuery != "" {
		whereConditions = append(whereConditions, "log_message LIKE ? COLLATE NOCASE")
		args = append(args, "%"+params.SearchQuery+"%")
	}

	// Build the final query
	query = "SELECT container_id, container_name, timestamp, stream_type, log_message FROM logs"
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}
	query += " ORDER BY id DESC LIMIT 101 OFFSET ?"
	args = append(args, params.Offset)

	rows, err := r.db.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var entry models.LogEntry
		if err := rows.Scan(&entry.ContainerID, &entry.ContainerName, &entry.Timestamp, &entry.StreamType, &entry.LogMessage); err != nil {
			log.Printf("Error scanning log row: %v", err)
			// Skip this row and continue (same as original)
			continue
		}
		// Keep full container ID for display
		logs = append(logs, entry)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating log rows: %v", err)
		// Still try to return what we have (same as original)
	}

	// Determine if there's a next page (same logic as original)
	hasNext := len(logs) > 100
	if hasNext {
		logs = logs[:100] // Remove the extra one for display
	}

	return logs, hasNext, nil
}

// GetContainers fetches distinct container names from the database
// This preserves the exact same logic from the original getAvailableContainers function
func (r *Repository) GetContainers(ctx context.Context) ([]string, error) {
	query := "SELECT DISTINCT container_name FROM logs ORDER BY container_name"
	rows, err := r.db.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query container names: %w", err)
	}
	defer rows.Close()

	var containers []string
	for rows.Next() {
		var containerName string
		if err := rows.Scan(&containerName); err != nil {
			log.Printf("Error scanning container name: %v", err)
			continue
		}
		containers = append(containers, containerName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating container names: %w", err)
	}

	return containers, nil
}
