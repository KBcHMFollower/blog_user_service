package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DBWrapper interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type DBDriver struct {
	*sql.DB
}

func New(connectionString string) (*DBDriver, *sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, nil, fmt.Errorf("Error in process db connection : %v", err)
	}

	return &DBDriver{db}, db, nil
}
