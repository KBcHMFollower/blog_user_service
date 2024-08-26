package messages

import (
	"github.com/google/uuid"
	"time"
)

type UserMessage struct {
	Id          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	FName       string    `json:"fname"`
	LName       string    `json:"lname"`
	Avatar      string    `json:"avatar"`
	AvatarMin   string    `json:"avatar_min"`
	PassHash    []byte    `json:"pass_hash"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

type UserDeletedMessage struct {
	EventId uuid.UUID
	User    UserMessage
}
