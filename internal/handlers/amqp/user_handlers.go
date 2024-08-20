package amqp_handlers

import (
	"context"
	"encoding/json"
	"fmt"
	authv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/auth"
	"github.com/KBcHMFollower/blog_user_service/internal/amqp_client/messages"
	"github.com/KBcHMFollower/blog_user_service/internal/services"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(usrService services.UserService) *UserHandler {
	return &UserHandler{
		userService: usrService,
	}
}

func (handler *UserHandler) HandlePostDeletingEvent(message []byte) error {
	var user *messages.User

	if err := json.Unmarshal(message, user); err != nil {
		return fmt.Errorf("can`t unmarshal user message: %v", err)
	}

	if _, err := handler.userService.RegisterUser(context.TODO(), authv1.RegisterDTO{
		Email:    user.Email,
		Password: "123",
		Fname:    user.FName,
		Lname:    user.LName,
	}); err != nil {
		return fmt.Errorf("can`t register user: %v", err)
	} //TODO

	return nil
}
