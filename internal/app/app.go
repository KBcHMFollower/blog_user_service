package app

import (
	"github.com/KBcHMFollower/blog_user_service/config"
	"github.com/KBcHMFollower/blog_user_service/database"
	grpcapp "github.com/KBcHMFollower/blog_user_service/internal/app/grpc_app"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	s3client "github.com/KBcHMFollower/blog_user_service/internal/s3"
	auth_service "github.com/KBcHMFollower/blog_user_service/internal/services"
	"log/slog"
)

type App struct {
	GRpcServer *grpcapp.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	op := "App.New"
	appLog := log.With(
		slog.String("op", op))

	driver, db, err := database.New(cfg.Storage.ConnectionString)
	if err != nil {
		log.Error("can`t connect to database", err)
		panic(err)
	}
	appLog.Info("Successfully connected to database")

	if err := database.ForceMigrate(db, cfg.Storage.MigrationPath); err != nil {
		log.Error("can`t migrate database", err)
		panic(err)
	}
	appLog.Info("Successfully migrated database")

	s3Client, err := s3client.New(cfg.Minio.Endpoint, cfg.Minio.AccessKey, cfg.Minio.SecretKey, cfg.Minio.Bucket)
	if err != nil {
		log.Error("can`t create S3 client", err)
		panic(err)
	}
	appLog.Info("Successfully created S3 client")

	userRepository, err := repository.NewUserRepository(driver)
	if err != nil {
		log.Error("can`t create user repository", err)
		panic(err)
	}

	authService := auth_service.New(log, cfg.JWT.TokenTTL, cfg.JWT.TokenSecret, userRepository, s3Client)

	grpcApp := grpcapp.New(log, cfg.GRpc.Port, authService)
	appLog.Info("Successfully created GRPC app")

	return &App{
		GRpcServer: grpcApp,
	}
}
