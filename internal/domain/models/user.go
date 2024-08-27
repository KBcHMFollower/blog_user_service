package models

import (
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
