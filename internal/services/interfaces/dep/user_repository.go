package services_dep_interfaces

import (
	"context"
	"database/sql"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type UserGetter interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, userId uuid.UUID, tx *sql.Tx) (*models.User, error)
	TryGetUserFromCache(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type SubscribeManager interface {
	GetUserSubscribers(ctx context.Context, subInfo repositories_transfer.GetUserSubscribersInfo) ([]*models.User, uint32, error)
	GetUserSubscriptions(ctx context.Context, subInfo repositories_transfer.GetUserSubscriptionsInfo) ([]*models.User, uint32, error)
	Subscribe(ctx context.Context, subInfo repositories_transfer.SubscribeToUserInfo) error
	Unsubscribe(ctx context.Context, unsubInfo repositories_transfer.UnsubscribeInfo) error
}

type UserUpdater interface {
	UpdateUser(ctx context.Context, updateData repositories_transfer.UpdateUserInfo, tx *sql.Tx) error
}

type UserCreator interface {
	CreateUser(ctx context.Context, createDto *repositories_transfer.CreateUserInfo) (uuid.UUID, error)
	SetUserToCache(ctx context.Context, user *models.User) error
	RollBackUser(ctx context.Context, user models.User) error
}

type UserDeleter interface {
	DeleteUser(ctx context.Context, delInfo repositories_transfer.DeleteUserInfo, tx *sql.Tx) error
	DeleteUserFromCache(ctx context.Context, id uuid.UUID) error
}
