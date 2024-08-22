package app

import (
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/app/amqp_app"
	grpcapp "github.com/KBcHMFollower/blog_user_service/internal/app/grpc_app"
	"github.com/KBcHMFollower/blog_user_service/internal/app/store_app"
	"github.com/KBcHMFollower/blog_user_service/internal/app/workers_app"
	"github.com/KBcHMFollower/blog_user_service/internal/config"
	amqp_handlers "github.com/KBcHMFollower/blog_user_service/internal/handlers/amqp"
	"github.com/KBcHMFollower/blog_user_service/internal/lib"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	auth_service "github.com/KBcHMFollower/blog_user_service/internal/services"
	"github.com/KBcHMFollower/blog_user_service/internal/workers"
	"log/slog"
)

type App struct {
	gRpcApp    *grpcapp.App
	amqpApp    *amqp_app.AmqpApp
	storeApp   *store_app.StoreApp
	workersApp *workers_app.WorkersApp
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	//op := "App.NewUserService"
	//appLog := log.With(
	//	slog.String("op", op))

	storageApp, err := store_app.New(cfg.Storage, cfg.Redis, cfg.Minio)
	lib.ContinueOrPanic(err)
	rabbitMqApp, err := amqp_app.NewAmqpApp(cfg.RabbitMq)
	lib.ContinueOrPanic(err)

	eventRepository, err := repository.NewEventRepository(storageApp.PostgresStore.Store)
	lib.ContinueOrPanic(err)
	userRepository, err := repository.NewUserRepository(storageApp.PostgresStore.Store, storageApp.RedisStore)
	lib.ContinueOrPanic(err)

	authService := auth_service.NewUserService(log, cfg.JWT.TokenTTL, cfg.JWT.TokenSecret, userRepository, eventRepository, storageApp.S3Client)

	amqpUsersHandler := amqp_handlers.NewUserHandler(authService)

	rabbitMqApp.RegisterHandler("posts-deleted", amqpUsersHandler.HandlePostDeletingEvent)

	workersApp := workers_app.New()
	grpcApp := grpcapp.New(log, cfg.GRpc.Port, authService)

	workersApp.AddWorker(workers.NewEventChecker(rabbitMqApp.Client, eventRepository, log))

	return &App{
		gRpcApp:    grpcApp,
		amqpApp:    rabbitMqApp,
		storeApp:   storageApp,
		workersApp: workersApp,
	}
}

func (a *App) Run() {
	err := a.storeApp.Run()
	lib.ContinueOrPanic(err)

	err = a.gRpcApp.Run()
	lib.ContinueOrPanic(err)

	err = a.amqpApp.Start()
	lib.ContinueOrPanic(err)

	err = a.workersApp.Run()
	lib.ContinueOrPanic(err)
}

func (a *App) Stop() error {
	a.gRpcApp.Stop()

	if err := a.storeApp.Stop(); err != nil {
		return fmt.Errorf("error in store stopping proccess: %w", err)
	}

	if err := a.amqpApp.Stop(); err != nil {
		return fmt.Errorf("error in amqp stopping proccess: %w", err)
	}

	a.workersApp.Stop()

	return nil
} //TODO: ФУНКЦИЯ НЕ ДОЛЖНА ЗАВЕРШАТЬСЯ ПОСЛЕ ПЕРВОЙ ОШИБКИ
