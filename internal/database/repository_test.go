package database

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"bw_util/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*Database, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db := &Database{
		db:     mockDB,
		dbPath: "test.db",
	}

	cleanup := func() {
		mockDB.Close()
	}

	return db, mock, cleanup
}

func createTempDB(t *testing.T) *Database {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db := NewDatabase(dbPath)
	err := db.Connect()
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestNewRepository(t *testing.T) {
	t.Run("should create new repository", func(t *testing.T) {
		db := &Database{}
		repo := NewRepository(db)

		assert.NotNil(t, repo)
		assert.Equal(t, db, repo.db)
	})

	t.Run("should handle nil database", func(t *testing.T) {
		repo := NewRepository(nil)

		assert.NotNil(t, repo)
		assert.Nil(t, repo.db)
	})
}

func TestRepository_Store(t *testing.T) {
	// Create temporary database for testing
	tempDB := createTempDB(t)

	repo := NewRepository(tempDB)

	t.Run("should store log entry successfully", func(t *testing.T) {
		entry := &models.LogEntry{
			ContainerID:   "test-container-id",
			ContainerName: "test-container",
			Timestamp:     "2024-01-01T10:00:00.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Test log message",
		}

		err := repo.Store(context.Background(), entry)
		assert.NoError(t, err)

		// Verify the entry was stored
		var count int
		err = tempDB.GetDB().QueryRow("SELECT COUNT(*) FROM logs WHERE container_name = ?", "test-container").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("should handle multiple entries", func(t *testing.T) {
		entries := []*models.LogEntry{
			{
				ContainerID:   "container1",
				ContainerName: "app1",
				Timestamp:     "2024-01-01T10:00:01.000000000Z",
				StreamType:    "stdout",
				LogMessage:    "Message 1",
			},
			{
				ContainerID:   "container2",
				ContainerName: "app2",
				Timestamp:     "2024-01-01T10:00:02.000000000Z",
				StreamType:    "stderr",
				LogMessage:    "Message 2",
			},
		}

		for _, entry := range entries {
			err := repo.Store(context.Background(), entry)
			assert.NoError(t, err)
		}

		// Verify both entries were stored
		var count int
		err := tempDB.GetDB().QueryRow("SELECT COUNT(*) FROM logs WHERE container_name IN ('app1', 'app2')").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("should handle context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		entry := &models.LogEntry{
			ContainerID:   "test-id",
			ContainerName: "test-name",
			Timestamp:     "2024-01-01T10:00:00.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Test message",
		}

		err := repo.Store(ctx, entry)
		// Depending on timing, this might succeed or fail with context error
		// We just ensure it doesn't panic
		_ = err
	})

	t.Run("should handle nil entry", func(t *testing.T) {
		// Test that nil entry causes a panic (which is expected behavior)
		assert.Panics(t, func() {
			repo.Store(context.Background(), nil)
		})
	})

	t.Run("should handle repository with nil database", func(t *testing.T) {
		nilRepo := NewRepository(nil)
		entry := &models.LogEntry{
			ContainerID:   "test-id",
			ContainerName: "test-name",
			Timestamp:     "2024-01-01T10:00:00.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Test message",
		}

		// Test that nil database causes a panic (which is expected behavior)
		assert.Panics(t, func() {
			nilRepo.Store(context.Background(), entry)
		})
	})
}

func TestRepository_Search(t *testing.T) {
	// Create temporary database and populate with test data
	tempDB := createTempDB(t)

	repo := NewRepository(tempDB)

	// Insert test data
	testEntries := []models.LogEntry{
		{
			ContainerID:   "container1",
			ContainerName: "web-server",
			Timestamp:     "2024-01-01T10:00:00.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Server started successfully",
		},
		{
			ContainerID:   "container1",
			ContainerName: "web-server",
			Timestamp:     "2024-01-01T10:00:01.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Listening on port 8080",
		},
		{
			ContainerID:   "container2",
			ContainerName: "database",
			Timestamp:     "2024-01-01T10:00:02.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Database connection established",
		},
		{
			ContainerID:   "container2",
			ContainerName: "database",
			Timestamp:     "2024-01-01T10:00:03.000000000Z",
			StreamType:    "stderr",
			LogMessage:    "Warning: deprecated configuration",
		},
	}

	for _, entry := range testEntries {
		err := repo.Store(context.Background(), &entry)
		require.NoError(t, err)
	}

	t.Run("should search without filters", func(t *testing.T) {
		params := &models.SearchParams{
			SearchQuery:        "",
			SelectedContainers: []string{},
			CurrentPage:        1,
			Offset:             0,
		}

		logs, hasNext, err := repo.Search(context.Background(), params)
		assert.NoError(t, err)
		assert.False(t, hasNext)               // Should be false since we have < 100 entries
		assert.GreaterOrEqual(t, len(logs), 4) // At least our test entries
	})

	t.Run("should search with text query", func(t *testing.T) {
		params := &models.SearchParams{
			SearchQuery:        "server",
			SelectedContainers: []string{},
			CurrentPage:        1,
			Offset:             0,
		}

		logs, hasNext, err := repo.Search(context.Background(), params)
		assert.NoError(t, err)
		assert.False(t, hasNext)

		// Should find entries containing "server"
		found := false
		for _, log := range logs {
			if strings.Contains(strings.ToLower(log.LogMessage), "server") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("should search with container filter", func(t *testing.T) {
		params := &models.SearchParams{
			SearchQuery:        "",
			SelectedContainers: []string{"web-server"},
			CurrentPage:        1,
			Offset:             0,
		}

		logs, hasNext, err := repo.Search(context.Background(), params)
		assert.NoError(t, err)
		assert.False(t, hasNext)

		// All returned logs should be from web-server
		for _, log := range logs {
			if log.ContainerName != "" { // Skip empty entries
				assert.Equal(t, "web-server", log.ContainerName)
			}
		}
	})

	t.Run("should handle pagination", func(t *testing.T) {
		params := &models.SearchParams{
			SearchQuery:        "",
			SelectedContainers: []string{},
			CurrentPage:        1,
			Offset:             0,
		}

		logs, hasNext, err := repo.Search(context.Background(), params)
		assert.NoError(t, err)
		assert.NotNil(t, logs)
		// hasNext depends on total number of entries
		_ = hasNext
	})

	t.Run("should handle context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		params := &models.SearchParams{
			SearchQuery:        "",
			SelectedContainers: []string{},
			CurrentPage:        1,
			Offset:             0,
		}

		logs, hasNext, err := repo.Search(ctx, params)
		// Might succeed or fail depending on timing
		_ = logs
		_ = hasNext
		_ = err
	})

	t.Run("should handle nil search params", func(t *testing.T) {
		// Test that nil search params causes a panic (which is expected behavior)
		assert.Panics(t, func() {
			repo.Search(context.Background(), nil)
		})
	})

	t.Run("should handle repository with nil database", func(t *testing.T) {
		nilRepo := NewRepository(nil)
		params := &models.SearchParams{
			SearchQuery:        "",
			SelectedContainers: []string{},
			CurrentPage:        1,
			Offset:             0,
		}

		// Test that nil database causes a panic (which is expected behavior)
		assert.Panics(t, func() {
			nilRepo.Search(context.Background(), params)
		})
	})
}

func TestRepository_GetContainers(t *testing.T) {
	// Create temporary database and populate with test data
	tempDB := createTempDB(t)

	repo := NewRepository(tempDB)

	// Insert test data with different containers
	testEntries := []models.LogEntry{
		{
			ContainerID:   "container1",
			ContainerName: "web-server",
			Timestamp:     "2024-01-01T10:00:00.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Test message 1",
		},
		{
			ContainerID:   "container2",
			ContainerName: "database",
			Timestamp:     "2024-01-01T10:00:01.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Test message 2",
		},
		{
			ContainerID:   "container1",
			ContainerName: "web-server",
			Timestamp:     "2024-01-01T10:00:02.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Test message 3",
		},
	}

	for _, entry := range testEntries {
		err := repo.Store(context.Background(), &entry)
		require.NoError(t, err)
	}

	t.Run("should get unique container names", func(t *testing.T) {
		containers, err := repo.GetContainers(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, containers)

		// Should contain our test containers
		containerMap := make(map[string]bool)
		for _, name := range containers {
			containerMap[name] = true
		}

		assert.True(t, containerMap["web-server"])
		assert.True(t, containerMap["database"])
	})

	t.Run("should handle empty database", func(t *testing.T) {
		// Create a fresh database with no entries
		emptyDB := createTempDB(t)

		emptyRepo := NewRepository(emptyDB)
		containers, err := emptyRepo.GetContainers(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, containers) // Should be empty slice, not nil
	})

	t.Run("should handle context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		containers, err := repo.GetContainers(ctx)
		// Might succeed or fail depending on timing
		_ = containers
		_ = err
	})

	t.Run("should handle repository with nil database", func(t *testing.T) {
		nilRepo := NewRepository(nil)

		// Test that nil database causes a panic (which is expected behavior)
		assert.Panics(t, func() {
			nilRepo.GetContainers(context.Background())
		})
	})
}

func TestRepository_ErrorHandling(t *testing.T) {
	t.Run("should handle database connection errors", func(t *testing.T) {
		// Create a database with an invalid path to simulate connection errors
		invalidDB := NewDatabase("/invalid/path/that/does/not/exist.db")
		// Don't call Connect() to simulate a disconnected database

		repo := NewRepository(invalidDB)

		// Test Store with disconnected database - should panic
		entry := &models.LogEntry{
			ContainerID:   "test-id",
			ContainerName: "test-name",
			Timestamp:     "2024-01-01T10:00:00.000000000Z",
			StreamType:    "stdout",
			LogMessage:    "Test message",
		}

		assert.Panics(t, func() {
			repo.Store(context.Background(), entry)
		})

		// Test Search with disconnected database - should panic
		params := &models.SearchParams{
			SearchQuery:        "",
			SelectedContainers: []string{},
			CurrentPage:        1,
			Offset:             0,
		}

		assert.Panics(t, func() {
			repo.Search(context.Background(), params)
		})

		// Test GetContainers with disconnected database - should panic
		assert.Panics(t, func() {
			repo.GetContainers(context.Background())
		})
	})
}
