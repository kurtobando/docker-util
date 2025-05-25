package docker

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseLogEntry(t *testing.T) {
	t.Run("should parse log with valid RFC3339Nano timestamp", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00.123456789Z This is a test log message"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00.123456789Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, "This is a test log message", messageStr)
	})

	t.Run("should parse log with RFC3339Nano timestamp without Z", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00.123456789 This is a test log message"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse("2006-01-02T15:04:05.000000000", "2024-01-01T12:00:00.123456789")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, "This is a test log message", messageStr)
	})

	t.Run("should handle log with simple RFC3339 timestamp", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00Z Simple log message"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, "Simple log message", messageStr)
	})

	t.Run("should handle log with trailing newlines", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00Z Log with newlines\n\r"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, "Log with newlines", messageStr)
	})

	t.Run("should handle log with multiple trailing newlines", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00Z Log with multiple newlines\n\n\r\n"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, "Log with multiple newlines", messageStr)
	})

	t.Run("should handle log with extra whitespace in message", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00Z    Log with extra spaces   "
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, "Log with extra spaces", messageStr)
	})

	t.Run("should fallback to current time for invalid timestamp", func(t *testing.T) {
		before := time.Now().UTC()
		logMessage := "invalid-timestamp This is a log message"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)
		after := time.Now().UTC()

		// Parse the returned timestamp
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Should be between before and after (current time)
		assert.True(t, parsedTime.After(before) || parsedTime.Equal(before))
		assert.True(t, parsedTime.Before(after) || parsedTime.Equal(after))
		// When parsing fails, it returns the whole message
		assert.Equal(t, "invalid-timestamp This is a log message", messageStr)
	})

	t.Run("should fallback to current time for log without space", func(t *testing.T) {
		before := time.Now().UTC()
		logMessage := "logmessagewithoutspace"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)
		after := time.Now().UTC()

		// Parse the returned timestamp
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Should be between before and after (current time)
		assert.True(t, parsedTime.After(before) || parsedTime.Equal(before))
		assert.True(t, parsedTime.Before(after) || parsedTime.Equal(after))
		assert.Equal(t, "logmessagewithoutspace", messageStr)
	})

	t.Run("should handle empty log message", func(t *testing.T) {
		before := time.Now().UTC()
		logMessage := ""
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)
		after := time.Now().UTC()

		// Parse the returned timestamp
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)

		// Should be between before and after (current time)
		assert.True(t, parsedTime.After(before) || parsedTime.Equal(before))
		assert.True(t, parsedTime.Before(after) || parsedTime.Equal(after))
		assert.Empty(t, messageStr)
	})

	t.Run("should handle log with only timestamp", func(t *testing.T) {
		before := time.Now().UTC()
		logMessage := "2024-01-01T12:00:00Z"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)
		after := time.Now().UTC()

		// Since there's no space after the timestamp, it falls back to current time
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
		assert.NoError(t, err)
		assert.True(t, parsedTime.After(before) || parsedTime.Equal(before))
		assert.True(t, parsedTime.Before(after) || parsedTime.Equal(after))
		// Returns the whole message when no space is found
		assert.Equal(t, "2024-01-01T12:00:00Z", messageStr)
	})

	t.Run("should handle log with timestamp and only spaces", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00Z    "
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Empty(t, messageStr)
	})

	t.Run("should handle multiline log message", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00Z Line 1\nLine 2\nLine 3"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, "Line 1\nLine 2\nLine 3", messageStr)
	})

	t.Run("should handle log with JSON content", func(t *testing.T) {
		logMessage := `2024-01-01T12:00:00Z {"level":"info","message":"test","timestamp":"2024-01-01T12:00:00Z"}`
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, `{"level":"info","message":"test","timestamp":"2024-01-01T12:00:00Z"}`, messageStr)
	})

	t.Run("should handle log with special characters", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00Z Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?", messageStr)
	})

	t.Run("should handle log with unicode characters", func(t *testing.T) {
		logMessage := "2024-01-01T12:00:00Z Unicode: 你好世界 🌍 🚀"
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, "Unicode: 你好世界 🌍 🚀", messageStr)
	})

	t.Run("should handle very long log message", func(t *testing.T) {
		longMessage := strings.Repeat("A", 10000)
		logMessage := "2024-01-01T12:00:00Z " + longMessage
		containerName := "test-container"

		timestampStr, messageStr := ParseLogEntry(logMessage, containerName)

		expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-01T12:00:00Z")
		expectedTimestamp := expectedTime.UTC().Format(time.RFC3339Nano)

		assert.Equal(t, expectedTimestamp, timestampStr)
		assert.Equal(t, longMessage, messageStr)
	})
}
