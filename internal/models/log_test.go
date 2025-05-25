package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	t.Run("should handle empty LogEntry", func(t *testing.T) {
		entry := LogEntry{}

		assert.Empty(t, entry.ContainerID)
		assert.Empty(t, entry.ContainerName)
		assert.Empty(t, entry.Timestamp)
		assert.Empty(t, entry.StreamType)
		assert.Empty(t, entry.LogMessage)
	})
}

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

	t.Run("should handle empty SearchParams", func(t *testing.T) {
		params := SearchParams{}

		assert.Empty(t, params.SearchQuery)
		assert.Nil(t, params.SelectedContainers)
		assert.Equal(t, 0, params.CurrentPage)
		assert.Equal(t, 0, params.Offset)
	})

	t.Run("should handle SearchParams with empty containers slice", func(t *testing.T) {
		params := SearchParams{
			SearchQuery:        "test",
			SelectedContainers: []string{},
			CurrentPage:        1,
			Offset:             0,
		}

		assert.Equal(t, "test", params.SearchQuery)
		assert.Empty(t, params.SelectedContainers)
		assert.Equal(t, 1, params.CurrentPage)
		assert.Equal(t, 0, params.Offset)
	})
}

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

	t.Run("should handle empty Container", func(t *testing.T) {
		container := Container{}

		assert.Empty(t, container.ID)
		assert.Empty(t, container.Name)
		assert.Empty(t, container.Image)
	})
}

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

	t.Run("should handle empty TemplateData", func(t *testing.T) {
		templateData := TemplateData{}

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
}
