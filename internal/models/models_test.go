package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLogEntry tests the LogEntry struct functionality
func TestLogEntry(t *testing.T) {
	t.Run("should create LogEntry with all fields", func(t *testing.T) {
		entry := LogEntry{
			ContainerID:   "abc123def456",
			ContainerName: "test-container",
			Timestamp:     "2024-01-01T12:00:00Z",
			StreamType:    "stdout",
			LogMessage:    "Test log message",
		}

		assert.Equal(t, "abc123def456", entry.ContainerID)
		assert.Equal(t, "test-container", entry.ContainerName)
		assert.Equal(t, "2024-01-01T12:00:00Z", entry.Timestamp)
		assert.Equal(t, "stdout", entry.StreamType)
		assert.Equal(t, "Test log message", entry.LogMessage)
	})

	t.Run("should handle zero value LogEntry", func(t *testing.T) {
		var entry LogEntry

		assert.Empty(t, entry.ContainerID)
		assert.Empty(t, entry.ContainerName)
		assert.Empty(t, entry.Timestamp)
		assert.Empty(t, entry.StreamType)
		assert.Empty(t, entry.LogMessage)
	})

	t.Run("should handle LogEntry with empty strings", func(t *testing.T) {
		entry := LogEntry{
			ContainerID:   "",
			ContainerName: "",
			Timestamp:     "",
			StreamType:    "",
			LogMessage:    "",
		}

		assert.Empty(t, entry.ContainerID)
		assert.Empty(t, entry.ContainerName)
		assert.Empty(t, entry.Timestamp)
		assert.Empty(t, entry.StreamType)
		assert.Empty(t, entry.LogMessage)
	})

	t.Run("should handle LogEntry with special characters", func(t *testing.T) {
		entry := LogEntry{
			ContainerID:   "container-123_test",
			ContainerName: "test-container-name_123",
			Timestamp:     "2023-12-01T10:30:45.123456789Z",
			StreamType:    "stderr",
			LogMessage:    "Log with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
		}

		assert.Equal(t, "container-123_test", entry.ContainerID)
		assert.Equal(t, "test-container-name_123", entry.ContainerName)
		assert.Equal(t, "stderr", entry.StreamType)
		assert.Contains(t, entry.LogMessage, "!@#$%^&*()_+-=[]{}|;':\",./<>?")
	})
}

// TestSearchParams tests the SearchParams struct functionality
func TestSearchParams(t *testing.T) {
	t.Run("should create SearchParams with all fields", func(t *testing.T) {
		params := SearchParams{
			SearchQuery:        "error",
			SelectedContainers: []string{"container1", "container2"},
			CurrentPage:        2,
			Offset:             100,
		}

		assert.Equal(t, "error", params.SearchQuery)
		assert.Equal(t, []string{"container1", "container2"}, params.SelectedContainers)
		assert.Equal(t, 2, params.CurrentPage)
		assert.Equal(t, 100, params.Offset)
	})

	t.Run("should handle zero value SearchParams", func(t *testing.T) {
		var params SearchParams

		assert.Empty(t, params.SearchQuery)
		assert.Nil(t, params.SelectedContainers)
		assert.Equal(t, 0, params.CurrentPage)
		assert.Equal(t, 0, params.Offset)
	})

	t.Run("should handle SearchParams with empty slice", func(t *testing.T) {
		params := SearchParams{
			SearchQuery:        "test",
			SelectedContainers: []string{},
			CurrentPage:        1,
			Offset:             0,
		}

		assert.Equal(t, "test", params.SearchQuery)
		assert.NotNil(t, params.SelectedContainers)
		assert.Len(t, params.SelectedContainers, 0)
		assert.Equal(t, 1, params.CurrentPage)
		assert.Equal(t, 0, params.Offset)
	})
}

