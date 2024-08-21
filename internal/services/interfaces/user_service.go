package services_interfaces

import (
	"context"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/google/uuid"
)

type UserService interface {
	RegisterUser(ctx context.Context, req *transfer.RegisterInfo) (*transfer.TokenResult, error)
	LoginUser(ctx context.Context, loginInfo *transfer.LoginInfo) (*transfer.TokenResult, error)
	CheckAuth(ctx context.Context, authInfo *transfer.CheckAuthInfo) (*transfer.TokenResult, error)
	GetUserById(ctx context.Context, userId uuid.UUID) (*transfer.GetUserResult, error)
	GetSubscribers(ctx context.Context, getInfo *transfer.GetSubscribersInfo) (*transfer.GetSubscribersResult, error)
	GetSubscriptions(ctx context.Context, getInfo *transfer.GetSubscriptionsInfo) (*transfer.GetSubscriptionsResult, error)
	UpdateUser(ctx context.Context, updateInfo *transfer.UpdateUserInfo) (*transfer.UpdateUserResult, error)
	Subscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error
	Unsubscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error
	DeleteUser(ctx context.Context, deleteInfo *transfer.DeleteUserInfo) error
	UploadAvatar(ctx context.Context, uploadInfo *transfer.UploadAvatarInfo) (*transfer.AvatarResult, error)
	CompensateDeletedUser(ctx context.Context, eventId uuid.UUID) error
}
