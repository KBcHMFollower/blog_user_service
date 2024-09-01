package services_dep_interfaces

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
)

type SubscribersGetter interface {
	Subs(ctx context.Context, getInfo transfer.GetSubsInfo, tx database.Transaction) ([]*models.Subscriber, error)
	Count(ctx context.Context, info transfer.GetSubsCountInfo, tx database.Transaction) (uint32, error)
}

type SubscribersDealer interface {
	Unsubscribe(ctx context.Context, unsubInfo transfer.UnsubscribeInfo) error
	Subscribe(ctx context.Context, subInfo transfer.SubscribeToUserInfo) error
}
