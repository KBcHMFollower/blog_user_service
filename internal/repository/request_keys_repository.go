package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	rep_utils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
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
	Create(ctx context.Context, info repositories_transfer.CreateRequestInfo, tx *sql.Tx) error
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

func (r *RequestsRepository) Create(ctx context.Context, info repositories_transfer.CreateRequestInfo, tx database.Transaction) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

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
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	if _, err := executor.ExecContext(ctx, toSql, args...); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return nil
}

func (r *RequestsRepository) Get(ctx context.Context, key uuid.UUID, tx database.Transaction) (*models.Request, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	query := builder.
		Select(rKeysAllCol).
		From(requestsTable).
		Where(squirrel.Eq{rKeysIdempotencyKeyCol: key})

	toSql, args, _ := query.ToSql()

	var request models.Request

	if err := executor.GetContext(ctx, &request, toSql, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return &request, nil
}
