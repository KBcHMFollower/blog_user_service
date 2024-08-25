package messages

import "github.com/google/uuid"

type PostsDeleted struct {
	EventId uuid.UUID `json:"event_id"`
	Status  string    `json:"status"`
}
