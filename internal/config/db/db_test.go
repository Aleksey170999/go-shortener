package db

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyMigrations(t *testing.T) {
	// Skip if test database is not configured
	if os.Getenv("TEST_DATABASE_DSN") == "" {
		t.Skip("Skipping test as TEST_DATABASE_DSN is not set")
	}

	// Create a test database connection
	db, err := sql.Open("postgres", os.Getenv("TEST_DATABASE_DSN"))
	require.NoError(t, err)
	defer db.Close()

	// Clean up any existing tables
	_, err = db.Exec(`
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO postgres;
		GRANT ALL ON SCHEMA public TO public;
	`)
	require.NoError(t, err)

	// Test applying migrations
	t.Run("apply migrations", func(t *testing.T) {
		err := ApplyMigrations(db)
		assert.NoError(t, err)

		// Verify that the migrations table exists
		var exists bool
		err = db.QueryRow(
			`SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'goose_db_version'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "goose_db_version table should exist after migrations")
	})

	// Test with invalid database connection
	t.Run("invalid database connection", func(t *testing.T) {
		invalidDB, err := sql.Open("postgres", "postgres://invalid:invalid@localhost:5432/invalid?sslmode=disable")
		require.NoError(t, err)
		defer invalidDB.Close()

		err = ApplyMigrations(invalidDB)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to apply migrations")
	})
}
