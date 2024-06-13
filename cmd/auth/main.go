package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/KBcHMFollower/auth-service/internal/app"
	"github.com/KBcHMFollower/auth-service/internal/config"
	"github.com/KBcHMFollower/auth-service/internal/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	app := app.New(log, int(cfg.GRPC.Port), cfg.ConnectionString, cfg.MigradionPath, cfg.TokenTTL, cfg.TocenSecret)
	log.Info("сервер запускается!")

	go app.GRpcServer.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign :=<-stop

	log.Info("stopping app ", slog.String("signal", sign.String()))

	app.GRpcServer.Stop()
}