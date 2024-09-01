package services

import (
	"context"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	servicesutils "github.com/KBcHMFollower/blog_user_service/internal/services/lib"
	dep "github.com/KBcHMFollower/blog_user_service/internal/workers/interfaces/dep"
)

type msgSvcMessagesStore interface {
	dep.EventUpdater
}

type MessagesService struct {
	messRep msgSvcMessagesStore
	log     logger.Logger
}

func NewMessagesService(messRep msgSvcMessagesStore, log logger.Logger) *MessagesService {
	return &MessagesService{messRep: messRep, log: log}
}

func (ms *MessagesService) UpdateMessage(ctx context.Context, updateInfo transfer.UpdateMessageInfo) error {
	err := ms.messRep.Update(ctx, repositoriestransfer.UpdateEventInfo{
		EventId:    updateInfo.EventId,
		UpdateData: servicesutils.ConvertMapKeysToStrings(updateInfo.UpdateData),
	})

	return err
}
