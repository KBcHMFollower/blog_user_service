package interceptors

import (
	"context"
	"errors"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorsTransformer struct {
	errorMap map[error]error
}

func NewErrorsTransformer() *ErrorsTransformer {
	return &ErrorsTransformer{errorMap: map[error]error{
		ctxerrors.ErrBadRequest:   status.Error(codes.InvalidArgument, "bad request"),
		ctxerrors.ErrUnauthorized: status.Error(codes.Unauthenticated, "unauthorized"),
		ctxerrors.ErrNotFound:     status.Error(codes.NotFound, "not found"),
		ctxerrors.ErrConflict:     status.Error(codes.AlreadyExists, "already exists"),
	}}
}

func (et *ErrorsTransformer) GetGrpcError(err error) error {
	if err == nil {
		return nil
	}

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
