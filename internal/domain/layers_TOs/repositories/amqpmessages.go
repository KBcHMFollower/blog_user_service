package repositories_transfer

import "github.com/google/uuid"

type CreateEventInfo struct {
	EventId   uuid.UUID `json:"event_id"`
	EventType string    `json:"event_type"`
	Payload   []byte    `json:"payload"`
}
