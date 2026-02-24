# Repository Tests

This document outlines the test strategy for the repository layer once Go is installed.

## Running Tests

To run repository tests once Go is installed:

```bash
# Run all tests
go test ./internal/repository/... -v

# Run with coverage
go test ./internal/repository/... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Structure

The repository tests should:

1. Use in-memory SQLite database (`:memory:`)
2. Test all CRUD operations
3. Test filtering and search functionality
4. Test edge cases (not found, duplicates, etc.)

## Example Test Setup

```go
package repository_test

import (
    "testing"
    "recipe-backend/internal/database"
    "recipe-backend/internal/repository"
)

func setupTestDB(t *testing.T) *sqlx.DB {
    db, err := database.New(":memory:")
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    
    // Run migrations
    err = database.RunMigrations(db, "../../migrations/001_create_tables.sql")
    if err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
    }
    
    return db
}
```

This will be implemented once Go is installed on the system.
