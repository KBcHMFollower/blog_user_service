package app

import (
	"errors"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/app/amqp_app"
	grpcapp "github.com/KBcHMFollower/blog_user_service/internal/app/grpc_app"
	"github.com/KBcHMFollower/blog_user_service/internal/app/store_app"
	"github.com/KBcHMFollower/blog_user_service/internal/app/workers_app"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient"
	"github.com/KBcHMFollower/blog_user_service/internal/config"
	amqphandlers "github.com/KBcHMFollower/blog_user_service/internal/handlers/amqp"
	"github.com/KBcHMFollower/blog_user_service/internal/interceptors"
	"github.com/KBcHMFollower/blog_user_service/internal/lib"
	"github.com/KBcHMFollower/blog_user_service/internal/lib/validators"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	authservice "github.com/KBcHMFollower/blog_user_service/internal/services"
	"github.com/KBcHMFollower/blog_user_service/internal/workers"
	"google.golang.org/grpc"
)

type App struct {
	gRpcApp    *grpcapp.App
	amqpApp    *amqp_app.AmqpApp
	storeApp   *store_app.StoreApp
	workersApp *workers_app.WorkersApp
	log        logger.Logger
	cfg        *config.Config
}

func New(
	cfg *config.Config,
	log logger.Logger,
) *App {
	storageApp, err := store_app.New(cfg.Storage, cfg.Redis, cfg.Minio)
	lib.ContinueOrPanic(err)
	rabbitMqApp, err := amqp_app.NewAmqpApp(cfg.RabbitMq, log)
	lib.ContinueOrPanic(err)

	vldor, err := validators.NewValidator()
	lib.ContinueOrPanic(err)

	eventRepository := repository.NewEventRepository(storageApp.PostgresStore.Store)
	subsRepository := repository.NewSubscriberRepository(storageApp.PostgresStore.Store)
	userRepository := repository.NewUserRepository(storageApp.PostgresStore.Store, storageApp.RedisStore)
	reqRepository := repository.NewRequestsRepository(storageApp.PostgresStore.Store)

	userService := authservice.NewUserService(
		log,
		storageApp.PostgresStore.Store,
		userRepository,
		eventRepository,
		storageApp.S3Client,
	)
	authService := authservice.NewAuthService(
		userRepository,
		log,
		cfg.JWT.TokenTTL,
		cfg.JWT.TokenSecret,
		storageApp.PostgresStore.Store,
	)
	reqService := authservice.NewRequestsService(reqRepository, log)
	subsService := authservice.NewSubscribersService(
		subsRepository,
		userRepository,
		storageApp.PostgresStore.Store,
		log,
	)
	messService := authservice.NewMessagesService(eventRepository, log)

	amqpUsersHandler := amqphandlers.NewUserHandler(userService, messService, log)

	rabbitMqApp.RegisterHandler(amqpclient.PostsDeletedEventKey, amqpUsersHandler.HandlePostDeletingEvent)

	interceptorsChain := grpc.ChainUnaryInterceptor(
		interceptors.ReqLoggingInterceptor(log),
		interceptors.IdempotencyInterceptor(reqService),
		interceptors.ErrorHandlerInterceptor(),
	)

	workersApp := workers_app.New()
	grpcApp := grpcapp.New(
		log,
		cfg.GRpc.Port,
		userService,
		authService,
		subsService,
		vldor,
		interceptorsChain,
	)

	workersApp.AddWorker(workers.NewEventChecker(rabbitMqApp.Client, eventRepository, log, storageApp.PostgresStore.Store))

	return &App{
		gRpcApp:    grpcApp,
		amqpApp:    rabbitMqApp,
		storeApp:   storageApp,
		workersApp: workersApp,
		log:        log,
	}
}

func (a *App) Run() {
	err := a.storeApp.Run()
	lib.ContinueOrPanic(err)

	err = a.amqpApp.Start()
	lib.ContinueOrPanic(err)

	err = a.workersApp.Run()
	lib.ContinueOrPanic(err)

	err = a.gRpcApp.Run()
	lib.ContinueOrPanic(err)
}

func (a *App) Stop() error {
	var resErr error = nil

	a.gRpcApp.Stop()

	if err := a.storeApp.Stop(); err != nil {
		resErr = errors.Join(resErr, fmt.Errorf("error in store stopping proccess: %w", err))
	}

	if err := a.amqpApp.Stop(); err != nil {
		resErr = errors.Join(resErr, fmt.Errorf("error in amqp stopping proccess: %w", err))
	}

	a.workersApp.Stop()

	return resErr
}
