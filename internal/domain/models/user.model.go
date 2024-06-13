package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id uuid.UUID
	Email string
	PassHash []byte
	CreatedDate time.Time
	UpdatedDate time.Time
}

func NewUserModel(email string, hashPass []byte) (*User){
	now := time.Now()
	id := uuid.New()

	return &User{
		Id: id,
		Email: email,
		PassHash: hashPass,
		CreatedDate: now,
		UpdatedDate: now,
	}
}