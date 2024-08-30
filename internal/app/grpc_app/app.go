package grpcapp

import (
	"fmt"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	grpcservers2 "github.com/KBcHMFollower/blog_user_service/internal/handlers/grpc"
	services_interfaces "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRpcServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
	userService services_interfaces.UserService,
	authService services_interfaces.AuthService,
	subsService services_interfaces.SubsService,
	interceptor grpc.ServerOption,
) *App {
	gRpcServer := grpc.NewServer(interceptor)

	grpcservers2.RegisterAuthServer(gRpcServer, authService, log)
	grpcservers2.RegisterUserServer(gRpcServer, userService, subsService, log)

	return &App{
		log:        log,
		port:       port,
		gRpcServer: gRpcServer,
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return ctxerrors.Wrap("can`t start listen", err)
	}

	if err := a.gRpcServer.Serve(l); err != nil {
		return ctxerrors.Wrap("error in get up grpc", err)
	}

	return nil
}

func (a *App) Stop() {
	a.gRpcServer.GracefulStop()
}
