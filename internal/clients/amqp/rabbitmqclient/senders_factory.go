package rabbitmqclient

import (
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqp"
	"github.com/streadway/amqp"
)

type SendersStore struct {
	ch         *amqp.Channel
	sendersMap map[string]amqp.AmqpSender
}

func NewSendersStore(ch *amqp.Channel) *SendersStore {
	sendersMap := map[string]amqp.AmqpSender{
		"userDeleted": &UserDeletedSender{ch: ch},
	}

	return &SendersStore{
		ch:         ch,
		sendersMap: sendersMap,
	}
}

func (ss *SendersStore) GetSender(senderName string) (amqp.AmqpSender, error) {
	sender, ok := ss.sendersMap[senderName]
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
