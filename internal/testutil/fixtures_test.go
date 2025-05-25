package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSampleLogEntries(t *testing.T) {
	t.Run("should provide valid sample log entries", func(t *testing.T) {
		assert.NotEmpty(t, SampleLogEntries)
		assert.Len(t, SampleLogEntries, 5)

		// Verify each entry has required fields
		for i, entry := range SampleLogEntries {
			assert.NotEmpty(t, entry.ContainerID, "Entry %d should have ContainerID", i)
			assert.NotEmpty(t, entry.ContainerName, "Entry %d should have ContainerName", i)
			assert.NotEmpty(t, entry.Timestamp, "Entry %d should have Timestamp", i)
			assert.NotEmpty(t, entry.StreamType, "Entry %d should have StreamType", i)
			assert.NotEmpty(t, entry.LogMessage, "Entry %d should have LogMessage", i)

			// Verify timestamp format
			_, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
			assert.NoError(t, err, "Entry %d should have valid timestamp format", i)

			// Verify stream type is valid
			assert.Contains(t, []string{"stdout", "stderr"}, entry.StreamType, "Entry %d should have valid stream type", i)
		}
	})

	t.Run("should have diverse container names", func(t *testing.T) {
		containerNames := make(map[string]bool)
		for _, entry := range SampleLogEntries {
			containerNames[entry.ContainerName] = true
		}

		// Should have multiple different container names
		assert.GreaterOrEqual(t, len(containerNames), 3)
		assert.Contains(t, containerNames, "web-server")
		assert.Contains(t, containerNames, "database")
		assert.Contains(t, containerNames, "cache-redis")
	})

	t.Run("should have both stdout and stderr entries", func(t *testing.T) {
		hasStdout := false
		hasStderr := false

		for _, entry := range SampleLogEntries {
			if entry.StreamType == "stdout" {
				hasStdout = true
			}
			if entry.StreamType == "stderr" {
				hasStderr = true
			}
		}

		assert.True(t, hasStdout, "Should have stdout entries")
		assert.True(t, hasStderr, "Should have stderr entries")
	})
}

func TestSampleContainers(t *testing.T) {
	t.Run("should provide valid sample containers", func(t *testing.T) {
		assert.NotEmpty(t, SampleContainers)
		assert.Len(t, SampleContainers, 3)

		// Verify each container has required fields
		for i, container := range SampleContainers {
			assert.NotEmpty(t, container.ID, "Container %d should have ID", i)
			assert.NotEmpty(t, container.Name, "Container %d should have Name", i)
			assert.NotEmpty(t, container.Image, "Container %d should have Image", i)
		}
	})

	t.Run("should have unique container IDs and names", func(t *testing.T) {
		ids := make(map[string]bool)
		names := make(map[string]bool)

		for _, container := range SampleContainers {
			assert.False(t, ids[container.ID], "Container ID should be unique: %s", container.ID)
			assert.False(t, names[container.Name], "Container name should be unique: %s", container.Name)
			ids[container.ID] = true
			names[container.Name] = true
		}
	})

	t.Run("should match log entry container names", func(t *testing.T) {
		containerNames := make(map[string]bool)
		for _, container := range SampleContainers {
			containerNames[container.Name] = true
		}

		logContainerNames := make(map[string]bool)
		for _, entry := range SampleLogEntries {
			logContainerNames[entry.ContainerName] = true
		}

		// All log container names should exist in sample containers
		for name := range logContainerNames {
			assert.True(t, containerNames[name], "Log container name %s should exist in sample containers", name)
		}
	})
}

func TestCreateSampleLogEntry(t *testing.T) {
	t.Run("should create valid log entry", func(t *testing.T) {
		entry := CreateSampleLogEntry("test-container", "test message")

		assert.NotNil(t, entry)
		assert.Equal(t, "test-container", entry.ContainerName)
		assert.Equal(t, "test message", entry.LogMessage)
		assert.Equal(t, "test123456789", entry.ContainerID)
		assert.Equal(t, "stdout", entry.StreamType)
		assert.NotEmpty(t, entry.Timestamp)

		// Verify timestamp is valid and recent
		timestamp, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
		assert.NoError(t, err)
		assert.WithinDuration(t, time.Now(), timestamp, time.Second)
	})

	t.Run("should create entries with different timestamps", func(t *testing.T) {
		entry1 := CreateSampleLogEntry("container1", "message1")
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		entry2 := CreateSampleLogEntry("container2", "message2")

		assert.NotEqual(t, entry1.Timestamp, entry2.Timestamp)
		assert.Equal(t, "container1", entry1.ContainerName)
		assert.Equal(t, "container2", entry2.ContainerName)
	})
}

func TestCreateSampleContainer(t *testing.T) {
	t.Run("should create valid container", func(t *testing.T) {
		container := CreateSampleContainer("test-id", "test-name", "test-image")

		assert.Equal(t, "test-id", container.ID)
		assert.Equal(t, "test-name", container.Name)
		assert.Equal(t, "test-image", container.Image)
	})

	t.Run("should handle empty values", func(t *testing.T) {
		container := CreateSampleContainer("", "", "")

		assert.Equal(t, "", container.ID)
		assert.Equal(t, "", container.Name)
		assert.Equal(t, "", container.Image)
	})
}

func TestSampleSearchParams(t *testing.T) {
	t.Run("should provide valid search parameters", func(t *testing.T) {
		assert.NotNil(t, SampleSearchParams)
		assert.Equal(t, "server", SampleSearchParams.SearchQuery)
		assert.Equal(t, []string{"web-server", "database"}, SampleSearchParams.SelectedContainers)
		assert.Equal(t, 1, SampleSearchParams.CurrentPage)
		assert.Equal(t, 0, SampleSearchParams.Offset)
	})
}

func TestSampleTemplateData(t *testing.T) {
	t.Run("should provide valid template data", func(t *testing.T) {
		assert.NotNil(t, SampleTemplateData)
		assert.Equal(t, SampleLogEntries, SampleTemplateData.Logs)
		assert.Equal(t, "/tmp/test.db", SampleTemplateData.DBPath)
		assert.Equal(t, "test query", SampleTemplateData.SearchQuery)
		assert.Equal(t, 5, SampleTemplateData.ResultCount)
		assert.True(t, SampleTemplateData.HasSearch)
		assert.Equal(t, 1, SampleTemplateData.CurrentPage)
		assert.True(t, SampleTemplateData.HasNext)
		assert.False(t, SampleTemplateData.HasPrev)
		assert.Equal(t, "/?page=2&search=test", SampleTemplateData.NextPageURL)
		assert.Equal(t, "", SampleTemplateData.PrevPageURL)
		assert.Equal(t, []string{"web-server", "database", "cache-redis"}, SampleTemplateData.AvailableContainers)
		assert.Equal(t, []string{"web-server"}, SampleTemplateData.SelectedContainers)
		assert.True(t, SampleTemplateData.HasContainerFilter)
	})

	t.Run("should have consistent data relationships", func(t *testing.T) {
		// Result count should match logs length
		assert.Equal(t, len(SampleTemplateData.Logs), SampleTemplateData.ResultCount)

		// Available containers should include selected containers
		for _, selected := range SampleTemplateData.SelectedContainers {
			assert.Contains(t, SampleTemplateData.AvailableContainers, selected)
		}

		// HasContainerFilter should be true when containers are selected
		assert.Equal(t, len(SampleTemplateData.SelectedContainers) > 0, SampleTemplateData.HasContainerFilter)

		// HasSearch should be true when search query exists
		assert.Equal(t, SampleTemplateData.SearchQuery != "", SampleTemplateData.HasSearch)
	})
}