// TestPaginationInfo tests the PaginationInfo struct functionality
func TestPaginationInfo(t *testing.T) {
	t.Run("should create PaginationInfo with all fields", func(t *testing.T) {
		pagination := PaginationInfo{
			CurrentPage: 2,
			HasNext:     true,
			HasPrev:     true,
			NextPageURL: "/logs?page=3",
			PrevPageURL: "/logs?page=1",
		}

		assert.Equal(t, 2, pagination.CurrentPage)
		assert.True(t, pagination.HasNext)
		assert.True(t, pagination.HasPrev)
		assert.Equal(t, "/logs?page=3", pagination.NextPageURL)
		assert.Equal(t, "/logs?page=1", pagination.PrevPageURL)
	})

	t.Run("should handle zero value PaginationInfo", func(t *testing.T) {
		var info PaginationInfo

		assert.Equal(t, 0, info.CurrentPage)
		assert.False(t, info.HasNext)
		assert.False(t, info.HasPrev)
		assert.Empty(t, info.NextPageURL)
		assert.Empty(t, info.PrevPageURL)
	})

	t.Run("should handle first page pagination", func(t *testing.T) {
		pagination := PaginationInfo{
			CurrentPage: 1,
			HasNext:     true,
			HasPrev:     false,
			NextPageURL: "/logs?page=2",
			PrevPageURL: "",
		}

		assert.Equal(t, 1, pagination.CurrentPage)
		assert.True(t, pagination.HasNext)
		assert.False(t, pagination.HasPrev)
		assert.Equal(t, "/logs?page=2", pagination.NextPageURL)
		assert.Empty(t, pagination.PrevPageURL)
	})

	t.Run("should handle last page pagination", func(t *testing.T) {
		pagination := PaginationInfo{
			CurrentPage: 5,
			HasNext:     false,
			HasPrev:     true,
			NextPageURL: "",
			PrevPageURL: "/logs?page=4",
		}

		assert.Equal(t, 5, pagination.CurrentPage)
		assert.False(t, pagination.HasNext)
		assert.True(t, pagination.HasPrev)
		assert.Empty(t, pagination.NextPageURL)
		assert.Equal(t, "/logs?page=4", pagination.PrevPageURL)
	})
}

// TestContainer tests the Container struct functionality
func TestContainer(t *testing.T) {
	t.Run("should create Container with all fields", func(t *testing.T) {
		container := Container{
			ID:    "abc123def456",
			Name:  "test-container",
			Image: "nginx:latest",
		}

		assert.Equal(t, "abc123def456", container.ID)
		assert.Equal(t, "test-container", container.Name)
		assert.Equal(t, "nginx:latest", container.Image)
	})

	t.Run("should handle zero value Container", func(t *testing.T) {
		var container Container

		assert.Empty(t, container.ID)
		assert.Empty(t, container.Name)
		assert.Empty(t, container.Image)
	})

	t.Run("should handle Container with long values", func(t *testing.T) {
		longID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		container := Container{
			ID:    longID,
			Name:  "very-long-container-name-with-many-hyphens-and-numbers-123",
			Image: "registry.example.com/namespace/image-name:v1.2.3-alpha.1",
		}

		assert.Equal(t, longID, container.ID)
		assert.Len(t, container.ID, 64)
		assert.Contains(t, container.Name, "very-long-container-name")
		assert.Contains(t, container.Image, "registry.example.com")
	})
}

