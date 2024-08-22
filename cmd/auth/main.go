package main

import (
	"github.com/KBcHMFollower/blog_user_service/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/KBcHMFollower/blog_user_service/internal/app"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
)

func main() {
	cfg := config.MustLoad()

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
}
