package database

import (
	"database/sql"
	"fmt"
)

type DBDriver struct {
	*sql.DB
}

func New(connectionString string) (*DBDriver, *sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, nil, fmt.Errorf("Error in process db connection : %w", err)
	}
	return &DBDriver{db}, db, nil
}

func (db *DBDriver) Stop() error {
	if err := db.Close(); err != nil {
		return fmt.Errorf("error in close process db connection : %w", err)
	}

	return nil
}
