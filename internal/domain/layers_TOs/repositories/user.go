package repositories_transfer

import "github.com/google/uuid"

type UserFieldTarget string

const (
	UserIdCondition    UserFieldTarget = "id"
	UserEmailCondition UserFieldTarget = "email"
)

type GetUsersCountInfo struct {
	Condition map[UserFieldTarget]interface{}
}

type GetUsersInfo struct {
	Size      uint64
	Page      uint64
	Condition map[UserFieldTarget]interface{}
}

type GetUserInfo struct {
	Condition map[UserFieldTarget]interface{}
}

type UpdateUserInfo struct {
	Id         uuid.UUID
	UpdateInfo map[string]any
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
