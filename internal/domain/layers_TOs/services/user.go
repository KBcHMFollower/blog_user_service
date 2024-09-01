package services_transfer

import (
	usersv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/users"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type UserFieldTarget string

const (
	UserEmailUpdateTarget UserFieldTarget = "email"
	UserFNameUpdateTarget UserFieldTarget = "fname"
	UserLNameUpdateTarget UserFieldTarget = "lname"
)

type UpdateUserInfo struct {
	Id           uuid.UUID               `validate:"required,uuid"`
	UpdateFields map[UserFieldTarget]any `validate:"required,mapkeys-user-update"`
}

type DeleteUserInfo struct {
	Id uuid.UUID `validate:"required,uuid"`
}

type UserResult struct {
	Email      string
	FName      string
	LName      string
	Avatar     string
	AvatarMini string
	Id         uuid.UUID
}

type GetUserResult struct {
	User UserResult
}

type UpdateUserResult struct {
	User UserResult
}

func GetUserResultFromModel(user *models.User) UserResult {
	return UserResult{
		Email:      user.Email,
		FName:      user.FName,
		LName:      user.LName,
		Avatar:     user.Avatar,
		Id:         user.Id,
		AvatarMini: user.AvatarMin,
	}
}

func ConvertUserResToProto(user *UserResult) *usersv1.User {
	return &usersv1.User{
		Id:        user.Id.String(),
		Email:     user.Email,
		Fname:     user.FName,
		Lname:     user.LName,
		Avatar:    user.Avatar,
		AvatarMin: user.AvatarMini,
	}
}
