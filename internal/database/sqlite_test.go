package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabase(t *testing.T) {
	t.Run("should create new database instance", func(t *testing.T) {
		dbPath := "/test/path.db"
		db := NewDatabase(dbPath)

		assert.NotNil(t, db)
		assert.Equal(t, dbPath, db.dbPath)
		assert.Nil(t, db.db) // Should be nil until Connect() is called
	})

	t.Run("should handle empty path", func(t *testing.T) {
		db := NewDatabase("")

		assert.NotNil(t, db)
		assert.Equal(t, "", db.dbPath)
		assert.Nil(t, db.db)
	})
}

func TestDatabase_Connect(t *testing.T) {
	// Create temporary directory for test databases
	tempDir := t.TempDir()

	t.Run("should connect and create schema successfully", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "test.db")
		db := NewDatabase(dbPath)

		err := db.Connect()
		require.NoError(t, err)
		defer db.Close()

		// Verify database connection is established
		assert.NotNil(t, db.db)

		// Verify tables were created
		var tableName string
		err = db.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='logs'").Scan(&tableName)
		require.NoError(t, err)
		assert.Equal(t, "logs", tableName)

		// Verify indexes were created
		rows, err := db.db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='logs'")
		require.NoError(t, err)
		defer rows.Close()

		indexes := make([]string, 0)
		for rows.Next() {
			var indexName string
			err := rows.Scan(&indexName)
			require.NoError(t, err)
			indexes = append(indexes, indexName)
		}

		// Should have at least our custom indexes (SQLite may create additional ones)
		assert.Contains(t, indexes, "idx_logs_message")
		assert.Contains(t, indexes, "idx_logs_container_name")
	})

	t.Run("should handle connecting to existing database", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "existing.db")

		// Create database first time
		db1 := NewDatabase(dbPath)
		err := db1.Connect()
		require.NoError(t, err)
		db1.Close()

		// Connect to existing database
		db2 := NewDatabase(dbPath)
		err = db2.Connect()
		require.NoError(t, err)
		defer db2.Close()

		assert.NotNil(t, db2.db)

		// Verify table still exists
		var tableName string
		err = db2.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='logs'").Scan(&tableName)
		require.NoError(t, err)
		assert.Equal(t, "logs", tableName)
	})

	t.Run("should handle invalid database path", func(t *testing.T) {
		// Try to create database in non-existent directory without permission
		dbPath := "/root/nonexistent/test.db"
		db := NewDatabase(dbPath)

		err := db.Connect()
		// Should get an error (exact error depends on system)
		assert.Error(t, err)
		// Note: db.db might not be nil even on error due to SQLite behavior
	})

	t.Run("should handle in-memory database", func(t *testing.T) {
		db := NewDatabase(":memory:")

		err := db.Connect()
		require.NoError(t, err)
		defer db.Close()

		assert.NotNil(t, db.db)

		// Verify table was created in memory
		var tableName string
		err = db.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='logs'").Scan(&tableName)
		require.NoError(t, err)
		assert.Equal(t, "logs", tableName)
	})
}

func TestDatabase_Close(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("should close database connection successfully", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "close_test.db")
		db := NewDatabase(dbPath)

		err := db.Connect()
		require.NoError(t, err)

		// Verify connection is active
		assert.NotNil(t, db.db)
		err = db.db.Ping()
		require.NoError(t, err)

		// Close the database
		err = db.Close()
		assert.NoError(t, err)
	})

	t.Run("should handle closing nil database", func(t *testing.T) {
		db := NewDatabase("test.db")
		// Don't call Connect(), so db.db remains nil

		err := db.Close()
		assert.NoError(t, err) // Should not error on nil database
	})

	t.Run("should handle multiple close calls", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "multi_close_test.db")
		db := NewDatabase(dbPath)

		err := db.Connect()
		require.NoError(t, err)

		// Close multiple times
		err1 := db.Close()
		err2 := db.Close()

		assert.NoError(t, err1)
		// Second close might error, but shouldn't panic
		// SQLite typically returns "sql: database is closed" error
		_ = err2 // We don't assert on this as behavior may vary
	})
}

func TestDatabase_GetDB(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("should return database connection", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "getdb_test.db")
		db := NewDatabase(dbPath)

		err := db.Connect()
		require.NoError(t, err)
		defer db.Close()

		sqlDB := db.GetDB()
		assert.NotNil(t, sqlDB)
		assert.IsType(t, &sql.DB{}, sqlDB)

		// Verify we can use the returned connection
		err = sqlDB.Ping()
		assert.NoError(t, err)
	})

	t.Run("should return nil for unconnected database", func(t *testing.T) {
		db := NewDatabase("test.db")
		// Don't call Connect()

		sqlDB := db.GetDB()
		assert.Nil(t, sqlDB)
	})
}

func TestDatabase_SchemaValidation(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("should create correct table schema", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "schema_test.db")
		db := NewDatabase(dbPath)

		err := db.Connect()
		require.NoError(t, err)
		defer db.Close()

		// Check table schema
		rows, err := db.db.Query("PRAGMA table_info(logs)")
		require.NoError(t, err)
		defer rows.Close()

		columns := make(map[string]string)
		for rows.Next() {
			var cid int
			var name, dataType, notNull, pk string
			var defaultValue sql.NullString
			err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
			require.NoError(t, err)
			columns[name] = dataType
		}

		// Verify all required columns exist with correct types
		assert.Equal(t, "INTEGER", columns["id"])
		assert.Equal(t, "TEXT", columns["container_id"])
		assert.Equal(t, "TEXT", columns["container_name"])
		assert.Equal(t, "TEXT", columns["timestamp"])
		assert.Equal(t, "TEXT", columns["stream_type"])
		assert.Equal(t, "TEXT", columns["log_message"])
	})

	t.Run("should handle database file permissions", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping permission test when running as root")
		}

		dbPath := filepath.Join(tempDir, "permission_test.db")
		db := NewDatabase(dbPath)

		err := db.Connect()
		require.NoError(t, err)
		defer db.Close()

		// Change file permissions to read-only
		err = os.Chmod(dbPath, 0444)
		require.NoError(t, err)

		// Try to create a new connection to the read-only file
		db2 := NewDatabase(dbPath)
		err = db2.Connect()
		// Should still be able to connect for reading
		assert.NoError(t, err)
		if err == nil {
			db2.Close()
		}

		// Restore permissions for cleanup
		os.Chmod(dbPath, 0644)
	})
}
