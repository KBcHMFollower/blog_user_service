package grpcapp

import (
	"fmt"
	grpcservers "github.com/KBcHMFollower/auth-service/internal/grpc"
	auth_service "github.com/KBcHMFollower/auth-service/internal/services"
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
	userService *auth_service.UserService,
) *App {
	gRpcServer := grpc.NewServer()
	grpcservers.RegisterAuthServer(gRpcServer, userService)
	grpcservers.RegisterUserServer(gRpcServer, userService)

	return &App{
		log:        log,
		port:       port,
		gRpcServer: gRpcServer,
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	log.Info("starting gRpc server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is get up ", slog.Int("port", 1212))

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("stopping gRpc server")

	a.gRpcServer.GracefulStop()
}
