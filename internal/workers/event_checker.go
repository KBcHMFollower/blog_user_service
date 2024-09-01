package workers

import (
	"context"
	"errors"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	servicestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	dep "github.com/KBcHMFollower/blog_user_service/internal/workers/interfaces/dep"
	"github.com/google/uuid"
	"time"
)

const (
	workerNameLogKey = "worker-name"
)

type EventStore interface {
	dep.EventGetter
	dep.EventUpdater
}

type EventChecker struct {
	amqpClient amqpclient.AmqpClient
	eventRep   EventStore
	txCreator  dep.TransactionCreator
	logger     logger.Logger
	ctx        context.Context
}

func NewEventChecker(amqpClient amqpclient.AmqpClient, eventRep EventStore, logger logger.Logger, txCreator dep.TransactionCreator) *EventChecker {
	return &EventChecker{
		amqpClient: amqpClient,
		eventRep:   eventRep,
		logger:     logger,
		txCreator:  txCreator,
	}
}

func (as *EventChecker) Run(ctx context.Context) error {
	as.ctx = ctx
	ctx = logger.UpdateLoggerCtx(as.ctx, workerNameLogKey, "EventChecker")
	as.logger.InfoContext(ctx, "started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				events, err := as.eventRep.Events(ctx, repositoriestransfer.GetEventsInfo{
					Condition: map[repositoriestransfer.EventsConditionField]interface{}{
						repositoriestransfer.EventStatusCondition: repositoriestransfer.MessagesWaitingStatus,
					},
					Size: 50, //todo: возможно стоит передавать параметром откуда-то
				})
				if err != nil && !errors.Is(err, ctxerrors.ErrNotFound) {
					as.logger.WarnContext(ctx, "can`t get events from  db: ", "err", err.Error())
					time.Sleep(5 * time.Second)
					continue
				}

				rejectedList := make([]uuid.UUID, 0)
				sentList := make([]uuid.UUID, 0)
				for _, event := range events {
					ctx = logger.UpdateLoggerCtx(ctx, logger.EventIdKey, event.EventId)

					err = as.amqpClient.Publish(ctx, event.EventType, []byte(event.Payload))
					if err != nil {
						as.logger.ErrorContext(ctx, "can`t publish event: ", "err", err.Error())

						rejectedList = append(rejectedList, event.EventId)

						time.Sleep(5 * time.Second)
						continue
					}
					sentList = append(sentList, event.EventId)

					as.logger.InfoContext(ctx, "event published")
				}

				err = as.eventRep.UpdateMany(ctx, repositoriestransfer.UpdateManyEventInfo{
					EventId: rejectedList,
					UpdateData: map[string]interface{}{
						string(servicestransfer.StatusMsgUpdateTarget): repositoriestransfer.MessagesErrorStatus,
					},
				})

				err = as.eventRep.UpdateMany(ctx, repositoriestransfer.UpdateManyEventInfo{
					EventId: sentList,
					UpdateData: map[string]interface{}{
						string(servicestransfer.StatusMsgUpdateTarget): repositoriestransfer.MessagesSentStatus,
					},
				})

				time.Sleep(5 * time.Second)
			}

		}
	}()

	return nil
}

func (as *EventChecker) Stop() {
	as.ctx.Done()

	as.logger.InfoContext(as.ctx, "worker died")
}
