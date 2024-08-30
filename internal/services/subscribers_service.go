package services

import (
	"context"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
	"log/slog"
)

type subsStore interface {
	dep.SubscribersGetter
	dep.SubscribersDealer
}

type SubscribersService struct {
	subsRep subsStore
	log     *slog.Logger
}

func NewSubscribersService(subsRep subsStore, log *slog.Logger) *SubscribersService {
	return &SubscribersService{
		subsRep: subsRep,
		log:     log,
	}
}

func (srs *SubscribersService) GetSubscribers(ctx context.Context, getInfo *transfer.GetSubscribersInfo) (*transfer.GetSubscribersResult, error) {
	logger.UpdateLoggerCtx(ctx, bloggerIdLogKey, getInfo.BloggerId)

	srs.log.Debug("try to get subscribers")

	users, totalCount, err := srs.subsRep.Subs(ctx, repositories_transfer.GetSubsInfo{
		Target: repositories_transfer.SubscribersTarget,
		UserId: getInfo.BloggerId,
		Page:   uint64(getInfo.Page),
		Size:   uint64(getInfo.Size)})
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user subscribers from db", err))
	}

	srs.log.Debug("subscribers found in db")

	return &transfer.GetSubscribersResult{
		Subscribers: transfer.GetSubscribersArrayResultFromModel(users),
		TotalCount:  int32(totalCount),
	}, nil
}

func (srs *SubscribersService) GetSubscriptions(ctx context.Context, getInfo *transfer.GetSubscriptionsInfo) (*transfer.GetSubscriptionsResult, error) {
	srs.log.Debug("try to get subscriptions")

	users, totalCount, err := srs.subsRep.Subs(ctx, repositories_transfer.GetSubsInfo{
		Target: repositories_transfer.SubscriptionsTarget,
		UserId: getInfo.SubscriberId,
		Page:   uint64(getInfo.Page),
		Size:   uint64(getInfo.Size),
	})
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user subscriptions from db", err))
	}

	srs.log.Debug("subscribers found in db")

	return &transfer.GetSubscriptionsResult{
		Subscriptions: transfer.GetSubscribersArrayResultFromModel(users),
		TotalCount:    int32(totalCount),
	}, nil
}

func (srs *SubscribersService) Subscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error {
	logger.UpdateLoggerCtx(ctx, subscriberIdLogKey, subInfo.SubscriberId)
	logger.UpdateLoggerCtx(ctx, bloggerIdLogKey, subInfo.BloggerId)

	srs.log.InfoContext(ctx, "try to subscribe to blogger")

	err := srs.subsRep.Subscribe(ctx, repositories_transfer.SubscribeToUserInfo{
		BloggerId:    subInfo.BloggerId,
		SubscriberId: subInfo.SubscriberId,
	})
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `Subscribe`", err))
	}

	srs.log.InfoContext(ctx, "subscribed to blogger")

	return nil
}

func (srs *SubscribersService) Unsubscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error {
	logger.UpdateLoggerCtx(ctx, subscriberIdLogKey, subInfo.SubscriberId)
	logger.UpdateLoggerCtx(ctx, bloggerIdLogKey, subInfo.BloggerId)

	srs.log.InfoContext(ctx, "try to unsubscribe from blogger")

	err := srs.subsRep.Unsubscribe(ctx, repositories_transfer.UnsubscribeInfo{
		BloggerId:    subInfo.BloggerId,
		SubscriberId: subInfo.SubscriberId,
	})
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `Unsubscribe`", err))
	}

	srs.log.InfoContext(ctx, "unsubscribed from blogger")

	return nil
}
