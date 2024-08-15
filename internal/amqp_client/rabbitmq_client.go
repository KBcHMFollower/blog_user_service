package amqp_client

import (
	"fmt"
	"github.com/streadway/amqp"
)

const (
	DeleteUserExchange    = "direct-user-actions"
	UserDeletedQueue      = "user-deleted"
	UserPostsDeletedQueue = "user-posts-deleted"
	UserCompensateQueue   = "user-compensate"
)

type RabbitMQClient struct {
	senderMap map[string]AmqpSender
	conn      *amqp.Connection
	ch        *amqp.Channel
}

func NewRabbitMQClient(addr string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %s", err)
	}

	senderMap := map[string]AmqpSender{
		"userDeleted": &UserDeletedSender{ch: ch},
	}

	if err := DeclareExchanges(ch); err != nil {
		return nil, fmt.Errorf("failed to declare exchanges: %s", err)
	}
	if err := DeclareQueues(ch); err != nil {
		return nil, fmt.Errorf("failed to declare queues: %s", err)
	}

	return &RabbitMQClient{ch: ch, conn: conn, senderMap: senderMap}, nil
}

func (rc *RabbitMQClient) Close() error {
	if err := rc.ch.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ channel: %s", err)
	}
	if err := rc.conn.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ connection: %s", err)
	}

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

func (rc *RabbitMQClient) GetSender(senderName string) (AmqpSender, error) {
	sender, ok := rc.senderMap[senderName]
	if !ok {
		return nil, fmt.Errorf("sender not found for sender %s", senderName)
	}

	return sender, nil
}

type UserDeletedSender struct {
	ch *amqp.Channel
}

func (uds *UserDeletedSender) Send(message []byte) error {
	if err := uds.ch.Publish(
		DeleteUserExchange,
		UserDeletedQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	); err != nil {
		return fmt.Errorf("failed to send message: %s", err)
	}

	return nil
}
