package main

import (
	"github.com/KBcHMFollower/blog_user_service/config"
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

	app := app.New(log, cfg)
	log.Info("сервер запускается!")

	go app.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping app ", slog.String("signal", sign.String()))

	app.Stop()
}
