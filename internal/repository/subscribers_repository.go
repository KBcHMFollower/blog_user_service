package repository

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	reputils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
	"github.com/Masterminds/squirrel"
)

const (
	subsTable = "subscribers"
)

const (
	subsIdCol           = "id"
	subsAllCol          = "*"
	subsBloggerIdCol    = "blogger_id"
	subsSubscriberIdCol = "subscriber_id"
)

type SubscribersRepository struct {
	db       database.Executor
	qBuilder squirrel.StatementBuilderType
}

func NewSubscriberRepository(db database.Executor) *SubscribersRepository {
	return &SubscribersRepository{
		db:       db,
		qBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (sr *SubscribersRepository) Count(ctx context.Context, info transfer.GetSubsCountInfo, tx database.Transaction) (uint32, error) {
	executor := reputils.GetExecutor(sr.db, tx)

	query := sr.qBuilder.
		Select("COUNT(*)").
		From(subsTable).
		Where(squirrel.Eq(reputils.ConvertMapKeysToStrings(info.Condition)))

	toSql, args, err := query.ToSql()
	if err != nil {
		return 0, reputils.ReturnGenerateSqlError(ctx, err)
	}

	var totalCount uint32
	if err := executor.GetContext(ctx, &totalCount, toSql, args...); err != nil {
		return 0, reputils.ReturnExecuteSqlError(ctx, err)
	}

	return totalCount, nil
}

func (sr *SubscribersRepository) Subs(ctx context.Context, getInfo transfer.GetSubsInfo, tx database.Transaction) ([]*models.Subscriber, error) {
	executor := reputils.GetExecutor(sr.db, tx)

	offset := (getInfo.Page - 1) * getInfo.Size

	query := sr.qBuilder.
		Select(subsAllCol).
		From(subsTable).
		Where(squirrel.Eq(reputils.ConvertMapKeysToStrings(getInfo.Condition))).
		Limit(getInfo.Size).
		Offset(offset)

	toSql, args, err := query.ToSql()
	if err != nil {
		return nil, reputils.ReturnGenerateSqlError(ctx, err)
	}

	subscribers := make([]*models.Subscriber, 0)
	if err := executor.SelectContext(ctx, &subscribers, toSql, args...); err != nil {
		return nil, reputils.ReturnExecuteSqlError(ctx, err)
	}

	return subscribers, nil
}

func (sr *SubscribersRepository) Subscribe(ctx context.Context, subInfo transfer.SubscribeToUserInfo) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	subscribers := models.NewSubscriber(subInfo.BloggerId, subInfo.SubscriberId)

	query := builder.Insert(subsTable).
		SetMap(map[string]interface{}{
			subsIdCol:           subscribers.Id,
			subsBloggerIdCol:    subscribers.BloggerId,
			subsSubscriberIdCol: subscribers.SubscriberId,
		})

	sql, args, err := query.ToSql()
	if err != nil {
		return reputils.ReturnGenerateSqlError(ctx, err)
	}

	_, err = sr.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return reputils.ReturnExecuteSqlError(ctx, err)
	}

	return nil
}

func (sr *SubscribersRepository) Unsubscribe(ctx context.Context, unsubInfo transfer.UnsubscribeInfo) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.Delete(subsTable).
		Where(squirrel.Eq{
			subsBloggerIdCol:    unsubInfo.BloggerId,
			subsSubscriberIdCol: unsubInfo.SubscriberId,
		})

	sql, args, err := query.ToSql()
	if err != nil {
		return reputils.ReturnGenerateSqlError(ctx, err)
	}

	_, err = sr.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return reputils.ReturnExecuteSqlError(ctx, err)
	}

	return nil
}
