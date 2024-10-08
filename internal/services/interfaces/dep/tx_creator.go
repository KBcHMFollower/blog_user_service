package services_dep_interfaces

import (
	"context"
	"database/sql"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
)

type TransactionCreator interface {
	BeginTxCtx(ctx context.Context, opts *sql.TxOptions) (database.Transaction, error)
}
