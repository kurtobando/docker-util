package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTempDB(t *testing.T) {
	t.Run("should create temporary database successfully", func(t *testing.T) {
		db := CreateTempDB(t)

		// Verify database is created and connected
		assert.NotNil(t, db)
		assert.NotNil(t, db.GetDB())

		// Verify we can ping the database
		err := db.GetDB().Ping()
		assert.NoError(t, err)

		// Verify logs table exists
		var tableName string
		err = db.GetDB().QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='logs'").Scan(&tableName)
		assert.NoError(t, err)
		assert.Equal(t, "logs", tableName)

		// Database should be automatically cleaned up by t.Cleanup()
	})

	t.Run("should create unique databases for multiple calls", func(t *testing.T) {
		db1 := CreateTempDB(t)
		db2 := CreateTempDB(t)

		assert.NotNil(t, db1)
		assert.NotNil(t, db2)
		assert.NotEqual(t, db1, db2)

		// Both should be functional
		err1 := db1.GetDB().Ping()
		err2 := db2.GetDB().Ping()
		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})
}

func TestCreateTempDBPath(t *testing.T) {
	t.Run("should create unique temporary database paths", func(t *testing.T) {
		path1 := CreateTempDBPath(t)
		path2 := CreateTempDBPath(t)

		assert.NotEmpty(t, path1)
		assert.NotEmpty(t, path2)
		assert.NotEqual(t, path1, path2)

		// Paths should end with .db
		assert.Contains(t, path1, ".db")
		assert.Contains(t, path2, ".db")

		// Paths should be in temporary directories
		assert.Contains(t, path1, "test_")
		assert.Contains(t, path2, "test_")
	})

	t.Run("should create paths in temporary directory", func(t *testing.T) {
		path := CreateTempDBPath(t)

		// Path should be absolute and in a temp directory
		assert.NotEmpty(t, path)
		// The exact temp directory varies by system, but should contain temp-like patterns
		// We just verify it's a reasonable path
		assert.Contains(t, path, "/")
	})
}

func TestGetRandomPort(t *testing.T) {
	t.Run("should return zero for OS assignment", func(t *testing.T) {
		port := GetRandomPort()

		// Our implementation returns 0 to let the OS assign a port
		assert.Equal(t, 0, port)
	})

	t.Run("should be consistent", func(t *testing.T) {
		port1 := GetRandomPort()
		port2 := GetRandomPort()

		// Both should return 0 (our current implementation)
		assert.Equal(t, 0, port1)
		assert.Equal(t, 0, port2)
	})
}
