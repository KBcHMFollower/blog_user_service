package services_dep_interfaces

import (
	"context"
	"database/sql"
)

type TransactionCreator interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
