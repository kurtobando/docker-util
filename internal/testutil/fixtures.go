package testutil

import (
	"time"

	"bw_util/internal/models"
)

// SampleLogEntries provides consistent test data for log entries
var SampleLogEntries = []models.LogEntry{
	{
		ContainerID:   "abc123def456789",
		ContainerName: "web-server",
		Timestamp:     "2024-01-01T10:00:00.000000000Z",
		StreamType:    "stdout",
		LogMessage:    "Server started on port 8080",
	},
	{
		ContainerID:   "abc123def456789",
		ContainerName: "web-server",
		Timestamp:     "2024-01-01T10:00:01.123456789Z",
		StreamType:    "stdout",
		LogMessage:    "Listening for connections...",
	},
	{
		ContainerID:   "def456ghi789012",
		ContainerName: "database",
		Timestamp:     "2024-01-01T10:00:02.456789012Z",
		StreamType:    "stdout",
		LogMessage:    "Database connection established",
	},
	{
		ContainerID:   "def456ghi789012",
		ContainerName: "database",
		Timestamp:     "2024-01-01T10:00:03.789012345Z",
		StreamType:    "stderr",
		LogMessage:    "Warning: deprecated configuration option",
	},
	{
		ContainerID:   "ghi789jkl012345",
		ContainerName: "cache-redis",
		Timestamp:     "2024-01-01T10:00:04.012345678Z",
		StreamType:    "stdout",
		LogMessage:    "Redis server started",
	},
}

// SampleContainers provides consistent test data for containers
var SampleContainers = []models.Container{
	{
		ID:    "abc123def456789",
		Name:  "web-server",
		Image: "nginx:latest",
	},
	{
		ID:    "def456ghi789012",
		Name:  "database",
		Image: "postgres:13",
	},
	{
		ID:    "ghi789jkl012345",
		Name:  "cache-redis",
		Image: "redis:alpine",
	},
}

// CreateSampleLogEntry creates a log entry with customizable fields
func CreateSampleLogEntry(containerName, message string) *models.LogEntry {
	return &models.LogEntry{
		ContainerID:   "test123456789",
		ContainerName: containerName,
		Timestamp:     time.Now().Format(time.RFC3339Nano),
		StreamType:    "stdout",
		LogMessage:    message,
	}
}

// CreateSampleContainer creates a container with customizable fields
func CreateSampleContainer(id, name, image string) models.Container {
	return models.Container{
		ID:    id,
		Name:  name,
		Image: image,
	}
}

// SampleSearchParams provides test data for search parameters
var SampleSearchParams = &models.SearchParams{
	SearchQuery:        "server",
	SelectedContainers: []string{"web-server", "database"},
	CurrentPage:        1,
	Offset:             0,
}

// SampleTemplateData provides test data for template rendering
var SampleTemplateData = &models.TemplateData{
	Logs:                SampleLogEntries,
	DBPath:              "/tmp/test.db",
	SearchQuery:         "test query",
	ResultCount:         5,
	HasSearch:           true,
	CurrentPage:         1,
	HasNext:             true,
	HasPrev:             false,
	NextPageURL:         "/?page=2&search=test",
	PrevPageURL:         "",
	AvailableContainers: []string{"web-server", "database", "cache-redis"},
	SelectedContainers:  []string{"web-server"},
	HasContainerFilter:  true,
}
