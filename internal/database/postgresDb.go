package database

import (
	"database/sql"
	"fmt"
)

type DBDriver struct {
	*sql.DB
}

func (db *DBDriver) Stop() error {
	if err := db.Close(); err != nil {
		return fmt.Errorf("error in close process db connection : %w", err)
	}

	return nil
}
