package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogEntry_Enhanced(t *testing.T) {
	t.Run("should parse standard Docker log format", func(t *testing.T) {
		logMessage := "2024-01-15T10:30:45.123456789Z This is a test log message"
		containerName := "test-container"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45.123456789Z", timestamp)
		assert.Equal(t, "This is a test log message", message)
	})

	t.Run("should parse log with microsecond precision", func(t *testing.T) {
		logMessage := "2024-01-15T10:30:45.123456Z Application started successfully"
		containerName := "web-server"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45.123456Z", timestamp)
		assert.Equal(t, "Application started successfully", message)
	})

	t.Run("should parse log with millisecond precision", func(t *testing.T) {
		logMessage := "2024-01-15T10:30:45.123Z Database connection established"
		containerName := "database"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45.123Z", timestamp)
		assert.Equal(t, "Database connection established", message)
	})

	t.Run("should parse log without fractional seconds", func(t *testing.T) {
		logMessage := "2024-01-15T10:30:45Z Server listening on port 8080"
		containerName := "api-server"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45Z", timestamp)
		assert.Equal(t, "Server listening on port 8080", message)
	})

	t.Run("should handle log with multiple spaces after timestamp", func(t *testing.T) {
		logMessage := "2024-01-15T10:30:45.123456789Z    Multiple spaces before message"
		containerName := "test-container"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45.123456789Z", timestamp)
		assert.Equal(t, "Multiple spaces before message", message)
	})

	t.Run("should handle malformed timestamp gracefully", func(t *testing.T) {
		logMessage := "invalid-timestamp This is a log message"
		containerName := "test-container"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		// Should generate a timestamp and use the entire message
		assert.NotEmpty(t, timestamp)
		assert.Contains(t, timestamp, "T")
		assert.Contains(t, timestamp, "Z")
		assert.Equal(t, "invalid-timestamp This is a log message", message)
	})

	t.Run("should handle empty log message", func(t *testing.T) {
		logMessage := ""
		containerName := "test-container"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		// Should generate a timestamp and return empty message
		assert.NotEmpty(t, timestamp)
		assert.Contains(t, timestamp, "T")
		assert.Contains(t, timestamp, "Z")
		assert.Equal(t, "", message)
	})

	t.Run("should handle log message without timestamp", func(t *testing.T) {
		logMessage := "This is a log message without timestamp"
		containerName := "test-container"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		// Should generate a timestamp and use the entire message
		assert.NotEmpty(t, timestamp)
		assert.Contains(t, timestamp, "T")
		assert.Contains(t, timestamp, "Z")
		assert.Equal(t, "This is a log message without timestamp", message)
	})

	t.Run("should handle log with JSON content", func(t *testing.T) {
		logMessage := `2024-01-15T10:30:45.123456789Z {"level":"info","message":"User logged in","user_id":123}`
		containerName := "api-server"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45.123456789Z", timestamp)
		assert.Equal(t, `{"level":"info","message":"User logged in","user_id":123}`, message)
	})

	t.Run("should handle log with special characters", func(t *testing.T) {
		logMessage := "2024-01-15T10:30:45.123456789Z Error: Connection failed! @#$%^&*()_+-=[]{}|;':\",./<>?"
		containerName := "database"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45.123456789Z", timestamp)
		assert.Equal(t, "Error: Connection failed! @#$%^&*()_+-=[]{}|;':\",./<>?", message)
	})

	t.Run("should handle multiline log messages", func(t *testing.T) {
		logMessage := "2024-01-15T10:30:45.123456789Z Line 1\nLine 2\nLine 3"
		containerName := "test-container"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45.123456789Z", timestamp)
		assert.Equal(t, "Line 1\nLine 2\nLine 3", message)
	})

	t.Run("should handle different timezone formats", func(t *testing.T) {
		testCases := []struct {
			name        string
			logMsg      string
			expectedTS  string
			expectedMsg string
		}{
			{
				name:        "UTC timezone",
				logMsg:      "2024-01-15T10:30:45.123456789Z Message with UTC",
				expectedTS:  "2024-01-15T10:30:45.123456789Z",
				expectedMsg: "Message with UTC",
			},
			{
				name:        "Positive timezone offset (converted to UTC)",
				logMsg:      "2024-01-15T10:30:45.123456789+05:30 Message with positive offset",
				expectedTS:  "2024-01-15T05:00:45.123456789Z", // +05:30 converted to UTC
				expectedMsg: "Message with positive offset",
			},
			{
				name:        "Negative timezone offset (converted to UTC)",
				logMsg:      "2024-01-15T10:30:45.123456789-08:00 Message with negative offset",
				expectedTS:  "2024-01-15T18:30:45.123456789Z", // -08:00 converted to UTC
				expectedMsg: "Message with negative offset",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				timestamp, message := ParseLogEntry(tc.logMsg, "test-container")
				assert.Equal(t, tc.expectedTS, timestamp)
				assert.Equal(t, tc.expectedMsg, message)
			})
		}
	})

	t.Run("should handle edge cases with container names", func(t *testing.T) {
		testCases := []struct {
			name          string
			containerName string
			logMsg        string
		}{
			{
				name:          "empty container name",
				containerName: "",
				logMsg:        "2024-01-15T10:30:45.123456789Z Test message",
			},
			{
				name:          "container name with special characters",
				containerName: "test-container_123-abc.def",
				logMsg:        "2024-01-15T10:30:45.123456789Z Test message",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				timestamp, message := ParseLogEntry(tc.logMsg, tc.containerName)

				// Should parse correctly regardless of container name
				assert.Equal(t, "2024-01-15T10:30:45.123456789Z", timestamp)
				assert.Equal(t, "Test message", message)
			})
		}
	})
}
