package services_dep_interfaces

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type UserGetter interface {
	User(ctx context.Context, condition repositories_transfer.GetUserInfo, tx database.Transaction) (*models.User, error)
	TryGetFromCache(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type UserUpdater interface {
	Update(ctx context.Context, updateData repositories_transfer.UpdateUserInfo, tx database.Transaction) error
}

type UserCreator interface {
	Create(ctx context.Context, createDto *repositories_transfer.CreateUserInfo, tx database.Transaction) (uuid.UUID, error)
	SetToCache(ctx context.Context, user *models.User) error
	RollBackUser(ctx context.Context, user models.User) error
}

type UserDeleter interface {
	Delete(ctx context.Context, delInfo repositories_transfer.DeleteUserInfo, tx database.Transaction) error
	DeleteFromCache(ctx context.Context, id uuid.UUID) error
}
