package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type DBDriver struct {
	*sqlx.DB
}

func (db *DBDriver) Stop() error {
	if err := db.Close(); err != nil {
		return fmt.Errorf("error in close process db connection : %w", err)
	}
	return nil
}

func (db *DBDriver) BeginTxCtx(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
	return db.DB.BeginTxx(ctx, opts)
}
