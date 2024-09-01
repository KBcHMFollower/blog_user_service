package interceptors

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

const (
	methodLogKey = "method"
)

func ReqLoggingInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.InfoContext(ctx, "No metadata from incoming context")
		}

		reqId := getInfoFromMd(md, "req-id")
		userId := getInfoFromMd(md, "user-id")

		ctx = logger.UpdateLoggerCtx(ctx, logger.ReqIdKey, reqId)
		ctx = logger.UpdateLoggerCtx(ctx, logger.ReqUserKey, userId)
		ctx = logger.UpdateLoggerCtx(ctx, methodLogKey, info.FullMethod)

		log.InfoContext(ctx, "--Method starting execution--", "data", req)

		startTime := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)

		log.InfoContext(ctx, "--Method is executed--", "duration", duration, "err", err)

		return resp, err
	}
}

func getInfoFromMd(md metadata.MD, k string) string {
	v, ok := md[k]

	switch ok {
	case true:
		return v[0]
	default:
		return "undefined"
	}
}
