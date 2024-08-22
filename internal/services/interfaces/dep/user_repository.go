package services_dep_interfaces

import (
	"context"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type UserGetter interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error)
}

type SubscribeManager interface {
	GetUserSubscribers(ctx context.Context, subInfo repositories_transfer.GetUserSubscribersInfo) ([]*models.User, uint32, error)
	GetUserSubscriptions(ctx context.Context, subInfo repositories_transfer.GetUserSubscriptionsInfo) ([]*models.User, uint32, error)
	Subscribe(ctx context.Context, subInfo repositories_transfer.SubscribeToUserInfo) error
	Unsubscribe(ctx context.Context, unsubInfo repositories_transfer.UnsubscribeInfo) error
}

type UserUpdater interface {
	UpdateUser(ctx context.Context, updateData repositories_transfer.UpdateUserInfo) (*models.User, error)
}

type UserCreator interface {
	CreateUser(ctx context.Context, createDto *repositories_transfer.CreateUserInfo) (uuid.UUID, error)
}

type UserDeleter interface {
	DeleteUser(ctx context.Context, delInfo repositories_transfer.DeleteUserInfo) error
}
