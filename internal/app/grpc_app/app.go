package grpcapp

import (
	"fmt"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	handlersdep "github.com/KBcHMFollower/blog_user_service/internal/handlers/dep"
	grpcservers2 "github.com/KBcHMFollower/blog_user_service/internal/handlers/grpc"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	servicesinterfaces "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	log        logger.Logger
	gRpcServer *grpc.Server
	port       int
}

func New(
	log logger.Logger,
	port int,
	userService servicesinterfaces.UserService,
	authService servicesinterfaces.AuthService,
	subsService servicesinterfaces.SubsService,
	validator handlersdep.Validator,
	interceptor grpc.ServerOption,
) *App {
	gRpcServer := grpc.NewServer(interceptor)

	grpcservers2.RegisterAuthServer(gRpcServer, authService, validator, log)
	grpcservers2.RegisterUserServer(gRpcServer, userService, subsService, log, validator)

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
