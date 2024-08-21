package amqp_app

import (
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqp"
)

type AmqpApp struct {
	client   amqp.AmqpClient
	handlers map[string]amqp.AmqpHandlerFunc
}

func NewAmqpApp(client amqp.AmqpClient) *AmqpApp {
	return &AmqpApp{
		client:   client,
		handlers: make(map[string]amqp.AmqpHandlerFunc),
	}
}

func (app *AmqpApp) RegisterHandler(name string, handler amqp.AmqpHandlerFunc) {
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

//TODO: STOP
