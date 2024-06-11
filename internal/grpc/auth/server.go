package auth_server

import (
	"context"
	"fmt"
	"os"

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
	panic("sdawsd")
}

func (s *serverApi) Register(ctx context.Context, req *ssov1.RegisterDTO)(*ssov1.RegisterRTO, error)  {

	fmt.Fprint(os.Stdout, "dawsdawsd")

	token, err := s.authService.RegisterUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil{
		return  nil, err
	}

	return &ssov1.RegisterRTO{
		Token: token,
	} , nil
}

func (s *serverApi) CheckAuth(ctx context.Context, req *ssov1.CheckAuthDTO)(*ssov1.CheckAuthRTO, error)  {
	panic("sdawsd")
}