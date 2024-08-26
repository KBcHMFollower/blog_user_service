package models

import (
	"github.com/google/uuid"
)

type Subscriber struct {
	Id           uuid.UUID `db:"id"`
	BloggerId    uuid.UUID `db:"blogger_id"`
	SubscriberId uuid.UUID `db:"subscriber_id"`
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
