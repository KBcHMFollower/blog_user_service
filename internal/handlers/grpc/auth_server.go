package grpcservers

import (
	"context"
	authv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/auth"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	servicestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	handlersdep "github.com/KBcHMFollower/blog_user_service/internal/handlers/dep"
	handlersutils "github.com/KBcHMFollower/blog_user_service/internal/handlers/lib"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	servicesinterfaces "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces"
	"google.golang.org/grpc"
)

type GRPCAuth struct {
	authv1.UnimplementedAuthServer
	authService servicesinterfaces.AuthService
	log         logger.Logger
	validator   handlersdep.Validator
}

func RegisterAuthServer(gRPC *grpc.Server, authService servicesinterfaces.AuthService, validator handlersdep.Validator, log logger.Logger) {
	authv1.RegisterAuthServer(gRPC, &GRPCAuth{authService: authService, log: log, validator: validator})
}

func (s *GRPCAuth) Login(ctx context.Context, req *authv1.LoginDTO) (*authv1.LoginRTO, error) {
	logInfo := servicestransfer.LoginInfo{
		Email:    req.Email,
		Password: req.Password,
	}

	if err := s.validator.Struct(logInfo); err != nil {
		s.log.DebugContext(ctxerrors.ErrorCtx(ctx, err), err.Error())
		return nil, handlersutils.ReturnValidationError(err)
	}

	token, err := s.authService.Login(ctx, &logInfo)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "can`t login", logger.ErrKey, err.Error())
		return nil, err
	}

	return &authv1.LoginRTO{
		Token: token.AccessToken,
	}, nil
}

func (s *GRPCAuth) Register(ctx context.Context, req *authv1.RegisterDTO) (*authv1.RegisterRTO, error) {
	regInfo := servicestransfer.RegisterInfo{
		Email:    req.Email,
		Password: req.Password,
		FName:    req.Fname,
		LName:    req.Lname,
	}

	if err := s.validator.Struct(&regInfo); err != nil {
		s.log.DebugContext(ctxerrors.ErrorCtx(ctx, err), "validation err", logger.ErrKey, err.Error())
		return nil, handlersutils.ReturnValidationError(err)
	}

	token, err := s.authService.Register(ctx, &regInfo)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "can`t register", logger.ErrKey, err.Error())
		return nil, err
	}

	return &authv1.RegisterRTO{
		Token: token.AccessToken,
	}, nil
}

func (s *GRPCAuth) CheckAuth(ctx context.Context, req *authv1.CheckAuthDTO) (*authv1.CheckAuthRTO, error) {
	checkAuthInfo := servicestransfer.CheckAuthInfo{
		AccessToken: req.Token,
	}

	if err := s.validator.Struct(checkAuthInfo); err != nil {
		return nil, handlersutils.ReturnValidationError(err)
	}

	token, err := s.authService.CheckAuth(ctx, &checkAuthInfo)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "can`t check", logger.ErrKey, err.Error())
		return nil, err
	}

	return &authv1.CheckAuthRTO{
		Token: token.AccessToken,
	}, nil
}
