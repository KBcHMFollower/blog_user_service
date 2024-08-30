package repositories_transfer

import "github.com/google/uuid"

type MessageStatus string

const (
	MessagesSentStatus    MessageStatus = "sent"
	MessagesErrorStatus   MessageStatus = "error"
	MessagesWaitingStatus MessageStatus = "waiting"
)

type CreateEventInfo struct {
	EventId   uuid.UUID `json:"event_id"`
	EventType string    `json:"event_type"`
	Payload   []byte    `json:"payload"`
}
