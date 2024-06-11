package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	auth_server "github.com/KBcHMFollower/auth-service/internal/grpc/auth"
	auth_service "github.com/KBcHMFollower/auth-service/internal/services/auth"
	"google.golang.org/grpc"
)

type App struct {
	log *slog.Logger
	gRpcServer *grpc.Server
	port int
}

func New(
	log *slog.Logger,
	port int,
	authService *auth_service.AuthService,
)(*App){
	gRpcServer:=grpc.NewServer()
	auth_server.Register(gRpcServer, authService)

	return &App{
		log: log,
		port: port,
		gRpcServer: gRpcServer,
	}
}

func (a *App) Run() error{
	const op = "grpcapp.Run"

	log:=a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	log.Info("starting gRpc server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil{
		return fmt.Errorf("%s: %w", op, err)
	}
	

	log.Info("grpc server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(){
	const op = "grpcapp.Stop"

	log:=a.log.With(
		slog.String("op", op),
	)

	log.Info("stopping gRpc server")

	a.gRpcServer.GracefulStop()
}