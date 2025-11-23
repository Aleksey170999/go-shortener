package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/pressly/goose/v3"
)

// ApplyMigrations applies all available database migrations using the goose migration tool.
// It sets the PostgreSQL dialect and runs all migrations from the "./migrations/" directory.
//
// Parameters:
//   - db: An open database connection to apply migrations to
//
// Returns:
//   - error: An error if any migration fails, nil if all migrations are applied successfully
//
// Note: The function expects migration files to be in the "./migrations/" directory
// relative to the working directory of the application.
func ApplyMigrations(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "./migrations/"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	fmt.Println("Migrations applied successfully!")
	return nil
}
