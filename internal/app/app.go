package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/KBcHMFollower/auth-service/internal/app/grpcApp"
	"github.com/KBcHMFollower/auth-service/internal/repository"
	auth_service "github.com/KBcHMFollower/auth-service/internal/services/auth"
)

type App struct {
	GRpcServer *grpcapp.App
}

func New(
	log *slog.Logger,
	port int,
	connectionString string,
	migrationPath string,
	tokenTTL time.Duration,
	tokenSecret string,
) (*App){

	repository, err := repository.NewPostgressStore(connectionString)
	if err != nil{
		log.Error("Оштбка подключения к базе данных")
		panic(err)
	}

	migrateErr := repository.Migrate(migrationPath)
	if migrateErr != nil {
		log.Error("Оштбка миграции")
		panic(migrateErr)
	}

	authService := auth_service.New(log, tokenTTL,tokenSecret, repository, repository)

	grpcApp := grpcapp.New(log, port, authService)

	return &App{
		GRpcServer: grpcApp,
	}
} 

