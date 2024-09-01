package rep_utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	"github.com/lib/pq"
)

const (
	FailedToGenerateSqlMessage = "failed to generate sql query"
	FailedToExecuteQuery       = "failed to execute query"
)

func GetExecutor(r database.Executor, tx database.Transaction) database.Executor {
	if tx == nil {
		return r
	}
	return tx
}

func ReturnGenerateSqlError(ctx context.Context, err error) error {
	return ctxerrors.WrapCtx(
		ctx,
		ctxerrors.Wrap(
			FailedToGenerateSqlMessage,
			returnStatusError(err),
		),
	)
}

func ReturnExecuteSqlError(ctx context.Context, err error) error {
	return ctxerrors.WrapCtx(
		ctx,
		ctxerrors.Wrap(
			FailedToExecuteQuery,
			returnStatusError(err),
		),
	)
}

func returnStatusError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ctxerrors.ErrNotFound
	}

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ctxerrors.Wrap(fmt.Sprintf("err unique constraint: %s", pgErr.Message), ctxerrors.ErrConflict)
		}
	}

	return err
}

func ConvertMapKeysToStrings[T ~string](m map[T]any) map[string]any {
	result := make(map[string]any)
	for k, v := range m {
		result[string(k)] = v
	}

	return result
}

func GetPageAndSize(page uint64, size uint64) (uint64, uint64) {
	resPage := uint64(1)
	resSize := uint64(0)

	if page > 0 {
		resPage = page
	}
	if size > 0 {
		resSize = size
	}

	return resPage, resSize
}
