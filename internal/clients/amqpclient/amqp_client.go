package amqpclient

import "context"

const (
	PostsDeletedEventKey = "posts-deleted-feedback"
	UserDeletedEventKey  = "user-deleted"
)

type AmqpSender interface {
	Send(message []byte) error
}

type AmqpSenderFactory interface {
	GetSender(eventType string) (AmqpSender, error)
}

type AmqpHandlerFunc = func(ctx context.Context, message []byte) error

type AmqpClient interface {
	Publish(ctx context.Context, eventType string, body []byte) error
	Consume(target string, handler AmqpHandlerFunc) error
	Stop() error
}
