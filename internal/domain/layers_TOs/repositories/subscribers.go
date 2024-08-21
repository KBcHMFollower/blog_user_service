package repositories_transfer

import "github.com/google/uuid"

type GetSubscriptionInfo struct {
	UserId     uuid.UUID
	Page       uint64
	Size       uint64
	TargetType string
}

type GetUserSubscribersInfo struct {
	UserId uuid.UUID
	Page   uint64
	Size   uint64
}

type GetUserSubscriptionsInfo struct {
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
