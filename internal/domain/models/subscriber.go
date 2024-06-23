package models

import (
	"github.com/google/uuid"
)

type Subscriber struct {
	Id           uuid.UUID
	BloggerId    uuid.UUID
	SubscriberId uuid.UUID
}

func NewSubscriber(bloggerId uuid.UUID, sybscriberId uuid.UUID) *Subscriber {
	return &Subscriber{
		Id:           uuid.New(),
		BloggerId:    bloggerId,
		SubscriberId: sybscriberId,
	}
}

func (s *Subscriber) GetPointersArray() []interface{} {
	return []interface{}{
		&s.Id,
		&s.BloggerId,
		&s.SubscriberId,
	}
}
