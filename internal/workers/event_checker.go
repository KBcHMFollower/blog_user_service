package workers

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/amqp_client"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type EventChecker struct {
	amqpClient amqp_client.AmqpClient
	eventRep   repository.EventStore
	logger     *slog.Logger
}

func NewEventChecker(amqpClient amqp_client.AmqpClient, eventRep repository.EventStore, logger *slog.Logger) *EventChecker {
	return &EventChecker{
		amqpClient: amqpClient,
		eventRep:   eventRep,
		logger:     logger,
	}
}

func (as EventChecker) Run() {
	log := as.logger.
		With("op", "EventChecker.Run")

	go func() {
		for {
			events, err := as.eventRep.GetEvents(context.TODO(), "status", "waiting", 50)
			if err != nil {
				log.Error("can`t get events from  db: ", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, event := range events {
				sender, err := as.amqpClient.GetSender(event.EventType)
				if err != nil {
					log.Error("can`t get event sender: ", err)
					time.Sleep(5 * time.Second)
					continue
				}

				err = sender.Send([]byte(event.Payload))
				if err != nil {
					log.Error("can`t publish event: ", err)
					time.Sleep(5 * time.Second)
					continue
				}
			}

			eventIds := make([]uuid.UUID, 0, len(events))
			for _, event := range events {
				eventIds = append(eventIds, event.EventId)
			}

			err = as.eventRep.SetSentStatusesInEvents(context.TODO(), eventIds)
			if err != nil {
				log.Error("can`t set status in events: ", err)
				time.Sleep(5 * time.Second)
				continue
			}

			time.Sleep(5 * time.Second)
		}
	}()
}
