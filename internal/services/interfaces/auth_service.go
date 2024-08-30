package services_interfaces

import (
	"context"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
)

type AuthService interface {
	Register(ctx context.Context, req *transfer.RegisterInfo) (resToken *transfer.TokenResult, resErr error)
	Login(ctx context.Context, loginInfo *transfer.LoginInfo) (*transfer.TokenResult, error)
	CheckAuth(ctx context.Context, authInfo *transfer.CheckAuthInfo) (*transfer.TokenResult, error)
}
