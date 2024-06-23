package grpcservers

import (
	"context"
	usersv1 "github.com/KBcHMFollower/auth-service/api/protos/gen/users"
	"github.com/KBcHMFollower/auth-service/internal/services"
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
	return s.userService.GetUserById(ctx, req)
}
func (s *GRPCUsers) GetSubscribers(ctx context.Context, req *usersv1.GetSubscribersDTO) (*usersv1.GetSubscribersRDO, error) {
	return s.userService.GetSubscribers(ctx, req)
}
func (s *GRPCUsers) GetSubscriptions(ctx context.Context, req *usersv1.GetSubscriptionsDTO) (*usersv1.GetSubscriptionsRDO, error) {
	return s.userService.GetSubscriptions(ctx, req)
}
func (s *GRPCUsers) UpdateUser(ctx context.Context, req *usersv1.UpdateUserDTO) (*usersv1.UpdateUserRDO, error) {
	return s.userService.UpdateUser(ctx, req)
}
func (s *GRPCUsers) DeleteUser(ctx context.Context, req *usersv1.DeleteUserDTO) (*usersv1.DeleteUserRDO, error) {
	return s.userService.DeleteUser(ctx, req)
}

func (s *GRPCUsers) Subscribe(ctx context.Context, req *usersv1.SubscribeDTO) (*usersv1.SubscribeRDO, error) {
	return s.userService.Subscribe(ctx, req)
}
func (s *GRPCUsers) Unsubscribe(ctx context.Context, req *usersv1.SubscribeDTO) (*usersv1.SubscribeRDO, error) {
	return s.userService.Unsubscribe(ctx, req)
}

func (s *GRPCUsers) UploadAvatar(ctx context.Context, req *usersv1.UploadAvatarDTO) (*usersv1.UploadAvatarRDO, error) {
	return s.userService.UploadAvatar(ctx, req)
}
