package services_interfaces

import (
	"context"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/google/uuid"
)

type UserService interface {
	GetUserById(ctx context.Context, userId uuid.UUID) (*transfer.GetUserResult, error)
	UpdateUser(ctx context.Context, updateInfo *transfer.UpdateUserInfo) (*transfer.UpdateUserResult, error)
	DeleteUser(ctx context.Context, deleteInfo *transfer.DeleteUserInfo) error
	UploadAvatar(ctx context.Context, uploadInfo *transfer.UploadAvatarInfo) (*transfer.AvatarResult, error)
	CompensateDeletedUser(ctx context.Context, eventId uuid.UUID) error
}
