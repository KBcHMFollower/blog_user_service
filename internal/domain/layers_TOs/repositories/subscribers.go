package repositories_transfer

import "github.com/google/uuid"

type GetSubType = string

const (
	SubscribersTarget   GetSubType = "blogger_id"
	SubscriptionsTarget GetSubType = "subscriber_id"
)

type GetSubsCountInfo struct {
	Condition map[GetSubType]any
}

type GetSubsInfo struct {
	Condition map[GetSubType]any
	Page      uint64
	Size      uint64
}

type SubscribeToUserInfo struct {
	BloggerId    uuid.UUID
	SubscriberId uuid.UUID
}

type UnsubscribeInfo struct {
	BloggerId    uuid.UUID
	SubscriberId uuid.UUID
}
