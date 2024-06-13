package auth_server

import (
	"context"

	auth_service "github.com/KBcHMFollower/auth-service/internal/services/auth"
	ssov1 "github.com/KBcHMFollower/test_plate_auth_protos/gen/auth"
	"google.golang.org/grpc"
)

type serverApi struct{
	ssov1.UnimplementedAuthServer
	authService *auth_service.AuthService
}


func Register(gRPC *grpc.Server, authService *auth_service.AuthService){
	ssov1.RegisterAuthServer(gRPC, &serverApi{authService: authService})
}

func (s *serverApi) Login(ctx context.Context, req *ssov1.LoginDTO)(*ssov1.LoginRTO, error)  {
	token, err := s.authService.LoginUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &ssov1.LoginRTO{
		Token: token,
	}, nil
}

func (s *serverApi) Register(ctx context.Context, req *ssov1.RegisterDTO)(*ssov1.RegisterRTO, error)  {
	token, err := s.authService.RegisterUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil{
		return  nil, err
	}

	return &ssov1.RegisterRTO{
		Token: token,
	} , nil
}

func (s *serverApi) CheckAuth(ctx context.Context, req *ssov1.CheckAuthDTO)(*ssov1.CheckAuthRTO, error)  {
	token, err := s.authService.CheckAuth(ctx, req.GetToken())
	if err != nil {
		return nil, err
	}

	return &ssov1.CheckAuthRTO{
		Token: token,
	}, nil
}