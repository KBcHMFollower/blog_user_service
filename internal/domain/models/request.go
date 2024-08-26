package models

import (
	"github.com/google/uuid"
)

type Request struct {
	Id             uuid.UUID `db:"id"`
	IdempotencyKey uuid.UUID `db:"idempotency_key"`
}

func (r *Request) GetPointersArray() []interface{} {
	return []interface{}{
		&r.Id,
		&r.IdempotencyKey,
	}
}
