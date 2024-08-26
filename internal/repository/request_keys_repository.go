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

	builder := rep_utils.QBuilder.PHFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	request := models.Request{
		Id:             uuid.New(),
		IdempotencyKey: info.Key,
	}

	query := builder.
		Ins(database.RequestKeysTable).
		Cols(rKeysIdCol, rKeysIdempotencyKeyCol).
		Vls(request.Id, request.IdempotencyKey)

	if _, err := query.ExcCtx(ctx, executor); err != nil {
		return fmt.Errorf("%s : failed to execute query: %w", op, err)
	}

	return nil
}

func (r *RequestsRepository) Get(ctx context.Context, key uuid.UUID, tx *sql.Tx) (*models.Request, error) {
	op := "RequestsRepository.get"

	builder := rep_utils.QBuilder.PHFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	query := builder.
		Sel(rKeysAllCol).
		Frm(database.RequestKeysTable).
		Wr(squirrel.Eq{rKeysIdempotencyKeyCol: key})

	var request models.Request

	if err := query.QryRowCtx(ctx, executor, &request); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s : failed to execute query : %w", op, err)
	}

	return &request, nil
}
