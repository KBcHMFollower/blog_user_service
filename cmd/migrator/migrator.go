package migrator

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

func New(db *sql.DB) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create driver instance: %w", err)
	}

	return &Migrator{
		driver: driver,
	}, nil
}

func (m *Migrator) Migrate(pathToMigrations string, dbName string) error {
	// Ensure pathToMigrations has the correct scheme
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
