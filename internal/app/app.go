package app

import (
	"github.com/KBcHMFollower/blog_user_service/config"
	"github.com/KBcHMFollower/blog_user_service/database"
	"github.com/KBcHMFollower/blog_user_service/internal/app/amqp_app"
	grpcapp "github.com/KBcHMFollower/blog_user_service/internal/app/grpc_app"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqp/rabbitmqclient"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/cashe"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/s3"
	amqp_handlers "github.com/KBcHMFollower/blog_user_service/internal/handlers/amqp"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	auth_service "github.com/KBcHMFollower/blog_user_service/internal/services"
	"github.com/KBcHMFollower/blog_user_service/internal/workers"
	"log/slog"
)

type App struct {
	gRpcApp *grpcapp.App
	amqpApp *amqp_app.AmqpApp
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	//op := "App.NewUserService"
	//appLog := log.With(
	//	slog.String("op", op))

	driver, db, err := database.New(cfg.Storage.ConnectionString)
	ContinueOrPanic(err)
	err = database.ForceMigrate(db, cfg.Storage.MigrationPath)
	ContinueOrPanic(err)

	cacheStorage, err := cashe.NewRedisCache(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.CacheTTL)
	ContinueOrPanic(err)

	s3Client, err := s3client.New(cfg.Minio.Endpoint, cfg.Minio.AccessKey, cfg.Minio.SecretKey, cfg.Minio.Bucket)
	ContinueOrPanic(err)

	eventRepository, err := repository.NewEventRepository(driver)
	ContinueOrPanic(err)
	userRepository, err := repository.NewUserRepository(driver, cacheStorage)
	ContinueOrPanic(err)

	authService := auth_service.NewUserService(log, cfg.JWT.TokenTTL, cfg.JWT.TokenSecret, userRepository, eventRepository, s3Client)

	amqpUsersHandler := amqp_handlers.NewUserHandler(authService)

	rabbitmqClient, err := rabbitmqclient.NewRabbitMQClient(cfg.RabbitMq.Addr)
	ContinueOrPanic(err)
	rabbitMqApp := amqp_app.NewAmqpApp(rabbitmqClient)
	rabbitMqApp.RegisterHandler("posts-deleted", amqpUsersHandler.HandlePostDeletingEvent)

	amqpSender := workers.NewEventChecker(rabbitmqClient, eventRepository, log)
	amqpSender.Run()

	grpcApp := grpcapp.New(log, cfg.GRpc.Port, authService)

	return &App{
		gRpcApp: grpcApp,
		amqpApp: rabbitMqApp,
	}
}

func (a *App) Run() {
	err := a.gRpcApp.Run()
	ContinueOrPanic(err)

	err = a.amqpApp.Start()
	ContinueOrPanic(err)
}

func (a *App) Stop() error {
	a.gRpcApp.Stop()
	//a.amqpApp.Stop()
	return nil
} //TODO

func ContinueOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
