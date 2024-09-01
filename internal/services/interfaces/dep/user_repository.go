package services_dep_interfaces

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type UserGetter interface {
	Users(ctx context.Context, info *repositoriestransfer.GetUsersInfo, tx database.Transaction) ([]*models.User, error)
	User(ctx context.Context, condition repositoriestransfer.GetUserInfo, tx database.Transaction) (*models.User, error)
	Count(ctx context.Context, condition repositoriestransfer.GetUsersCountInfo, tx database.Transaction) (int64, error)
	TryGetFromCache(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type UserUpdater interface {
	Update(ctx context.Context, updateData repositoriestransfer.UpdateUserInfo, tx database.Transaction) error
}

type UserCreator interface {
	Create(ctx context.Context, createDto *repositoriestransfer.CreateUserInfo, tx database.Transaction) (uuid.UUID, error)
	SetToCache(ctx context.Context, user *models.User) error
	RollBackUser(ctx context.Context, user models.User) error
}

type UserDeleter interface {
	Delete(ctx context.Context, delInfo repositoriestransfer.DeleteUserInfo, tx database.Transaction) error
	DeleteFromCache(ctx context.Context, id uuid.UUID) error
}
