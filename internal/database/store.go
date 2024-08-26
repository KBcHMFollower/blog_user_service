package database

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
)

const (
	RequestKeysTable = "request_keys"
)

type DBWrapper interface {
	Executor
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Stop() error
}

type Executor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
