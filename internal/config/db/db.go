package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/pressly/goose/v3"
)

// PingDB verifies the connection to the database using the provided DSN.
// It attempts to establish a connection and ping the database to ensure it's reachable.
//
// Parameters:
//   - databaseDSN: Data Source Name for the database connection
//
// Returns:
//   - error: An error if the connection or ping fails, nil otherwise
func PingDB(databaseDSN string) error {
	db, err := sql.Open("postgres", databaseDSN)

	if err != nil {
		return fmt.Errorf("unable to connect to DB: %w", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("unable to ping DB: %w", err)
	}
	defer db.Close()

	return nil
}

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
