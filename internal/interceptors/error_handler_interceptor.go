package interceptors

import (
	"context"
	"errors"
	"github.com/KBcHMFollower/blog_user_service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorsTransformer struct {
	errorMap map[error]error
}

func NewErrorsTransformer() *ErrorsTransformer {
	return &ErrorsTransformer{errorMap: map[error]error{
		domain.ErrBadRequest:   status.Error(codes.InvalidArgument, "bad request"),
		domain.ErrUnauthorized: status.Error(codes.Unauthenticated, "unauthorized"),
		domain.ErrNotFound:     status.Error(codes.NotFound, "not found"),
		domain.ErrConflict:     status.Error(codes.AlreadyExists, "already exists"),
	}}
}

func (et *ErrorsTransformer) GetGrpcError(err error) error {
	errMapKeys := make([]error, 0, len(et.errorMap))
	for err := range et.errorMap {
		errMapKeys = append(errMapKeys, err)
	}

	for _, e := range errMapKeys {
		if errors.Is(err, e) {
			return et.errorMap[e]
		}
	}

	return status.Error(codes.Internal, "internal server error")
}

func ErrorHandlerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		errTransformer := NewErrorsTransformer()

		return resp, errTransformer.GetGrpcError(err)
	}
}
