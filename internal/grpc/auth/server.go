package auth_server

import (
	"context"
	ssov1 "github.com/KBcHMFollower/auth-service/api/protos/gen/auth"
	services "github.com/KBcHMFollower/auth-service/internal/services"

	"google.golang.org/grpc"
)

type serverApi struct {
	ssov1.UnimplementedAuthServer
	authService *services.UserService
}

func Register(gRPC *grpc.Server, authService *services.UserService) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{authService: authService})
}

func (s *serverApi) Login(ctx context.Context, req *ssov1.LoginDTO) (*ssov1.LoginRTO, error) {
	token, err := s.authService.LoginUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &ssov1.LoginRTO{
		Token: token,
	}, nil
}

func (s *serverApi) Register(ctx context.Context, req *ssov1.RegisterDTO) (*ssov1.RegisterRTO, error) {
	token, err := s.authService.RegisterUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return &ssov1.RegisterRTO{
		Token: token,
	}, nil
}

func (s *serverApi) CheckAuth(ctx context.Context, req *ssov1.CheckAuthDTO) (*ssov1.CheckAuthRTO, error) {
	token, err := s.authService.CheckAuth(ctx, req.GetToken())
	if err != nil {
		return nil, err
	}

	return &ssov1.CheckAuthRTO{
		Token: token,
	}, nil
}
