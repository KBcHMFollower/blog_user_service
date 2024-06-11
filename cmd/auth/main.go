package main

import (
	"authService/internal/config"
	"authService/internal/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("сервер запускается!")
}