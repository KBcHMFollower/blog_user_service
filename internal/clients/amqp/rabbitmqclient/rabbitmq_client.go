package rabbitmqclient

import (
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqp"
	"github.com/streadway/amqp"
	"log"
)

const (
	DeleteUserExchange    = "direct-user-actions"
	UserDeletedQueue      = "user-deleted"
	UserPostsDeletedQueue = "user-posts-deleted"
	UserCompensateQueue   = "user-compensate"
)

type RabbitMQClient struct {
	pubConn        *amqp.Connection
	pubCh          *amqp.Channel
	consConn       *amqp.Connection
	consCh         *amqp.Channel
	sendersFactory amqp.AmqpSenderFactory
}

func NewRabbitMQClient(addr string) (*RabbitMQClient, error) {
	pConn, err := amqp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %s", err)
	}
	cConn, err := amqp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %s", err)
	}

	pCh, err := pConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %s", err)
	}
	cCh, err := cConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %s", err)
	}

	if err := DeclareExchanges(pCh); err != nil {
		return nil, fmt.Errorf("failed to declare exchanges: %s", err)
	}
	if err := DeclareQueues(pCh); err != nil {
		return nil, fmt.Errorf("failed to declare queues: %s", err)
	}

	sendersFactory := NewSendersStore(pCh)

	return &RabbitMQClient{pubConn: cConn, consConn: cConn, pubCh: pCh, consCh: cCh, sendersFactory: sendersFactory}, nil
}

func (rc *RabbitMQClient) Close() error {
	if err := rc.pubCh.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ channel: %s", err)
	}
	if err := rc.pubConn.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ connection: %s", err)
	}
	if err := rc.consCh.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ channel: %s", err)
	}
	if err := rc.consConn.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ connection: %s", err)
	}

	return nil
}

func (rc *RabbitMQClient) Publish(eventType string, body []byte) error {
	sender, err := rc.sendersFactory.GetSender(eventType)
	if err != nil {
		return fmt.Errorf("failed to get sender for event %s: %s", eventType, err)
	}

	if err := sender.Send(body); err != nil {
		return fmt.Errorf("failed to send event %s: %s", eventType, err)
	}

	return nil
}

func (rc *RabbitMQClient) Consume(queueName string, handler amqp.AmqpHandlerFunc) error {
	del, err := rc.consCh.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		return fmt.Errorf("failed to register a consumer for queue %s: %s", queueName, err)
	}

	go func() {
		for d := range del {
			if err := handler(d.Body); err != nil {
				log.Printf("failed to handle a message from queue %s: %s", queueName, err)
				if err := d.Nack(false, false); err != nil {
					log.Printf("failed to nack: %s", err)
				}
				continue
			}
			if err := d.Ack(false); err != nil {
				log.Printf("failed to ack: %s", err)
			}
		}
	}()

	return nil
}

func DeclareExchanges(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(
		DeleteUserExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DeleteUser exchange: %s", err)
	}

	return nil
}

func DeclareQueues(ch *amqp.Channel) error {
	q, err := ch.QueueDeclare(
		UserDeletedQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare UserDeleted queue: %s", err)
	}

	if err = ch.QueueBind(
		q.Name,
		UserDeletedQueue,
		DeleteUserExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind UserDeleted queue: %s", err)
	}

	q, err = ch.QueueDeclare(
		UserPostsDeletedQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare UserDeleted queue: %s", err)
	}

	if err = ch.QueueBind(
		q.Name,
		UserPostsDeletedQueue,
		DeleteUserExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind UserDeleted queue: %s", err)
	}

	q, err = ch.QueueDeclare(
		UserCompensateQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare UserDeleted queue: %s", err)
	}

	if err = ch.QueueBind(
		q.Name,
		UserCompensateQueue,
		DeleteUserExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind UserDeleted queue: %s", err)
	}

	return nil
}
