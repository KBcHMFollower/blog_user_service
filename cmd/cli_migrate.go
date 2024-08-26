package main

import (
	"database/sql"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/config"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

func init() {
	var (
		argMigrateType string
		argDbName      string
		argConfigPath  string
	)

	runCmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.MustLoad(argConfigPath)
			log := logger.SetupLogger(cfg.Env)

			log.Info(fmt.Sprintf("starting migration with : type-%s, db-%s", argMigrateType, argDbName))

			db, err := sql.Open(argDbName, cfg.Storage.ConnectionString)
			if err != nil {
				panic(fmt.Errorf("can`t connect to postgres: %v", err))
			}

			if err := database.ForceMigrate(
				db,
				cfg.Storage.MigrationPath,
				database.MigrateType(argMigrateType),
				database.DbName(argDbName),
			); err != nil {
				panic(fmt.Errorf("can`t migrate: %v", err))
			}

			log.Info("Migration complete")
		},
	}
	runCmd.Flags().StringVar(&argMigrateType, "mt", "up", "migrate type(up or down)")
	runCmd.Flags().StringVar(&argDbName, "db", "postgres", "database name")
	runCmd.Flags().StringVar(&argConfigPath, "cfg", "", "config path")
	rootCmd.AddCommand(runCmd)
}
