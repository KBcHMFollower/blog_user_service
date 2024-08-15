package models

import (
	usersv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/users"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	FName       string    `json:"fname"`
	LName       string    `json:"lname"`
	Avatar      string    `json:"avatar"`
	AvatarMin   string    `json:"avatar_min"`
	IsDeleted   bool      `json:"is_deleted"`
	PassHash    []byte    `json:"pass_hash"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

func NewUserModel(email string, fName string, lName string, hashPass []byte) *User {
	return &User{
		Id:        uuid.New(),
		Email:     email,
		PassHash:  hashPass,
		Avatar:    "defaultAvatar",
		AvatarMin: "defaultAvatarMin",
		FName:     fName,
		LName:     lName,
	}
}

func (u *User) ConvertToProto() *usersv1.User {
	return &usersv1.User{
		Id:          u.Id.String(),
		Email:       u.Email,
		Fname:       u.FName,
		Lname:       u.LName,
		Avatar:      u.Avatar,
		AvatarMin:   u.AvatarMin,
		IsDeleted:   u.IsDeleted,
		CreatedDate: u.CreatedDate.String(),
		UpdatedDate: u.UpdatedDate.String(),
	}
}

func UsersArrayToProto(users []*User) []*usersv1.User {
	usersProto := make([]*usersv1.User, 0)

	for _, user := range users {
		usersProto = append(usersProto, user.ConvertToProto())
	}

	return usersProto
}

func (u *User) GetPointersArray() []interface{} {
	return []interface{}{
		&u.Id,
		&u.Email,
		&u.FName,
		&u.LName,
		&u.Avatar,
		&u.AvatarMin,
		&u.IsDeleted,
		&u.PassHash,
		&u.CreatedDate,
		&u.UpdatedDate,
	}
}
