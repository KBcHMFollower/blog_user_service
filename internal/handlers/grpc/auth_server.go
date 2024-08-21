package grpcservers

import (
	"context"
	authv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/auth"
	services_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	services_interfaces "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces"
	"google.golang.org/grpc"
)

type GRPCAuth struct {
	authv1.UnimplementedAuthServer
	authService services_interfaces.UserService
}

func RegisterAuthServer(gRPC *grpc.Server, authService services_interfaces.UserService) {
	authv1.RegisterAuthServer(gRPC, &GRPCAuth{authService: authService})
}

func (s *GRPCAuth) Login(ctx context.Context, req *authv1.LoginDTO) (*authv1.LoginRTO, error) {
	token, err := s.authService.LoginUser(ctx, &services_transfer.LoginInfo{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &authv1.LoginRTO{
		Token: token.AccessToken,
	}, nil
}

func (s *GRPCAuth) Register(ctx context.Context, req *authv1.RegisterDTO) (*authv1.RegisterRTO, error) {
	token, err := s.authService.RegisterUser(ctx, &services_transfer.RegisterInfo{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &authv1.RegisterRTO{
		Token: token.AccessToken,
	}, nil
}

func (s *GRPCAuth) CheckAuth(ctx context.Context, req *authv1.CheckAuthDTO) (*authv1.CheckAuthRTO, error) {
	token, err := s.authService.CheckAuth(ctx, &services_transfer.CheckAuthInfo{
		AccessToken: req.Token,
	})
	if err != nil {
		return nil, err
	}

	return &authv1.CheckAuthRTO{
		Token: token.AccessToken,
	}, nil
}
