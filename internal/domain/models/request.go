package models

import (
	"github.com/google/uuid"
)

type Request struct {
	Id             uuid.UUID `db:"id"`
	IdempotencyKey uuid.UUID `db:"idempotency_key"`
}
