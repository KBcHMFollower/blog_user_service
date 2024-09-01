package services_transfer

import (
	"github.com/google/uuid"
)

type EventUpdateTarget string

const (
	StatusMsgUpdateTarget EventUpdateTarget = "status"
	RetryMsgUpdateTarget  EventUpdateTarget = "retry"
)

type UpdateMessageInfo struct {
	EventId    uuid.UUID
	UpdateData map[EventUpdateTarget]any
}
