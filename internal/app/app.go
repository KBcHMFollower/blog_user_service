package app

import (
	"github.com/KBcHMFollower/auth-service/config"
	"github.com/KBcHMFollower/auth-service/database"
	grpcapp "github.com/KBcHMFollower/auth-service/internal/app/grpc_app"
	"github.com/KBcHMFollower/auth-service/internal/repository"
	auth_service "github.com/KBcHMFollower/auth-service/internal/services"
	"log/slog"
)

type App struct {
	GRpcServer *grpcapp.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {

	driver, db, err := database.New(cfg.Storage.ConnectionString)
	if err != nil {
		log.Error("can`t connect to database", err)
		panic(err)
	}

	if err := database.ForceMigrate(db, cfg.Storage.MigrationPath); err != nil {
		log.Error("can`t migrate database", err)
		panic(err)
	}

	userRepository, err := repository.NewUserRepository(driver)
	if err != nil {
		log.Error("can`t create user repository", err)
		panic(err)
	}

	authService := auth_service.New(log, cfg.JWT.TokenTTL, cfg.JWT.TokenSecret, userRepository, userRepository)

	grpcApp := grpcapp.New(log, cfg.GRpc.Port, authService)

	return &App{
		GRpcServer: grpcApp,
	}
}
