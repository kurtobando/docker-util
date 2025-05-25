package testutil

import (
	"context"
	"io"
	"testing"

	"bw_util/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestMockRepository(t *testing.T) {
	t.Run("should use custom Store function", func(t *testing.T) {
		called := false
		mockRepo := &MockRepository{
			StoreFunc: func(ctx context.Context, entry *models.LogEntry) error {
				called = true
				assert.Equal(t, "test-container", entry.ContainerName)
				return nil
			},
		}

		entry := &models.LogEntry{ContainerName: "test-container"}
		err := mockRepo.Store(context.Background(), entry)

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("should use custom Search function", func(t *testing.T) {
		expectedLogs := []models.LogEntry{
			{ContainerName: "test", LogMessage: "message"},
		}

		mockRepo := &MockRepository{
			SearchFunc: func(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
				assert.Equal(t, "test query", params.SearchQuery)
				return expectedLogs, true, nil
			},
		}

		params := &models.SearchParams{SearchQuery: "test query"}
		logs, hasNext, err := mockRepo.Search(context.Background(), params)

		assert.NoError(t, err)
		assert.True(t, hasNext)
		assert.Equal(t, expectedLogs, logs)
	})

	t.Run("should use custom GetContainers function", func(t *testing.T) {
		expectedContainers := []models.Container{
			{Name: "test1", ID: "id1"},
			{Name: "test2", ID: "id2"},
		}

		mockRepo := &MockRepository{
			GetContainersFunc: func(ctx context.Context) ([]models.Container, error) {
				return expectedContainers, nil
			},
		}

		containers, err := mockRepo.GetContainers(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, []string{"test1", "test2"}, containers)
	})

	t.Run("should return defaults when no functions set", func(t *testing.T) {
		mockRepo := &MockRepository{}

		// Test Store with defaults
		err := mockRepo.Store(context.Background(), &models.LogEntry{})
		assert.NoError(t, err)

		// Test Search with defaults
		logs, hasNext, err := mockRepo.Search(context.Background(), &models.SearchParams{})
		assert.NoError(t, err)
		assert.False(t, hasNext)
		assert.Empty(t, logs)

		// Test GetContainers with defaults
		containers, err := mockRepo.GetContainers(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, containers)
	})
}

func TestMockLogReader(t *testing.T) {
	t.Run("should read data correctly", func(t *testing.T) {
		testData := "line1\nline2\nline3"
		reader := NewMockLogReader(testData)

		// Read all data
		buffer := make([]byte, len(testData))
		n, err := reader.Read(buffer)

		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		assert.Equal(t, testData, string(buffer))
	})

	t.Run("should handle partial reads", func(t *testing.T) {
		testData := "hello world"
		reader := NewMockLogReader(testData)

		// Read in small chunks
		buffer := make([]byte, 5)
		n1, err1 := reader.Read(buffer)
		assert.NoError(t, err1)
		assert.Equal(t, 5, n1)
		assert.Equal(t, "hello", string(buffer))

		n2, err2 := reader.Read(buffer)
		assert.NoError(t, err2)
		assert.Equal(t, 5, n2)
		assert.Equal(t, " worl", string(buffer))

		// Read remaining
		buffer = make([]byte, 5)
		n3, err3 := reader.Read(buffer)
		assert.Equal(t, 1, n3)
		assert.Equal(t, "d", string(buffer[:n3]))
		if err3 != nil {
			assert.Equal(t, io.EOF, err3)
		}
	})

	t.Run("should return EOF when data exhausted", func(t *testing.T) {
		reader := NewMockLogReader("test")

		// Read all data
		buffer := make([]byte, 10)
		n, err := reader.Read(buffer)
		assert.Equal(t, 4, n)
		assert.NoError(t, err)

		// Try to read more
		n, err = reader.Read(buffer)
		assert.Equal(t, 0, n)
		assert.Equal(t, io.EOF, err)
	})

	t.Run("should handle close correctly", func(t *testing.T) {
		reader := NewMockLogReader("test data")

		// Close the reader
		err := reader.Close()
		assert.NoError(t, err)

		// Try to read after close
		buffer := make([]byte, 10)
		n, err := reader.Read(buffer)
		assert.Equal(t, 0, n)
		assert.Equal(t, io.EOF, err)
	})

	t.Run("should handle empty data", func(t *testing.T) {
		reader := NewMockLogReader("")

		buffer := make([]byte, 10)
		n, err := reader.Read(buffer)
		assert.Equal(t, 0, n)
		assert.Equal(t, io.EOF, err)
	})
}
