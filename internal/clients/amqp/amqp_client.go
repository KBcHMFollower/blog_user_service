package amqp

type AmqpSender interface {
	Send(message []byte) error
}

type AmqpSenderFactory interface {
	GetSender(eventType string) (AmqpSender, error)
}

type AmqpHandlerFunc = func(message []byte) error

type AmqpClient interface {
	Publish(eventType string, body []byte) error
	Consume(target string, handler AmqpHandlerFunc) error
}
