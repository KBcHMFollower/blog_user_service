package repository

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	rep_utils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
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
	//todo: возможно executor
}

func NewSubscriberRepository(db database.Executor) *SubscribersRepository {
	return &SubscribersRepository{
		db:       db,
		qBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (sr *SubscribersRepository) getSubInfo(ctx context.Context, getInfo transfer.GetSubscriptionInfo) ([]*models.Subscriber, uint32, error) {
	offset := (getInfo.Page - 1) * getInfo.Size

	query := sr.qBuilder.
		Select(subsAllCol).
		From(subsTable).
		Where(squirrel.Eq{getInfo.TargetType: getInfo.UserId}).
		Limit(getInfo.Size).
		Offset(offset)

	toSql, args, err := query.ToSql()
	if err != nil {
		return nil, 0, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	subscribers := make([]*models.Subscriber, 0)
	if err := sr.db.SelectContext(ctx, &subscribers, toSql, args...); err != nil {
		return nil, 0, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	query = sr.qBuilder.
		Select("COUNT(*)").
		From(subsTable).
		Where(squirrel.Eq{getInfo.TargetType: getInfo.UserId})

	toSql, args, err = query.ToSql()
	if err != nil {
		return subscribers, 0, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	var totalCount uint32
	if err := sr.db.GetContext(ctx, &totalCount, toSql, args...); err != nil {
		return nil, 0, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return subscribers, totalCount, nil
}

func (sr *SubscribersRepository) Subs(ctx context.Context, getInfo transfer.GetSubsInfo) ([]*models.User, uint32, error) {
	subscribers, totalCount, err := sr.getSubInfo(ctx, transfer.GetSubscriptionInfo{
		UserId:     getInfo.UserId,
		Page:       getInfo.Page,
		Size:       getInfo.Size,
		TargetType: getInfo.Target,
	})
	if err != nil {
		return nil, totalCount, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	subscribersId := make([]uuid.UUID, 0)
	for _, subscriber := range subscribers {
		subscribersId = append(subscribersId, subscriber.SubscriberId)
	}

	sql, args, err := sr.qBuilder.Select(subsAllCol).
		From(subsTable).
		Where(squirrel.Eq{subsIdCol: subscribersId}).
		ToSql()
	if err != nil {
		return nil, totalCount, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	users := make([]*models.User, 0)
	if err := sr.db.SelectContext(ctx, &users, sql, args...); err != nil {
		return nil, totalCount, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return users, totalCount, nil
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
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	_, err = sr.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
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
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	_, err = sr.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return nil
}
