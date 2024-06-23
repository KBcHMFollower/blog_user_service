package grpcservers

import (
	"context"
	authv1 "github.com/KBcHMFollower/auth-service/api/protos/gen/auth"
	services "github.com/KBcHMFollower/auth-service/internal/services"
	"google.golang.org/grpc"
)

type GRPCAuth struct {
	authv1.UnimplementedAuthServer
	authService *services.UserService
}

func RegisterAuthServer(gRPC *grpc.Server, authService *services.UserService) {
	authv1.RegisterAuthServer(gRPC, &GRPCAuth{authService: authService})
}

func (s *GRPCAuth) Login(ctx context.Context, req *authv1.LoginDTO) (*authv1.LoginRTO, error) {
	token, err := s.authService.LoginUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &authv1.LoginRTO{
		Token: token,
	}, nil
}

func (s *GRPCAuth) Register(ctx context.Context, req *authv1.RegisterDTO) (*authv1.RegisterRTO, error) {
	token, err := s.authService.RegisterUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return &authv1.RegisterRTO{
		Token: token,
	}, nil
}

func (s *GRPCAuth) CheckAuth(ctx context.Context, req *authv1.CheckAuthDTO) (*authv1.CheckAuthRTO, error) {
	token, err := s.authService.CheckAuth(ctx, req.GetToken())
	if err != nil {
		return nil, err
	}

	return &authv1.CheckAuthRTO{
		Token: token,
	}, nil
}
