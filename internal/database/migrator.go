package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Ensure the file source driver is imported
	_ "github.com/lib/pq"
)

type MigrateType string
type DbName string

const (
	MigrateUp   MigrateType = "up"
	MigrateDown MigrateType = "down"
)

const (
	Postgres = "postgres"
)

func ForceMigrate(db *sql.DB, pathToMigrates string, migrateType MigrateType, dbName DbName) error {
	migrator, err := newMigrator(db)
	if err != nil {
		return fmt.Errorf("can`t create migrator : %v", err)
	}

	err = migrator.migrate(pathToMigrates, dbName, migrateType)
	if err != nil {
		return fmt.Errorf("can`t migrate : %v", err)
	}

	return nil
}

type Migrator struct {
	driver       database.Driver
	migrateTypes map[MigrateType]func(migrate *migrate.Migrate) error
}

func newMigrator(db *sql.DB) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create driver instance: %w", err)
	}

	return &Migrator{
		driver: driver,
		migrateTypes: map[MigrateType]func(migrate *migrate.Migrate) error{
			MigrateUp:   migrateUp,
			MigrateDown: migrateDown,
		},
	}, nil
}

func (m *Migrator) migrate(pathToMigrations string, dbName DbName, migrateType MigrateType) (resErr error) {

	if !strings.HasPrefix(pathToMigrations, "file://") {
		pathToMigrations = "file://" + pathToMigrations
	}

	migration, err := migrate.NewWithDatabaseInstance(pathToMigrations, string(dbName), m.driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer func() {
		if sErr, dbErr := migration.Close(); dbErr != nil || sErr != nil {
			resErr = errors.Join(resErr, sErr, dbErr)
		}
	}()

	err = m.migrateTypes[migrateType](migration)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not apply migrations: %w", err)
	}

	return nil
}

func migrateUp(migrate *migrate.Migrate) error {
	return migrate.Up()
}

func migrateDown(migrate *migrate.Migrate) error {
	return migrate.Down()
}
