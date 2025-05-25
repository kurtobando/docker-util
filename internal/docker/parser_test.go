package docker

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestParseLogEntry_BasicFunctionality tests the core parsing functionality
func TestParseLogEntry_BasicFunctionality(t *testing.T) {
	t.Run("should parse valid RFC3339Nano timestamp", func(t *testing.T) {
		logMessage := "2023-12-01T10:30:45.123456789Z This is a test log message"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)
		assert.Equal(t, "2023-12-01T10:30:45.123456789Z", parsedTime.UTC().Format(time.RFC3339Nano))

		// Verify message is extracted correctly
		assert.Equal(t, "This is a test log message", messageStr)
	})

	t.Run("should parse timestamp without Z suffix", func(t *testing.T) {
		logMessage := "2023-12-01T10:30:45.123456789 This is a test log message"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Should fallback to second parsing format
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify message is extracted correctly
		assert.Equal(t, "This is a test log message", messageStr)
	})

	t.Run("should handle timestamp with microseconds", func(t *testing.T) {
		logMessage := "2023-12-01T10:30:45.123456Z This is a test log message"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify message is extracted correctly
		assert.Equal(t, "This is a test log message", messageStr)
	})

	t.Run("should handle timestamp with milliseconds", func(t *testing.T) {
		logMessage := "2023-12-01T10:30:45.123Z This is a test log message"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify message is extracted correctly
		assert.Equal(t, "This is a test log message", messageStr)
	})

	t.Run("should handle timestamp without fractional seconds", func(t *testing.T) {
		logMessage := "2024-01-15T10:30:45Z Server listening on port 8080"
		containerName := "api-server"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45Z", timestamp)
		assert.Equal(t, "Server listening on port 8080", message)
	})
}

// TestParseLogEntry_MessageFormatting tests message formatting and cleanup
func TestParseLogEntry_MessageFormatting(t *testing.T) {
	t.Run("should handle log message with trailing newlines", func(t *testing.T) {
		logMessage := "2023-12-01T10:30:45.123456789Z This is a test log message\n\r"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify message has trailing newlines removed
		assert.Equal(t, "This is a test log message", messageStr)
		assert.False(t, strings.HasSuffix(messageStr, "\n"))
		assert.False(t, strings.HasSuffix(messageStr, "\r"))
	})

	t.Run("should handle log message with multiple spaces", func(t *testing.T) {
		logMessage := "2023-12-01T10:30:45.123456789Z    This is a test log message with spaces   "
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify message is trimmed correctly
		assert.Equal(t, "This is a test log message with spaces", messageStr)
	})

	t.Run("should handle log with multiple spaces after timestamp", func(t *testing.T) {
		logMessage := "2024-01-15T10:30:45.123456789Z    Multiple spaces before message"
		containerName := "test-container"

		timestamp, message := ParseLogEntry(logMessage, containerName)

		assert.Equal(t, "2024-01-15T10:30:45.123456789Z", timestamp)
		assert.Equal(t, "Multiple spaces before message", message)
	})

	t.Run("should handle empty message after timestamp", func(t *testing.T) {
		logMessage := "2023-12-01T10:30:45.123456789Z "
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify message is empty
		assert.Equal(t, "", messageStr)
	})

	t.Run("should handle multiline log message", func(t *testing.T) {
		logMessage := "2023-12-01T10:30:45.123456789Z Line 1\nLine 2\nLine 3"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify message preserves internal newlines but removes trailing ones
		assert.Equal(t, "Line 1\nLine 2\nLine 3", messageStr)
	})
}

// TestParseLogEntry_ErrorHandling tests error conditions and fallback behavior
func TestParseLogEntry_ErrorHandling(t *testing.T) {
	t.Run("should fallback to current time for invalid timestamp", func(t *testing.T) {
		logMessage := "invalid-timestamp This is a test log message"
		containerName := "test-container"

		beforeCall := time.Now().UTC()
		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)
		afterCall := time.Now().UTC()

		// Verify timestamp is current time (within reasonable bounds)
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)
		assert.True(t, parsedTime.After(beforeCall.Add(-time.Second)))
		assert.True(t, parsedTime.Before(afterCall.Add(time.Second)))

		// Verify entire message is returned
		assert.Equal(t, "invalid-timestamp This is a test log message", messageStr)
	})

	t.Run("should fallback to current time for message without space", func(t *testing.T) {
		logMessage := "single-word-message"
		containerName := "test-container"

		beforeCall := time.Now().UTC()
		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)
		afterCall := time.Now().UTC()

		// Verify timestamp is current time (within reasonable bounds)
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)
		assert.True(t, parsedTime.After(beforeCall.Add(-time.Second)))
		assert.True(t, parsedTime.Before(afterCall.Add(time.Second)))

		// Verify entire message is returned
		assert.Equal(t, "single-word-message", messageStr)
	})

	t.Run("should handle empty log message", func(t *testing.T) {
		logMessage := ""
		containerName := "test-container"

		beforeCall := time.Now().UTC()
		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)
		afterCall := time.Now().UTC()

		// Verify timestamp is current time (within reasonable bounds)
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)
		assert.True(t, parsedTime.After(beforeCall.Add(-time.Second)))
		assert.True(t, parsedTime.Before(afterCall.Add(time.Second)))

		// Verify message is empty
		assert.Equal(t, "", messageStr)
	})

	t.Run("should handle log message with only whitespace", func(t *testing.T) {
		logMessage := "   \n\r\t   "
		containerName := "test-container"

		beforeCall := time.Now().UTC()
		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)
		afterCall := time.Now().UTC()

		// Verify timestamp is current time (within reasonable bounds)
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)
		assert.True(t, parsedTime.After(beforeCall.Add(-time.Second)))
		assert.True(t, parsedTime.Before(afterCall.Add(time.Second)))

		// Verify message is empty after trimming
		assert.Equal(t, "", messageStr)
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
}

// TestParseLogEntry_SpecialContent tests handling of special content types
func TestParseLogEntry_SpecialContent(t *testing.T) {
	t.Run("should handle log message with JSON content", func(t *testing.T) {
		logMessage := `2023-12-01T10:30:45.123456789Z {"level":"info","message":"test message","timestamp":"2023-12-01T10:30:45Z"}`
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify JSON message is extracted correctly
		expectedMessage := `{"level":"info","message":"test message","timestamp":"2023-12-01T10:30:45Z"}`
		assert.Equal(t, expectedMessage, messageStr)
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

	t.Run("should handle unicode characters in message", func(t *testing.T) {
		logMessage := "2023-12-01T10:30:45.123456789Z Unicode: 你好世界 🌍 émojis 🚀"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify unicode characters are preserved
		assert.Equal(t, "Unicode: 你好世界 🌍 émojis 🚀", messageStr)
	})
}

// TestParseLogEntry_TimezoneHandling tests different timezone formats
func TestParseLogEntry_TimezoneHandling(t *testing.T) {
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
}

// TestParseLogEntry_EdgeCases tests edge cases and performance scenarios
func TestParseLogEntry_EdgeCases(t *testing.T) {
	t.Run("should handle very long log message", func(t *testing.T) {
		longMessage := strings.Repeat("a", 10000)
		logMessage := "2023-12-01T10:30:45.123456789Z " + longMessage
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		// Verify timestamp is parsed correctly
		_, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Verify long message is handled correctly
		assert.Equal(t, longMessage, messageStr)
		assert.Equal(t, 10000, len(messageStr))
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
}
