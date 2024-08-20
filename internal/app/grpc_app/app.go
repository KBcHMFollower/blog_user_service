package grpcapp

import (
	"fmt"
	grpcservers2 "github.com/KBcHMFollower/blog_user_service/internal/handlers/grpc"
	auth_service "github.com/KBcHMFollower/blog_user_service/internal/services"
	"google.golang.org/grpc/peer"
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
	gRpcServer := grpc.NewServer(grpc.StreamInterceptor(logStreamInterceptor))
	grpcservers2.RegisterAuthServer(gRpcServer, userService)
	grpcservers2.RegisterUserServer(gRpcServer, userService)

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

func logStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	p, ok := peer.FromContext(ss.Context())
	if !ok {
		fmt.Printf("Failed to get peer information from context")
	} else {
		fmt.Printf("Client connected from %s", p.Addr)
	}
	return handler(srv, ss)
}
