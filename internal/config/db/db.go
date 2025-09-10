package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/pressly/goose/v3"
)

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
