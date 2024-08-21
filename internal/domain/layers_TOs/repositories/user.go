package repositories_transfer

import "github.com/google/uuid"

type UserFieldInfo struct {
	Name  string
	Value string
}

type UpdateUserInfo struct {
	Id         uuid.UUID
	UpdateInfo []*UserFieldInfo
}

type CreateUserInfo struct {
	Email    string
	HashPass []byte
	FName    string
	LName    string
}

type DeleteUserInfo struct {
	Id uuid.UUID
}
