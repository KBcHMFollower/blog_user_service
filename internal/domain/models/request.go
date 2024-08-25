package models

import (
	"github.com/google/uuid"
)

type Request struct {
	Id             uuid.UUID
	IdempotencyKey uuid.UUID
	Payload        []byte //TODO: УБРАТЬ
	Status         string
}

func (r *Request) GetPointersArray() []interface{} {
	return []interface{}{
		&r.Id,
		&r.IdempotencyKey,
		&r.Payload,
		&r.Status,
	}
}
