package repository

import (
	"context"
	"database/sql"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	reputils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const (
	requestsTable = "request_keys"
)

const (
	rKeysIdCol             = "id"
	rKeysIdempotencyKeyCol = "idempotency_key"
	rKeysPayloadCol        = "payload"
	rKeysAllCol            = "*"
)

type RequestsStore interface {
	Create(ctx context.Context, info repositoriestransfer.CreateRequestInfo, tx *sql.Tx) error
	Get(ctx context.Context, key uuid.UUID, tx *sql.Tx) (*models.Request, error)
}

type RequestsRepository struct {
	db database.DBWrapper
}

func NewRequestsRepository(db database.DBWrapper) *RequestsRepository {
	return &RequestsRepository{
		db: db,
	}
}

func (r *RequestsRepository) Create(ctx context.Context, info repositoriestransfer.CreateRequestInfo, tx database.Transaction) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := reputils.GetExecutor(r.db, tx)

	request := models.Request{
		Id:             uuid.New(),
		IdempotencyKey: info.Key,
	}

	query := builder.
		Insert(requestsTable).
		Columns(rKeysIdCol, rKeysIdempotencyKeyCol).
		Values(request.Id, request.IdempotencyKey)

	toSql, args, err := query.ToSql()
	if err != nil {
		return reputils.ReturnGenerateSqlError(ctx, err)
	}

	if _, err := executor.ExecContext(ctx, toSql, args...); err != nil {
		return reputils.ReturnExecuteSqlError(ctx, err)
	}

	return nil
}

func (r *RequestsRepository) Get(ctx context.Context, key uuid.UUID, tx database.Transaction) (*models.Request, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := reputils.GetExecutor(r.db, tx)

	query := builder.
		Select(rKeysAllCol).
		From(requestsTable).
		Where(squirrel.Eq{rKeysIdempotencyKeyCol: key})

	toSql, args, _ := query.ToSql()

	var request models.Request

	if err := executor.GetContext(ctx, &request, toSql, args...); err != nil {
		return nil, reputils.ReturnExecuteSqlError(ctx, err)
	}

	return &request, nil
}
