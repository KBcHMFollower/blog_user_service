package interceptors

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/lib/circuid_breaker"
	"google.golang.org/grpc"
)

func CircuitBreakerInterceptor(cb *circuid_breaker.CircuitBreaker) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		resp, err := cb.Do(ctx, func() (interface{}, error) {
			return handler(ctx, req)
		},
		)

		return resp, err
	}
}
