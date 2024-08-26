package models

import (
	usersv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/users"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id          uuid.UUID `db:"id"`
	Email       string    `db:"email"`
	FName       string    `db:"fname"`
	LName       string    `db:"lname"`
	Avatar      string    `db:"avatar"`
	AvatarMin   string    `db:"avatar_min"`
	PassHash    []byte    `db:"pass_hash"`
	CreatedDate time.Time `db:"created_date"`
	UpdatedDate time.Time `db:"updated_date"`
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
		&u.PassHash,
		&u.CreatedDate,
		&u.UpdatedDate,
	}
}
