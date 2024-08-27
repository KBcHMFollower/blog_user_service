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
	BeginTxCtx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	Stop() error
}

type Transaction interface {
	Executor
	Commit() error
	Rollback() error
}

type Executor interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
