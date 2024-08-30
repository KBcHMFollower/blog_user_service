package services_interfaces

import (
	"context"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
)

type SubsService interface {
	GetSubscribers(ctx context.Context, getInfo *transfer.GetSubscribersInfo) (*transfer.GetSubscribersResult, error)
	GetSubscriptions(ctx context.Context, getInfo *transfer.GetSubscriptionsInfo) (*transfer.GetSubscriptionsResult, error)
	Subscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error
	Unsubscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error
}
