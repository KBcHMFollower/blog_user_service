package grpcservers

import (
	"context"
	usersv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/users"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	servicestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	handlersdep "github.com/KBcHMFollower/blog_user_service/internal/handlers/dep"
	handlersutils "github.com/KBcHMFollower/blog_user_service/internal/handlers/lib"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	servicesinterfaces "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type GRPCUsers struct {
	usersv1.UnimplementedUsersServiceServer
	userService servicesinterfaces.UserService
	subsService servicesinterfaces.SubsService
	log         logger.Logger
	validator   handlersdep.Validator
}

func RegisterUserServer(
	gRPC *grpc.Server,
	userService servicesinterfaces.UserService,
	subsService servicesinterfaces.SubsService,
	log logger.Logger,
	validator handlersdep.Validator,
) {
	usersv1.RegisterUsersServiceServer(gRPC, &GRPCUsers{
		userService: userService,
		subsService: subsService,
		log:         log,
		validator:   validator,
	})
}

func (s *GRPCUsers) GetUser(ctx context.Context, req *usersv1.GetUserDTO) (*usersv1.GetUserRDO, error) {
	userUuid, err := uuid.Parse(req.Id)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse user uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	user, err := s.userService.GetUserById(ctx, userUuid)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to get user", logger.ErrKey, err.Error())
		return nil, err
	}

	return &usersv1.GetUserRDO{
		User: servicestransfer.ConvertUserResToProto(&user.User),
	}, nil
}

func (s *GRPCUsers) GetSubscribers(ctx context.Context, req *usersv1.GetSubscribersDTO) (*usersv1.GetSubscribersRDO, error) {
	bloggerId, err := uuid.Parse(req.BloggerId)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse blogger uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	getSubsInfo := servicestransfer.GetSubscribersInfo{
		BloggerId: bloggerId,
		Page:      req.Page,
		Size:      req.Size,
	}

	if err := s.validator.Struct(getSubsInfo); err != nil {
		s.log.WarnContext(ctxerrors.ErrorCtx(ctx, err), err.Error())
		return nil, handlersutils.ReturnValidationError(err)
	}

	subscribers, err := s.subsService.GetSubscribers(ctx, &getSubsInfo)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to get subscribers", logger.ErrKey, err.Error())
		return nil, err
	}

	return &usersv1.GetSubscribersRDO{
		Subscribers: servicestransfer.ConvertSubscribersToProto(subscribers.Subscribers),
		TotalCount:  subscribers.TotalCount,
	}, nil
}

func (s *GRPCUsers) GetSubscriptions(ctx context.Context, req *usersv1.GetSubscriptionsDTO) (*usersv1.GetSubscriptionsRDO, error) {
	subscriberId, err := uuid.Parse(req.SubscriberId)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse subscriber uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	getSubsInfo := servicestransfer.GetSubscriptionsInfo{
		SubscriberId: subscriberId,
		Page:         req.Page,
		Size:         req.Size,
	}

	if err := s.validator.Struct(getSubsInfo); err != nil {
		return nil, handlersutils.ReturnValidationError(err)
	}

	subscriptions, err := s.subsService.GetSubscriptions(ctx, &getSubsInfo)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to get subscriptions", logger.ErrKey, err.Error())
		return nil, err
	}

	return &usersv1.GetSubscriptionsRDO{
		Subscriptions: servicestransfer.ConvertSubscribersToProto(subscriptions.Subscriptions),
		TotalCount:    subscriptions.TotalCount,
	}, nil
}

func (s *GRPCUsers) UpdateUser(ctx context.Context, req *usersv1.UpdateUserDTO) (*usersv1.UpdateUserRDO, error) {
	userId, err := uuid.Parse(req.Id)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse user uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	updateFields := make(map[servicestransfer.UserFieldTarget]any)
	for k, v := range req.UpdateData {
		updateFields[servicestransfer.UserFieldTarget(k)] = v
	}

	updateInfo := servicestransfer.UpdateUserInfo{
		Id:           userId,
		UpdateFields: updateFields,
	}

	if err := s.validator.Struct(updateInfo); err != nil {
		return nil, handlersutils.ReturnValidationError(err)
	}

	user, err := s.userService.UpdateUser(ctx, &updateInfo)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to update user", logger.ErrKey, err.Error())
		return nil, err
	}

	return &usersv1.UpdateUserRDO{
		User: servicestransfer.ConvertUserResToProto(&user.User),
	}, nil
}

func (s *GRPCUsers) DeleteUser(ctx context.Context, req *usersv1.DeleteUserDTO) (*usersv1.DeleteUserRDO, error) {
	userId, err := uuid.Parse(req.Id)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse user uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	deleteInfo := servicestransfer.DeleteUserInfo{
		Id: userId,
	}

	if err := s.validator.Struct(deleteInfo); err != nil {
		return nil, handlersutils.ReturnValidationError(err)
	}

	if err := s.userService.DeleteUser(ctx, &deleteInfo); err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to delete user", logger.ErrKey, err.Error())
		return &usersv1.DeleteUserRDO{
			IsDeleted: false,
		}, err
	}

	return &usersv1.DeleteUserRDO{
		IsDeleted: true,
	}, nil
}

func (s *GRPCUsers) Subscribe(ctx context.Context, req *usersv1.SubscribeDTO) (*usersv1.SubscribeRDO, error) {
	bloggerId, err := uuid.Parse(req.BloggerId)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse blogger uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	subscriberId, err := uuid.Parse(req.SubscriberId)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse subscriber uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	if err := s.subsService.Subscribe(ctx, &servicestransfer.SubscribeInfo{
		SubscriberId: subscriberId,
		BloggerId:    bloggerId,
	}); err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to subscribe to blogger", logger.ErrKey, err.Error())
		return &usersv1.SubscribeRDO{
			IsSubscribe: false,
		}, err
	}

	return &usersv1.SubscribeRDO{
		IsSubscribe: true,
	}, nil
}

func (s *GRPCUsers) Unsubscribe(ctx context.Context, req *usersv1.SubscribeDTO) (*usersv1.SubscribeRDO, error) {
	bloggerId, err := uuid.Parse(req.BloggerId)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse blogger uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	subscriberId, err := uuid.Parse(req.SubscriberId)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse subscriber uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	if err := s.subsService.Unsubscribe(ctx, &servicestransfer.SubscribeInfo{
		SubscriberId: subscriberId,
		BloggerId:    bloggerId,
	}); err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to unsubscribe from blogger", logger.ErrKey, err.Error())
		return &usersv1.SubscribeRDO{
			IsSubscribe: false,
		}, err
	}

	return &usersv1.SubscribeRDO{
		IsSubscribe: true,
	}, nil
}

func (s *GRPCUsers) UploadAvatar(ctx context.Context, req *usersv1.UploadAvatarDTO) (*usersv1.UploadAvatarRDO, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to parse user uuid", logger.ErrKey, err.Error())
		return nil, err
	}

	res, err := s.userService.UploadAvatar(ctx, &servicestransfer.UploadAvatarInfo{
		UserId: userId,
		Image:  req.Image,
	})
	if err != nil {
		s.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Failed to upload avatar", logger.ErrKey, err.Error())
		return nil, err
	}

	return &usersv1.UploadAvatarRDO{
		UserId:        userId.String(),
		AvatarUrl:     res.Avatar,
		AvatarMiniUrl: res.AvatarMini,
	}, nil
}
