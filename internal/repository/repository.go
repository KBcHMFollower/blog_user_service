package repository

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type UpdateInfo struct {
	Name  string
	Value string
}

type UpdateData struct {
	Id         uuid.UUID
	UpdateInfo []*UpdateInfo
}

type UserStore interface {
	CreateUser(ctx context.Context, createDto *CreateUserDto) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error)
	GetUserSubscribers(ctx context.Context, userId uuid.UUID, page uint64, size uint64) ([]*models.User, uint32, error)
	GetUserSubscriptions(ctx context.Context, userId uuid.UUID, page uint64, size uint64) ([]*models.User, uint32, error)
	UpdateUser(ctx context.Context, updateData UpdateData) (*models.User, error)
	Subscribe(ctx context.Context, bloggerId uuid.UUID, subscriberId uuid.UUID) error
	Unsubscribe(ctx context.Context, bloggerId uuid.UUID, subscriberId uuid.UUID) error
	DeleteUser(ctx context.Context, userId uuid.UUID) error
}
