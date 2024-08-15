package amqp_client

type AmqpSender interface {
	Send(message []byte) error
}

type AmqpClient interface {
	GetSender(eventType string) (AmqpSender, error)
}
