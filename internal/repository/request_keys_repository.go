package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	rep_utils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
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

func (r *RequestsRepository) Create(ctx context.Context, info repositories_transfer.CreateRequestInfo, tx *sql.Tx) error {
	op := "RequestsRepository.create"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	request := models.Request{
		Id:             uuid.New(),
		IdempotencyKey: info.Key,
	}

	query := builder.
		Insert(database.RequestKeysTable).
		Columns(rKeysIdCol, rKeysIdempotencyKeyCol).
		Values(request.Id, request.IdempotencyKey)

	toSql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s : failed to generate sql query : %w", op, err)
	}

	res := executor.QueryRowContext(ctx, toSql, args...)
	if err := res.Err(); err != nil {
		return fmt.Errorf("%s : failed to execute query : %w", op, err)
	}

	return nil
}

func (r *RequestsRepository) Get(ctx context.Context, key uuid.UUID, tx *sql.Tx) (*models.Request, error) {
	op := "RequestsRepository.get"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	query := builder.
		Select(rKeysAllCol).
		From(database.RequestKeysTable).
		Where(squirrel.Eq{rKeysIdempotencyKeyCol: key})

	toSql, args, _ := query.ToSql()

	var request models.Request

	row := executor.QueryRowContext(ctx, toSql, args...)
	err := row.Scan(request.GetPointersArray()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s : failed to execute query : %w", op, err)
	}

	return &request, nil
}
