package models

import "github.com/google/uuid"

type EventInfo struct {
	EventId    uuid.UUID `db:"event_id"`
	EventType  string    `db:"event_type"`
	Payload    string    `db:"payload"`
	Status     string    `db:"status"`
	RetryCount int32     `db:"retry_count"`
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
