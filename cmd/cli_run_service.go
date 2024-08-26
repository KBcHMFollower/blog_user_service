package main

import (
	"github.com/KBcHMFollower/blog_user_service/internal/app"
	"github.com/KBcHMFollower/blog_user_service/internal/config"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	var (
		argConfigPath string
	)

	runCmd := &cobra.Command{
		Use:   "run-gw",
		Short: "Start server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.MustLoad(argConfigPath)

			log := logger.SetupLogger(cfg.Env)

			webApp := app.New(log, cfg)
			log.Info("сервер запускается!")

			go webApp.Run()

			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

			sign := <-stop

			log.Info("stopping webApp ", slog.String("signal", sign.String()))

			if err := webApp.Stop(); err != nil {
				log.Error("Error in stopping app", err)
			}
		},
	}
	runCmd.Flags().StringVar(&argConfigPath, "cfg", "", "path to config")
	rootCmd.AddCommand(runCmd)
}
