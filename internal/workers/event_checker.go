package workers

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	dep "github.com/KBcHMFollower/blog_user_service/internal/workers/interfaces/dep"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type EventStore interface {
	dep.EventGetter
	dep.EventUpdater
}

type EventChecker struct {
	amqpClient amqpclient.AmqpClient
	eventRep   EventStore
	txCreator  dep.TransactionCreator
	logger     *slog.Logger
}

func NewEventChecker(amqpClient amqpclient.AmqpClient, eventRep EventStore, logger *slog.Logger, txCreator dep.TransactionCreator) *EventChecker {
	return &EventChecker{
		amqpClient: amqpClient,
		eventRep:   eventRep,
		logger:     logger,
		txCreator:  txCreator,
	}
}

func (as EventChecker) Run(ctx context.Context) error {
	log := as.logger.
		With("op", "EventChecker.Run")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				events, err := as.eventRep.GetEventsWithStatus(ctx, repository.MessagesWaitingStatus, 50)
				if err != nil {
					log.Error("can`t get events from  db: ", err)
					time.Sleep(5 * time.Second)
					continue
				}

				rejectedList := make([]uuid.UUID, 0)
				sentList := make([]uuid.UUID, 0)
				for _, event := range events {
					err = as.amqpClient.Publish(event.EventType, []byte(event.Payload))
					if err != nil {
						log.Error("can`t publish event: ", err)
						rejectedList = append(rejectedList, event.EventId)
						time.Sleep(5 * time.Second)
						continue
					}
					sentList = append(sentList, event.EventId)
				}

				err = as.eventRep.SetStatusInEvents(ctx, rejectedList, repository.MessagesErrorStatus)
				if err != nil {
					log.Error("can`t set status in events: ", err)
					time.Sleep(5 * time.Second)
					continue
				}

				err = as.eventRep.SetStatusInEvents(ctx, sentList, repository.MessagesSentStatus)
				if err != nil {
					log.Error("can`t set status in events: ", err)
					time.Sleep(5 * time.Second)
					continue
				}

				time.Sleep(5 * time.Second)
			}

		}
	}()

	return nil
}
