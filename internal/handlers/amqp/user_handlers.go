package amqp_handlers

import (
	"context"
	"encoding/json"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient/messages"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	servicestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	servicesinterfaces "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces"
)

type UserHandler struct {
	messService servicesinterfaces.MessagesService
	userService servicesinterfaces.UserService
	log         logger.Logger
}

func NewUserHandler(
	usrService servicesinterfaces.UserService,
	messService servicesinterfaces.MessagesService,
	log logger.Logger,
) *UserHandler {
	return &UserHandler{
		userService: usrService,
		messService: messService,
		log:         log,
	}
}

func (handler *UserHandler) HandlePostDeletingEvent(ctx context.Context, message []byte) error {

	var resInfo messages.PostsDeleted

	if err := json.Unmarshal(message, &resInfo); err != nil {
		handler.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Error unmarshalling post deleting event", logger.ErrKey, err.Error())
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("error unmarshalling post deleting event", err))
	}

	ctx = logger.UpdateLoggerCtx(ctx, logger.EventIdKey, resInfo.EventId)

	if resInfo.Status == "OK" {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(
			"error in setting success status",
			handler.messService.UpdateMessage(ctx, servicestransfer.UpdateMessageInfo{
				EventId: resInfo.EventId,
				UpdateData: map[servicestransfer.EventUpdateTarget]any{
					servicestransfer.StatusMsgUpdateTarget: repositoriestransfer.MessagesSuccessStatus,
				},
			}),
		))
	}
	if err := handler.userService.CompensateDeletedUser(ctx, resInfo.EventId); err != nil {
		handler.log.ErrorContext(ctxerrors.ErrorCtx(ctx, err), "Error calling CompensateDeletedUser", logger.ErrKey, err.Error())
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("error calling CompensateDeletedUser", err))
	}

	return nil
}
