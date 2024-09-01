package services_interfaces

import (
	"context"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
)

type MessagesService interface {
	UpdateMessage(ctx context.Context, updateInfo transfer.UpdateMessageInfo) error
}
