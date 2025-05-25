package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Database wraps the SQLite database connection
type Database struct {
	db     *sql.DB
	dbPath string
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) *Database {
	return &Database{
		dbPath: dbPath,
	}
}

// Connect opens the database connection and initializes the schema
// This preserves the exact same logic from the original initDB() function
func (d *Database) Connect() error {
	var err error
	d.db, err = sql.Open("sqlite3", d.dbPath+"?_busy_timeout=5000") // Added busy_timeout
	if err != nil {
		return fmt.Errorf("failed to open database at %s: %w", d.dbPath, err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		container_id TEXT NOT NULL,
		container_name TEXT NOT NULL,
		timestamp TEXT NOT NULL,
		stream_type TEXT NOT NULL, 
		log_message TEXT NOT NULL
	);`

	_, err = d.db.Exec(createTableSQL)
	if err != nil {
		d.db.Close()
		return fmt.Errorf("failed to create logs table in %s: %w", d.dbPath, err)
	}

	// Add index on log_message for search performance
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_logs_message ON logs(log_message);`
	_, err = d.db.Exec(createIndexSQL)
	if err != nil {
		d.db.Close()
		return fmt.Errorf("failed to create index on log_message in %s: %w", d.dbPath, err)
	}

	// Add index on container_name for filtering performance
	createContainerIndexSQL := `CREATE INDEX IF NOT EXISTS idx_logs_container_name ON logs(container_name);`
	_, err = d.db.Exec(createContainerIndexSQL)
	if err != nil {
		d.db.Close()
		return fmt.Errorf("failed to create index on container_name in %s: %w", d.dbPath, err)
	}

	log.Printf("Database initialized at %s and logs table ensured.", d.dbPath)
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetDB returns the underlying sql.DB instance
func (d *Database) GetDB() *sql.DB {
	return d.db
}
