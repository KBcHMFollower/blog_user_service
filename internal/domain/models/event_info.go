package models

import "github.com/google/uuid"

type EventInfo struct {
	EventId    uuid.UUID `json:"event_id"`
	EventType  string    `json:"event_type"`
	Payload    string    `json:"payload"`
	Status     string    `json:"status"`
	RetryCount int32     `json:"retry_count"`
}

func (s *EventInfo) GetPointersArray() []interface{} {
	return []interface{}{
		&s.EventId,
		&s.EventType,
		&s.Payload,
		&s.Status,
		&s.RetryCount,
	}
}
