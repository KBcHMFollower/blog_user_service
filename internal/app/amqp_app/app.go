package amqp_app

import (
	"github.com/KBcHMFollower/blog_user_service/internal/amqp_client"
)

type AmqpApp struct {
	client   amqp_client.AmqpClient
	handlers map[string]amqp_client.AmqpHandlerFunc
}

func NewAmqpApp(client amqp_client.AmqpClient) *AmqpApp {
	return &AmqpApp{
		client:   client,
		handlers: make(map[string]amqp_client.AmqpHandlerFunc),
	}
}

func (app *AmqpApp) RegisterHandler(name string, handler amqp_client.AmqpHandlerFunc) {
	app.handlers[name] = handler
}

func (app *AmqpApp) Start() error {
	for name, handler := range app.handlers {
		err := app.client.Consume(name, handler)
		if err != nil {
			return err
		}
	}

	return nil
}
