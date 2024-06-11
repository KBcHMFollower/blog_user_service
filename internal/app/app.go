package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/KBcHMFollower/auth-service/internal/app/grpcApp"
	auth_service "github.com/KBcHMFollower/auth-service/internal/services/auth"
)

type App struct {
	GRpcServer *grpcapp.App
}

func New(
	log *slog.Logger,
	port int,
	storagePath string,
	tokenTTL time.Duration,
	tokenSecret string,
) (*App){

	authService := auth_service.New(log, tokenTTL,tokenSecret)

	grpcApp := grpcapp.New(log, port, authService)

	return &App{
		GRpcServer: grpcApp,
	}
} 

