package docker

import (
	"strings"
	"time"
)

// ParseLogEntry extracts timestamp and message from a raw log line
// This preserves the exact same parsing logic from the original parseLogEntry function
func ParseLogEntry(logMessage string, containerName string) (timestampStr string, messageStr string) {
	logMessage = strings.TrimRight(logMessage, "\n\r") // Clean trailing newlines first

	firstSpaceIndex := strings.Index(logMessage, " ")
	if firstSpaceIndex > 0 {
		timestampCandidate := logMessage[:firstSpaceIndex]
		// Docker's own timestamp format with nanoseconds sometimes has 'Z' and sometimes doesn't directly.
		// time.RFC3339Nano expects 'Z' or a numeric offset.
		// We need to be a bit flexible or ensure Docker consistently provides 'Z'.
		// For now, assuming it's mostly compliant or we fall back.
		parsedTime, err := time.Parse(time.RFC3339Nano, timestampCandidate)
		if err == nil {
			return parsedTime.UTC().Format(time.RFC3339Nano), strings.TrimSpace(logMessage[firstSpaceIndex+1:])
		}
		// Fallback for slightly different formats, e.g. missing 'Z' but otherwise ok (common in some loggers)
		parsedTime, err = time.Parse("2006-01-02T15:04:05.000000000", timestampCandidate) // Example without Z
		if err == nil {
			return parsedTime.UTC().Format(time.RFC3339Nano), strings.TrimSpace(logMessage[firstSpaceIndex+1:])
		}
	}

	// Fallback: If no space or parsing failed, use current time in UTC and the whole message.
	// log.Printf("Debug: Could not parse timestamp from log line for %s. Line: '%s'. Using current time.", containerName, logMessage)
	return time.Now().UTC().Format(time.RFC3339Nano), strings.TrimSpace(logMessage)
}
