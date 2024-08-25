package interceptors

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"time"
)

func ReqLoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Info("No metadata from incoming context")
		}

		var reqId string
		ids, ok := md["req-id"]

		switch ok {
		case true:
			reqId = ids[0]
		default:
			reqId = "unknown"
		}

		log.Info(fmt.Sprintf("ReqID: %s; Method: %s; Starting execution", reqId, info.FullMethod))

		startTime := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)

		log.Info(fmt.Sprintf("ReqID: %s; Method: %s; Duration: %s; Error: %v", reqId, info.FullMethod, duration, err))

		return resp, err
	}
}
