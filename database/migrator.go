package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Ensure the file source driver is imported
	_ "github.com/lib/pq"
)

type Migrator struct {
	driver database.Driver
}

func NewMigrator(db *sql.DB) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create driver instance: %w", err)
	}

	return &Migrator{
		driver: driver,
	}, nil
}

func (m *Migrator) Migrate(pathToMigrations string, dbName string) error {

	if !strings.HasPrefix(pathToMigrations, "file://") {
		pathToMigrations = "file://" + pathToMigrations
	}

	migration, err := migrate.NewWithDatabaseInstance(pathToMigrations, dbName, m.driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not apply migrations: %w", err)
	}

	return nil
}

func ForceMigrate(db *sql.DB, pathToMigrates string) error {
	migrator, err := NewMigrator(db)
	if err != nil {
		return fmt.Errorf("can`t create migrator : %v", err)
	}

	err = migrator.Migrate(pathToMigrates, "postgres")
	if err != nil {
		return fmt.Errorf("can`t migrate : %v", err)
	}

	return nil
}
