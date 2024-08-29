package workers

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	dep "github.com/KBcHMFollower/blog_user_service/internal/workers/interfaces/dep"
	"github.com/google/uuid"
	"log/slog"
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
	logger     *slog.Logger
	ctx        context.Context
}

func NewEventChecker(amqpClient amqpclient.AmqpClient, eventRep EventStore, logger *slog.Logger, txCreator dep.TransactionCreator) *EventChecker {
	return &EventChecker{
		amqpClient: amqpClient,
		eventRep:   eventRep,
		logger:     logger,
		txCreator:  txCreator,
	}
}

func (as *EventChecker) Run(ctx context.Context) error {
	as.ctx = ctx
	logger.UpdateLoggerCtx(as.ctx, workerNameLogKey, "EventChecker")
	as.logger.Info("started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				events, err := as.eventRep.GetEventsWithStatus(ctx, repository.MessagesWaitingStatus, 50)
				if err != nil {
					as.logger.Warn("can`t get events from  db: ", "err", err.Error())
					time.Sleep(5 * time.Second)
					continue
				}

				rejectedList := make([]uuid.UUID, 0)
				sentList := make([]uuid.UUID, 0)
				for _, event := range events {
					logger.UpdateLoggerCtx(ctx, logger.EventIdKey, event.EventId)

					err = as.amqpClient.Publish(ctx, event.EventType, []byte(event.Payload))
					if err != nil {
						as.logger.Error("can`t publish event: ", "err", err.Error())

						rejectedList = append(rejectedList, event.EventId)

						time.Sleep(5 * time.Second)
						continue
					}
					sentList = append(sentList, event.EventId)

					as.logger.Info("event published")
				}

				err = as.eventRep.SetStatusInEvents(ctx, rejectedList, repository.MessagesErrorStatus)
				if err != nil {
					as.logger.Error("can`t set status in events: ", "err", err.Error())
					time.Sleep(5 * time.Second)
					continue
				}

				err = as.eventRep.SetStatusInEvents(ctx, sentList, repository.MessagesSentStatus)
				if err != nil {
					as.logger.Error("can`t set status in events: ", "err", err.Error())
					time.Sleep(5 * time.Second)
					continue
				}

				time.Sleep(5 * time.Second)
			}

		}
	}()

	return nil
}

func (as *EventChecker) Stop() {
	as.ctx.Done()

	as.logger.Info("worker died")
}
