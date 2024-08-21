package grpcservers

import (
	"context"
	usersv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/users"
	services_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/services"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type GRPCUsers struct {
	usersv1.UnimplementedUsersServiceServer
	userService *services.UserService
}

func RegisterUserServer(gRPC *grpc.Server, userService *services.UserService) {
	usersv1.RegisterUsersServiceServer(gRPC, &GRPCUsers{userService: userService})
}

func (s *GRPCUsers) GetUser(ctx context.Context, req *usersv1.GetUserDTO) (*usersv1.GetUserRDO, error) {

	userUuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	user, err := s.userService.GetUserById(ctx, userUuid)
	if err != nil {
		return nil, err
	}

	return &usersv1.GetUserRDO{
		User: services_transfer.ConvertUserResToProto(&user.User),
	}, nil //TODO : Change RTO
}

func (s *GRPCUsers) GetSubscribers(ctx context.Context, req *usersv1.GetSubscribersDTO) (*usersv1.GetSubscribersRDO, error) {
	bloggerId, err := uuid.Parse(req.BloggerId)
	if err != nil {
		return nil, err
	}

	subscribers, err := s.userService.GetSubscribers(ctx, &services_transfer.GetSubscribersInfo{
		BloggerId: bloggerId,
		Page:      req.Page,
		Size:      req.Size,
	})
	if err != nil {
		return nil, err
	}

	return &usersv1.GetSubscribersRDO{
		Subscribers: services_transfer.ConvertSubscribersToProto(subscribers.Subscribers),
		TotalCount:  subscribers.TotalCount,
	}, nil
}

func (s *GRPCUsers) GetSubscriptions(ctx context.Context, req *usersv1.GetSubscriptionsDTO) (*usersv1.GetSubscriptionsRDO, error) {
	subscriberId, err := uuid.Parse(req.SubscriberId)
	if err != nil {
		return nil, err
	}

	subscriptions, err := s.userService.GetSubscriptions(ctx, &services_transfer.GetSubscriptionsInfo{
		SubscriberId: subscriberId,
		Page:         req.Page,
		Size:         req.Size,
	})
	if err != nil {
		return nil, err
	}

	return &usersv1.GetSubscriptionsRDO{
		Subscriptions: services_transfer.ConvertSubscribersToProto(subscriptions.Subscriptions),
		TotalCount:    subscriptions.TotalCount,
	}, nil
}

func (s *GRPCUsers) UpdateUser(ctx context.Context, req *usersv1.UpdateUserDTO) (*usersv1.UpdateUserRDO, error) {
	userId, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	user, err := s.userService.UpdateUser(ctx, &services_transfer.UpdateUserInfo{
		Id:           userId,
		UpdateFields: services_transfer.ConvertUpdateFieldsInfoFromProto(req.UpdateData),
	})
	if err != nil {
		return nil, err
	}

	return &usersv1.UpdateUserRDO{
		User: services_transfer.ConvertUserResToProto(&user.User),
	}, nil
}

func (s *GRPCUsers) DeleteUser(ctx context.Context, req *usersv1.DeleteUserDTO) (*usersv1.DeleteUserRDO, error) {
	userId, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	if err := s.userService.DeleteUser(ctx, &services_transfer.DeleteUserInfo{
		Id: userId,
	}); err != nil {
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
		return nil, err
	}

	subscriberId, err := uuid.Parse(req.SubscriberId)
	if err != nil {
		return nil, err
	}

	if err := s.userService.Subscribe(ctx, &services_transfer.SubscribeInfo{
		SubscriberId: subscriberId,
		BloggerId:    bloggerId,
	}); err != nil {
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
		return nil, err
	}

	subscriberId, err := uuid.Parse(req.SubscriberId)
	if err != nil {
		return nil, err
	}

	if err := s.userService.Unsubscribe(ctx, &services_transfer.SubscribeInfo{
		SubscriberId: subscriberId,
		BloggerId:    bloggerId,
	}); err != nil {
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
		return nil, err
	}

	res, err := s.userService.UploadAvatar(ctx, &services_transfer.UploadAvatarInfo{
		UserId: userId,
		Image:  req.Image,
	})
	if err != nil {
		return nil, err
	}

	return &usersv1.UploadAvatarRDO{
		UserId:        userId.String(),
		AvatarUrl:     res.Avatar,
		AvatarMiniUrl: res.AvatarMini,
	}, nil
}