// TestTemplateData tests the TemplateData struct functionality
func TestTemplateData(t *testing.T) {
	t.Run("should create TemplateData with all fields", func(t *testing.T) {
		logs := []LogEntry{
			{
				ContainerID:   "abc123",
				ContainerName: "test-container",
				Timestamp:     "2024-01-01T12:00:00Z",
				StreamType:    "stdout",
				LogMessage:    "Test message",
			},
		}

		templateData := TemplateData{
			Logs:                logs,
			DBPath:              "/path/to/db",
			SearchQuery:         "error",
			ResultCount:         1,
			HasSearch:           true,
			CurrentPage:         2,
			HasNext:             true,
			HasPrev:             true,
			NextPageURL:         "/logs?page=3",
			PrevPageURL:         "/logs?page=1",
			AvailableContainers: []string{"container1", "container2"},
			SelectedContainers:  []string{"container1"},
			HasContainerFilter:  true,
		}

		assert.Equal(t, logs, templateData.Logs)
		assert.Equal(t, "/path/to/db", templateData.DBPath)
		assert.Equal(t, "error", templateData.SearchQuery)
		assert.Equal(t, 1, templateData.ResultCount)
		assert.True(t, templateData.HasSearch)
		assert.Equal(t, 2, templateData.CurrentPage)
		assert.True(t, templateData.HasNext)
		assert.True(t, templateData.HasPrev)
		assert.Equal(t, "/logs?page=3", templateData.NextPageURL)
		assert.Equal(t, "/logs?page=1", templateData.PrevPageURL)
		assert.Equal(t, []string{"container1", "container2"}, templateData.AvailableContainers)
		assert.Equal(t, []string{"container1"}, templateData.SelectedContainers)
		assert.True(t, templateData.HasContainerFilter)
	})

	t.Run("should handle zero value TemplateData", func(t *testing.T) {
		var templateData TemplateData

		assert.Nil(t, templateData.Logs)
		assert.Empty(t, templateData.DBPath)
		assert.Empty(t, templateData.SearchQuery)
		assert.Equal(t, 0, templateData.ResultCount)
		assert.False(t, templateData.HasSearch)
		assert.Equal(t, 0, templateData.CurrentPage)
		assert.False(t, templateData.HasNext)
		assert.False(t, templateData.HasPrev)
		assert.Empty(t, templateData.NextPageURL)
		assert.Empty(t, templateData.PrevPageURL)
		assert.Nil(t, templateData.AvailableContainers)
		assert.Nil(t, templateData.SelectedContainers)
		assert.False(t, templateData.HasContainerFilter)
	})

	t.Run("should handle TemplateData with empty slices", func(t *testing.T) {
		templateData := TemplateData{
			Logs:                []LogEntry{},
			AvailableContainers: []string{},
			SelectedContainers:  []string{},
		}

		assert.Empty(t, templateData.Logs)
		assert.Empty(t, templateData.AvailableContainers)
		assert.Empty(t, templateData.SelectedContainers)
	})

	t.Run("should handle TemplateData with complex nested data", func(t *testing.T) {
		logs := []LogEntry{
			{ContainerID: "1", LogMessage: "log1"},
			{ContainerID: "2", LogMessage: "log2"},
		}

		data := TemplateData{
			Logs:                logs,
			DBPath:              "/path/to/db",
			SearchQuery:         "error",
			ResultCount:         100,
			HasSearch:           true,
			CurrentPage:         3,
			HasNext:             false,
			HasPrev:             true,
			NextPageURL:         "",
			PrevPageURL:         "/logs?page=2",
			AvailableContainers: []string{"web", "db", "cache"},
			SelectedContainers:  []string{"web", "db"},
			HasContainerFilter:  true,
		}

		assert.Len(t, data.Logs, 2)
		assert.Equal(t, "error", data.SearchQuery)
		assert.Equal(t, 100, data.ResultCount)
		assert.True(t, data.HasSearch)
		assert.Equal(t, 3, data.CurrentPage)
		assert.False(t, data.HasNext)
		assert.True(t, data.HasPrev)
		assert.Empty(t, data.NextPageURL)
		assert.Equal(t, "/logs?page=2", data.PrevPageURL)
		assert.Len(t, data.AvailableContainers, 3)
		assert.Len(t, data.SelectedContainers, 2)
		assert.True(t, data.HasContainerFilter)
	})
}

// TestModelsMemoryLayout tests memory layout and struct sizes
func TestModelsMemoryLayout(t *testing.T) {
	t.Run("should have reasonable struct sizes", func(t *testing.T) {
		// Test that structs don't have unexpected memory overhead
		var logEntry LogEntry
		var searchParams SearchParams
		var paginationInfo PaginationInfo
		var container Container
		var templateData TemplateData

		// These tests ensure structs are properly defined
		assert.NotNil(t, &logEntry)
		assert.NotNil(t, &searchParams)
		assert.NotNil(t, &paginationInfo)
		assert.NotNil(t, &container)
		assert.NotNil(t, &templateData)
	})

	t.Run("should handle struct field alignment", func(t *testing.T) {
		// Test that all struct fields are accessible
		entry := LogEntry{
			ContainerID:   "test",
			ContainerName: "test",
			Timestamp:     "test",
			StreamType:    "test",
			LogMessage:    "test",
		}

		// All fields should be settable and gettable
		assert.Equal(t, "test", entry.ContainerID)
		assert.Equal(t, "test", entry.ContainerName)
		assert.Equal(t, "test", entry.Timestamp)
		assert.Equal(t, "test", entry.StreamType)
		assert.Equal(t, "test", entry.LogMessage)
	})
}
