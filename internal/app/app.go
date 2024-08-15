package app

import (
	"github.com/KBcHMFollower/blog_user_service/config"
	"github.com/KBcHMFollower/blog_user_service/database"
	"github.com/KBcHMFollower/blog_user_service/internal/amqp_client"
	grpcapp "github.com/KBcHMFollower/blog_user_service/internal/app/grpc_app"
	"github.com/KBcHMFollower/blog_user_service/internal/cashe"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	s3client "github.com/KBcHMFollower/blog_user_service/internal/s3"
	auth_service "github.com/KBcHMFollower/blog_user_service/internal/services"
	"github.com/KBcHMFollower/blog_user_service/internal/workers"
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

	casheStorage, err := cashe.NewRedisCache(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.CacheTTL)
	if err != nil {
		log.Error("can`t connect to redis", err)
		panic(err)
	}
	appLog.Info("Successfully connected to redis")

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

	rabbitmqClient, err := amqp_client.NewRabbitMQClient(cfg.RabbitMq.Addr)
	if err != nil {
		log.Error("can`t create RabbitMQ client", err)
		panic(err)
	}

	eventRepository, err := repository.NewEventRepository(driver)
	if err != nil {
		log.Error("can`t create S3 client", err)
		panic(err)
	}

	userRepository, err := repository.NewUserRepository(driver, casheStorage)
	if err != nil {
		log.Error("can`t create user repository", err)
		panic(err)
	}

	authService := auth_service.New(log, cfg.JWT.TokenTTL, cfg.JWT.TokenSecret, userRepository, s3Client)

	amqpSender := workers.NewEventChecker(rabbitmqClient, eventRepository, log)
	amqpSender.Run()

	grpcApp := grpcapp.New(log, cfg.GRpc.Port, authService)
	appLog.Info("Successfully created GRPC app")

	return &App{
		GRpcServer: grpcApp,
	}
}
