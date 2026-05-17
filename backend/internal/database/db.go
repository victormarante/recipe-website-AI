package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// New creates a new database connection
func New(dbPath string) (*sqlx.DB, error) {
	// Create database file if it doesn't exist
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
	}

	// Open database connection
	db, err := sqlx.Connect("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(1) // SQLite works best with a single connection
	db.SetMaxIdleConns(1)

	// Enable foreign keys (important for SQLite)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return db, nil
}

// RunMigrations executes all SQL migration files
func RunMigrations(db *sqlx.DB, migrationPath string) error {
	// Read migration file
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// AddColumnIfNotExists adds a column to a table only if it doesn't already exist.
// Needed because SQLite does not support ALTER TABLE ... ADD COLUMN IF NOT EXISTS.
func AddColumnIfNotExists(db *sqlx.DB, table, column, definition string) error {
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name=?`, table, column); err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}
	if count > 0 {
		return nil
	}
	_, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition))
	if err != nil {
		return fmt.Errorf("failed to add column %s: %w", column, err)
	}
	return nil
}

// Close closes the database connection
func Close(db *sqlx.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// HealthCheck verifies database connectivity
func HealthCheck(db *sqlx.DB) error {
	var result int
	if err := db.Get(&result, "SELECT 1"); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}
