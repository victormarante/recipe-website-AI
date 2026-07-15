package database_test

import (
	"path/filepath"
	"testing"

	"recipe-backend/internal/database"
)

func TestRunMigrationsFreshDatabase(t *testing.T) {
	db, err := database.New(filepath.Join(t.TempDir(), "fresh.db"))
	if err != nil {
		t.Fatalf("database.New: %v", err)
	}
	t.Cleanup(func() { database.Close(db) })

	if err := database.RunMigrations(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
	assertColumnExists(t, db, "recipes", "oven_temperature")
	assertColumnExists(t, db, "recipes", "image_url")

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM schema_migrations`); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 migrations, got %d", count)
	}

	if err := database.RunMigrations(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("RunMigrations second run: %v", err)
	}
}

func TestRunMigrationsUpgradesOldDatabase(t *testing.T) {
	db, err := database.New(filepath.Join(t.TempDir(), "old.db"))
	if err != nil {
		t.Fatalf("database.New: %v", err)
	}
	t.Cleanup(func() { database.Close(db) })

	if _, err := db.Exec(`
		CREATE TABLE recipes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			categories TEXT NOT NULL,
			ingredients TEXT NOT NULL,
			steps TEXT NOT NULL,
			links TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO recipes (title, description, categories, ingredients, steps, links)
		VALUES ('Old Recipe', 'kept', '["dinner"]', '["salt"]', '["cook"]', '[]');
	`); err != nil {
		t.Fatalf("seed old schema: %v", err)
	}

	if err := database.RunMigrations(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	assertColumnExists(t, db, "recipes", "oven_temperature")
	assertColumnExists(t, db, "recipes", "image_url")

	var title string
	if err := db.Get(&title, `SELECT title FROM recipes WHERE id = 1`); err != nil {
		t.Fatalf("query existing recipe: %v", err)
	}
	if title != "Old Recipe" {
		t.Fatalf("existing data was not preserved: %q", title)
	}
}

func TestRunMigrationsHandlesLegacyRuntimeColumns(t *testing.T) {
	db, err := database.New(filepath.Join(t.TempDir(), "legacy.db"))
	if err != nil {
		t.Fatalf("database.New: %v", err)
	}
	t.Cleanup(func() { database.Close(db) })

	if _, err := db.Exec(`
		CREATE TABLE recipes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			categories TEXT NOT NULL,
			ingredients TEXT NOT NULL,
			steps TEXT NOT NULL,
			links TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			oven_temperature INTEGER DEFAULT NULL,
			image_url TEXT
		);
	`); err != nil {
		t.Fatalf("seed legacy schema: %v", err)
	}

	if err := database.RunMigrations(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
}

func assertColumnExists(t *testing.T, db interface {
	Get(dest interface{}, query string, args ...interface{}) error
}, table, column string) {
	t.Helper()

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, table, column); err != nil {
		t.Fatalf("check column %s.%s: %v", table, column, err)
	}
	if count != 1 {
		t.Fatalf("expected column %s.%s to exist", table, column)
	}
}
