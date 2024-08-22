package amqp_handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqp/messages"
	services_interfaces "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces"
)

type UserHandler struct {
	userService services_interfaces.UserService
}

func NewUserHandler(usrService services_interfaces.UserService) *UserHandler {
	return &UserHandler{
		userService: usrService,
	}
}

func (handler *UserHandler) HandlePostDeletingEvent(message []byte) error {
	var resInfo messages.PostsDeleted

	if err := json.Unmarshal(message, &resInfo); err != nil {
		return fmt.Errorf("can`t unmarshal user message: %v", err)
	}

	if resInfo.Status != "OK" {
		return fmt.Errorf("status not OK: %v", resInfo.Status)
	}
	if err := handler.userService.CompensateDeletedUser(context.TODO(), resInfo.EventId); err != nil {
		return fmt.Errorf("can`t compensate user: %v", err)
	}

	return nil
}
