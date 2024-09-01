package interceptors

import (
	"context"
	servicestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/interceptors/interfaces/dep"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrMissingIdmKey     = status.Error(codes.InvalidArgument, "Missing idm key")
	ErrDuplicatedRequest = status.Error(codes.InvalidArgument, "Duplicated idm key")
	ErrCantParse         = status.Error(codes.InvalidArgument, "Cant parse idm key")
	ErrInternal          = status.Error(codes.Internal, "Internal error")
)

func IdempotencyInterceptor(idmChecker dep.IdempotencyKeysChecker) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrMissingIdmKey
		}

		keys := md["req-id"]
		if len(keys) == 0 {
			return nil, ErrMissingIdmKey
		}

		idmKey, err := uuid.Parse(keys[0])
		if err != nil {
			return nil, ErrCantParse
		}

		exist, err := idmChecker.CheckAndCreate(ctx, servicestransfer.RequestsCheckExistsInfo{
			Key: idmKey,
		})
		if err != nil {
			return nil, ErrInternal
		}
		if exist {
			return nil, ErrDuplicatedRequest
		}

		return handler(ctx, req)
	}
}
