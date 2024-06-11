package grpcapp

import (
	"log/slog"

	"google.golang.org/grpc"
)

type App struct {
	log *slog.Logger
	grpcServer *grpc.Server
	port int
}