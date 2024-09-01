package repositories_transfer

import (
	"github.com/google/uuid"
)

type EventsConditionField string

const (
	EventStatusCondition EventsConditionField = "status"
)

type MessageStatus string

const (
	MessagesSuccessStatus MessageStatus = "success"
	MessagesSentStatus    MessageStatus = "sent"
	MessagesErrorStatus   MessageStatus = "error"
	MessagesWaitingStatus MessageStatus = "waiting"
)

type GetEventsInfo struct {
	Page      uint64
	Size      uint64
	Condition map[EventsConditionField]interface{}
}

type CreateEventInfo struct {
	EventId   uuid.UUID `json:"event_id"`
	EventType string    `json:"event_type"`
	Payload   []byte    `json:"payload"`
}

type UpdateEventInfo struct {
	EventId    uuid.UUID `json:"event_id"`
	UpdateData map[string]any
}

type UpdateManyEventInfo struct {
	EventId    []uuid.UUID `json:"event_id"`
	UpdateData map[string]any
}
