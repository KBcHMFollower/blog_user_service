package services

import (
	"context"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
	servicesutils "github.com/KBcHMFollower/blog_user_service/internal/services/lib"
	"github.com/google/uuid"
)

type subsSvcStore interface {
	dep.SubscribersGetter
	dep.SubscribersDealer
}

type subsUsrStore interface {
	dep.UserGetter
}

type SubscribersService struct {
	subsRep   subsSvcStore
	usersRep  subsUsrStore
	txCreator dep.TransactionCreator
	log       logger.Logger
}

func NewSubscribersService(subsRep subsSvcStore, usersRep subsUsrStore, txCreator dep.TransactionCreator, log logger.Logger) *SubscribersService {
	return &SubscribersService{
		subsRep:   subsRep,
		log:       log,
		usersRep:  usersRep,
		txCreator: txCreator,
	}
}

func (srs *SubscribersService) GetSubscribers(ctx context.Context, getInfo *transfer.GetSubscribersInfo) (resUsers *transfer.GetSubscribersResult, resErr error) {
	ctx = logger.UpdateLoggerCtx(ctx, bloggerIdLogKey, getInfo.BloggerId)

	srs.log.DebugContext(ctx, "try to get subscribers")

	tx, err := srs.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t begin transaction", err))
	}
	defer func() {
		resErr = servicesutils.HandleErrInTransaction(resErr, tx)
	}()

	subscribers, err := srs.subsRep.Subs(ctx, repositoriestransfer.GetSubsInfo{
		Condition: map[repositoriestransfer.GetSubType]any{
			repositoriestransfer.SubscribersTarget: getInfo.BloggerId,
		},
		Page: uint64(getInfo.Page),
		Size: uint64(getInfo.Size)}, tx)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user subscribers from db", err))
	}

	count, err := srs.subsRep.Count(ctx, repositoriestransfer.GetSubsCountInfo{
		Condition: map[repositoriestransfer.GetSubType]any{
			repositoriestransfer.SubscribersTarget: getInfo.BloggerId,
		},
	}, tx)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get subscribers count from db", err))
	}

	var usersIds []uuid.UUID
	for _, sub := range subscribers {
		usersIds = append(usersIds, sub.SubscriberId)
	}

	users, err := srs.usersRep.Users(ctx, &repositoriestransfer.GetUsersInfo{
		Condition: map[repositoriestransfer.UserFieldTarget]interface{}{
			repositoriestransfer.UserIdCondition: usersIds,
		},
	}, tx)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get users from db", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t commit transaction", err))
	}

	srs.log.DebugContext(ctx, "subscribers found in db")

	return &transfer.GetSubscribersResult{
		Subscribers: transfer.GetSubscribersArrayResultFromModel(users),
		TotalCount:  int32(count),
	}, nil
}

// todo: дублирующийся код
func (srs *SubscribersService) GetSubscriptions(ctx context.Context, getInfo *transfer.GetSubscriptionsInfo) (resUsers *transfer.GetSubscriptionsResult, resErr error) {
	ctx = logger.UpdateLoggerCtx(ctx, subscriberIdLogKey, getInfo.SubscriberId)

	srs.log.DebugContext(ctx, "try to get subscribers")

	tx, err := srs.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t begin transaction", err))
	}
	defer func() {
		resErr = servicesutils.HandleErrInTransaction(resErr, tx)
	}()

	subscribers, err := srs.subsRep.Subs(ctx, repositoriestransfer.GetSubsInfo{
		Condition: map[repositoriestransfer.GetSubType]any{
			repositoriestransfer.SubscriptionsTarget: getInfo.SubscriberId,
		},
		Page: uint64(getInfo.Page),
		Size: uint64(getInfo.Size)}, tx)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user subscribers from db", err))
	}

	count, err := srs.subsRep.Count(ctx, repositoriestransfer.GetSubsCountInfo{
		Condition: map[repositoriestransfer.GetSubType]any{
			repositoriestransfer.SubscriptionsTarget: getInfo.SubscriberId,
		},
	}, tx)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get subscribers count from db", err))
	}

	var usersIds []uuid.UUID
	for _, sub := range subscribers {
		usersIds = append(usersIds, sub.BloggerId)
	}

	users, err := srs.usersRep.Users(ctx, &repositoriestransfer.GetUsersInfo{
		Condition: map[repositoriestransfer.UserFieldTarget]interface{}{
			repositoriestransfer.UserIdCondition: usersIds,
		},
	}, tx)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get users from db", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t commit transaction", err))
	}

	srs.log.DebugContext(ctx, "subscribers found in db")

	return &transfer.GetSubscriptionsResult{
		Subscriptions: transfer.GetSubscribersArrayResultFromModel(users),
		TotalCount:    int32(count),
	}, nil
}

func (srs *SubscribersService) Subscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error {
	logger.UpdateLoggerCtx(ctx, subscriberIdLogKey, subInfo.SubscriberId)
	logger.UpdateLoggerCtx(ctx, bloggerIdLogKey, subInfo.BloggerId)

	srs.log.InfoContext(ctx, "try to subscribe to blogger")

	err := srs.subsRep.Subscribe(ctx, repositoriestransfer.SubscribeToUserInfo{
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
	ctx = logger.UpdateLoggerCtx(ctx, subscriberIdLogKey, subInfo.SubscriberId)
	ctx = logger.UpdateLoggerCtx(ctx, bloggerIdLogKey, subInfo.BloggerId)

	srs.log.InfoContext(ctx, "try to unsubscribe from blogger")

	err := srs.subsRep.Unsubscribe(ctx, repositoriestransfer.UnsubscribeInfo{
		BloggerId:    subInfo.BloggerId,
		SubscriberId: subInfo.SubscriberId,
	})
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `Unsubscribe`", err))
	}

	srs.log.InfoContext(ctx, "unsubscribed from blogger")

	return nil
}
