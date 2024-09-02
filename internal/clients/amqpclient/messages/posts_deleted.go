package messages

import "github.com/google/uuid"

const (
	Success = "success"
	Failed  = "failed"
)

type PostsDeleted struct {
	EventId uuid.UUID `json:"event_id"`
	Status  string    `json:"status"`
}
