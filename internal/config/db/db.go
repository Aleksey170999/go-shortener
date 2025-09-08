package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func PingDB(databaseDSN string) error {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return fmt.Errorf("unable to connect to DB: %w", err)
	}
	defer db.Close()

	return nil
}
