package amqp_handlers

import (
	"context"
	"encoding/json"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient/messages"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	services_interfaces "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces"
	"log/slog"
)

type UserHandler struct {
	userService services_interfaces.UserService
	log         *slog.Logger
}

func NewUserHandler(usrService services_interfaces.UserService, log *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: usrService,
		log:         log,
	}
}

func (handler *UserHandler) HandlePostDeletingEvent(ctx context.Context, message []byte) error {
	var resInfo messages.PostsDeleted

	if err := json.Unmarshal(message, &resInfo); err != nil {
		handler.log.ErrorContext(ctx, "Error unmarshalling post deleting event", logger.ErrKey, err.Error())
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("error unmarshalling post deleting event", err))
	}

	if resInfo.Status == "OK" {
		return nil
	} //TODO: ЧЕТО НУЖНО ДЕЛАТЬ
	if err := handler.userService.CompensateDeletedUser(context.TODO(), resInfo.EventId); err != nil {
		handler.log.ErrorContext(ctx, "Error calling CompensateDeletedUser", logger.ErrKey, err.Error())
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("error calling CompensateDeletedUser", err))
	}

	return nil
}
