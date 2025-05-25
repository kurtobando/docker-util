package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Save original command line args and restore after tests
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	t.Run("should load default configuration", func(t *testing.T) {
		// Reset flag.CommandLine for this test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test"}

		config := Load()

		assert.Equal(t, "./docker_logs.db", config.DBPath)
		assert.Equal(t, "9123", config.ServerPort)
	})

	t.Run("should load custom dbpath from flags", func(t *testing.T) {
		// Reset flag.CommandLine for this test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-dbpath", "/custom/path/logs.db"}

		config := Load()

		assert.Equal(t, "/custom/path/logs.db", config.DBPath)
		assert.Equal(t, "9123", config.ServerPort) // Should still be default
	})

	t.Run("should load custom port from flags", func(t *testing.T) {
		// Reset flag.CommandLine for this test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-port", "8080"}

		config := Load()

		assert.Equal(t, "./docker_logs.db", config.DBPath) // Should still be default
		assert.Equal(t, "8080", config.ServerPort)
	})

	t.Run("should load both custom dbpath and port from flags", func(t *testing.T) {
		// Reset flag.CommandLine for this test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-dbpath", "/custom/logs.db", "-port", "3000"}

		config := Load()

		assert.Equal(t, "/custom/logs.db", config.DBPath)
		assert.Equal(t, "3000", config.ServerPort)
	})

	t.Run("should handle flags in different order", func(t *testing.T) {
		// Reset flag.CommandLine for this test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-port", "4000", "-dbpath", "/another/path.db"}

		config := Load()

		assert.Equal(t, "/another/path.db", config.DBPath)
		assert.Equal(t, "4000", config.ServerPort)
	})

	t.Run("should handle empty string values", func(t *testing.T) {
		// Reset flag.CommandLine for this test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-dbpath", "", "-port", ""}

		config := Load()

		assert.Equal(t, "", config.DBPath)
		assert.Equal(t, "", config.ServerPort)
	})

	t.Run("should handle special characters in paths", func(t *testing.T) {
		// Reset flag.CommandLine for this test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-dbpath", "/path with spaces/logs-2024.db", "-port", "9999"}

		config := Load()

		assert.Equal(t, "/path with spaces/logs-2024.db", config.DBPath)
		assert.Equal(t, "9999", config.ServerPort)
	})
}

func TestConfigStruct(t *testing.T) {
	t.Run("should create Config struct with values", func(t *testing.T) {
		config := &Config{
			DBPath:     "/test/path.db",
			ServerPort: "8080",
		}

		assert.Equal(t, "/test/path.db", config.DBPath)
		assert.Equal(t, "8080", config.ServerPort)
	})

	t.Run("should create empty Config struct", func(t *testing.T) {
		config := &Config{}

		assert.Empty(t, config.DBPath)
		assert.Empty(t, config.ServerPort)
	})

	t.Run("should handle zero values", func(t *testing.T) {
		var config Config

		assert.Empty(t, config.DBPath)
		assert.Empty(t, config.ServerPort)
	})
}
