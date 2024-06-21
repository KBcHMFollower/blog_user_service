package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id          uuid.UUID
	Email       string
	Fname       string
	Lname       string
	PassHash    []byte
	CreatedDate time.Time
	UpdatedDate time.Time
}

func NewUserModel(email string, fname string, lname string, hashPass []byte) *User {
	now := time.Now()
	id := uuid.New()

	return &User{
		Id:          id,
		Email:       email,
		PassHash:    hashPass,
		Fname:       fname,
		Lname:       lname,
		CreatedDate: now,
		UpdatedDate: now,
	}
}
