package repositories_transfer

import "github.com/google/uuid"

type getSubType = string

const (
	SubscribersTarget   getSubType = "blogger_id"
	SubscriptionsTarget getSubType = "subscriber_id"
)

type GetSubscriptionInfo struct {
	UserId     uuid.UUID
	Page       uint64
	Size       uint64
	TargetType getSubType
}

type GetSubsInfo struct {
	Target getSubType
	UserId uuid.UUID
	Page   uint64
	Size   uint64
}

type SubscribeToUserInfo struct {
	BloggerId    uuid.UUID
	SubscriberId uuid.UUID
}

type UnsubscribeInfo struct {
	BloggerId    uuid.UUID
	SubscriberId uuid.UUID
}
