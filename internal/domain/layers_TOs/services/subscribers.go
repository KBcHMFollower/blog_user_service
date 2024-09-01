package services_transfer

import (
	usersv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/users"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type SubscriberResult struct {
	Email      string
	FName      string
	LName      string
	Avatar     string
	AvatarMini string
	Id         uuid.UUID
}

type GetSubscribersInfo struct {
	BloggerId uuid.UUID `validate:"required,uuid"`
	Page      int32     `validate:"required,gte=1,lte=100"`
	Size      int32     `validate:"required,gte=1,lte=1000"`
}

type GetSubscriptionsInfo struct {
	SubscriberId uuid.UUID `validate:"required,uuid"`
	Page         int32     `validate:"required,gte=1,lte=100"`
	Size         int32     `validate:"required,gte=1,lte=1000"`
}

type SubscribeInfo struct {
	BloggerId    uuid.UUID `validate:"required,uuid"`
	SubscriberId uuid.UUID `validate:"required,uuid"`
}

type GetSubscribersResult struct {
	Subscribers []SubscriberResult
	TotalCount  int32
}

type GetSubscriptionsResult struct {
	Subscriptions []SubscriberResult
	TotalCount    int32
}

func GetSubscribersArrayResultFromModel(users []*models.User) []SubscriberResult {
	var results []SubscriberResult = make([]SubscriberResult, 0, len(users))

	for _, user := range users {
		results = append(results, SubscriberResult{
			Email:      user.Email,
			LName:      user.LName,
			Avatar:     user.Avatar,
			AvatarMini: user.AvatarMin,
			Id:         user.Id,
		})
	}

	return results
}

func ConvertSubscribersToProto(users []SubscriberResult) []*usersv1.User {
	var results []*usersv1.User = make([]*usersv1.User, 0, len(users))

	for _, user := range users {
		results = append(results, &usersv1.User{
			Email:     user.Email,
			Fname:     user.FName,
			Lname:     user.LName,
			Avatar:    user.Avatar,
			AvatarMin: user.AvatarMini,
			Id:        user.Id.String(),
		})
	}

	return results
}
