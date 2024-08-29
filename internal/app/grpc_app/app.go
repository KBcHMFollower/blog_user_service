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
	interceptor grpc.ServerOption,
) *App {
	gRpcServer := grpc.NewServer(
		interceptor)
	grpcservers2.RegisterAuthServer(gRpcServer, userService, log)
	grpcservers2.RegisterUserServer(gRpcServer, userService, log)

	return &App{
		log:        log,
		port:       port,
		gRpcServer: gRpcServer,
	}
}

// todo
func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}

	if err := a.gRpcServer.Serve(l); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() {
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
