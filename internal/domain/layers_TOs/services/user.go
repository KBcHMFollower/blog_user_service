package services_transfer

import (
	usersv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/users"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
	"time"
)

type UpdateFieldInfo struct {
	Name  string
	Value string
}

type UpdateUserInfo struct {
	Id           uuid.UUID
	UpdateFields []UpdateFieldInfo
}

type DeleteUserInfo struct {
	Id uuid.UUID
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

func ConvertUpdateFieldsInfoFromProto(updateFields []*usersv1.UpdateUserItem) []UpdateFieldInfo {
	var results []UpdateFieldInfo = make([]UpdateFieldInfo, 0, len(updateFields))

	for _, field := range updateFields {
		results = append(results, UpdateFieldInfo{
			Name:  field.Name,
			Value: field.Value,
		})
	}

	return results
}

func ConvertUserResToProto(user *UserResult) *usersv1.User {
	return &usersv1.User{
		Id:          user.Id.String(),
		Email:       user.Email,
		Fname:       user.FName,
		Lname:       user.FName,
		Avatar:      user.Avatar,
		AvatarMin:   user.AvatarMini,
		CreatedDate: time.Now().String(),
		UpdatedDate: time.Now().String(),
		IsDeleted:   false,
	} //TODO: Change RDO
}
