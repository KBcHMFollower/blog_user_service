package store_app

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/cashe"
	s3client "github.com/KBcHMFollower/blog_user_service/internal/clients/s3"
	"github.com/KBcHMFollower/blog_user_service/internal/config"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	_ "github.com/lib/pq"
	"log"
)

type PostgresStore struct {
	Store         database.DBWrapper
	migrationPath string
	db            *sql.DB
} //TODO: ПЕРЕПИСАТЬ МИГРАТОР, ЭТО НЕ НОРМ

type StoreApp struct {
	PostgresStore *PostgresStore
	RedisStore    cashe.CasheStorage
	S3Client      s3client.S3Client
}

func New(postgresConnectionInfo config.Storage, redisConnectionInfo config.Redis, minioConnectInfo config.Minio) (*StoreApp, error) {
	db, err := sql.Open("postgres", postgresConnectionInfo.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("error in process db connection : %w", err)
	}

	cacheStorage, err := cashe.NewRedisCache(redisConnectionInfo.Addr, redisConnectionInfo.Password, redisConnectionInfo.DB, redisConnectionInfo.CacheTTL)
	if err != nil {
		return nil, fmt.Errorf("error in process db connection : %w", err)
	}

	minioClient, err := s3client.NewMinioClient(minioConnectInfo.Endpoint, minioConnectInfo.AccessKey, minioConnectInfo.SecretKey, minioConnectInfo.Bucket)
	if err != nil {
		return nil, fmt.Errorf("error in process db connection : %w", err)
	}

	return &StoreApp{
		PostgresStore: &PostgresStore{
			Store:         &database.DBDriver{db},
			migrationPath: postgresConnectionInfo.MigrationPath,
			db:            db,
		},
		RedisStore: cacheStorage,
		S3Client:   minioClient,
	}, nil
}

func (app *StoreApp) Run() error {
	if err := database.ForceMigrate(
		app.PostgresStore.db,
		app.PostgresStore.migrationPath,
		database.MigrateUp,
		database.Postgres,
	); err != nil {
		log.Fatalf("error in process db connection : %v", err)
		return err
	}

	return nil
}

func (app *StoreApp) Stop() error {
	var resErr error = nil

	if err := app.PostgresStore.Store.Stop(); err != nil {
		resErr = errors.Join(resErr, fmt.Errorf("error in stop postgres store : %w", err))
	}

	if err := app.RedisStore.Stop(); err != nil {
		resErr = errors.Join(resErr, fmt.Errorf("error in stop redis store : %w", err))
	}

	if err := app.S3Client.Stop(); err != nil {
		resErr = errors.Join(resErr, fmt.Errorf("error in stop s3 client : %w", err))
	}

	return resErr
}
