package testutil

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"bw_util/internal/database"
)

// CreateTempDB creates a temporary database for testing with auto-cleanup
func CreateTempDB(t *testing.T) *database.Database {
	t.Helper()

	// Create a temporary directory
	tempDir := t.TempDir()

	// Generate a random database name
	rand.Seed(time.Now().UnixNano())
	dbName := fmt.Sprintf("test_%d_%d.db", time.Now().Unix(), rand.Intn(10000))
	dbPath := filepath.Join(tempDir, dbName)

	// Create the database
	db := database.NewDatabase(dbPath)

	// Connect and initialize
	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("Warning: Failed to close temp database: %v", err)
		}
		// t.TempDir() automatically cleans up the directory
	})

	return db
}

// CreateTempDBPath creates a temporary database path without connecting
func CreateTempDBPath(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	rand.Seed(time.Now().UnixNano())
	dbName := fmt.Sprintf("test_%d_%d.db", time.Now().Unix(), rand.Intn(10000))
	return filepath.Join(tempDir, dbName)
}

// GetRandomPort returns a random available port for testing
func GetRandomPort() int {
	// Use port 0 to let the OS assign an available port
	// This will be handled by the server implementation
	return 0
}
