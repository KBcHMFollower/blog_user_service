package amqp_app

import (
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqp"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqp/rabbitmqclient"
	"github.com/KBcHMFollower/blog_user_service/internal/config"
)

type AmqpApp struct {
	Client   amqp.AmqpClient
	handlers map[string]amqp.AmqpHandlerFunc
}

func NewAmqpApp(rabbitmqConnectInfo config.RabbitMq) (*AmqpApp, error) {
	rabbitMqApp, err := rabbitmqclient.NewRabbitMQClient(rabbitmqConnectInfo.Addr)
	if err != nil {
		return nil, fmt.Errorf("new rabbitmq Client error: %v", err)
	}

	return &AmqpApp{
		Client:   rabbitMqApp,
		handlers: make(map[string]amqp.AmqpHandlerFunc),
	}, nil
}

func (app *AmqpApp) RegisterHandler(name string, handler amqp.AmqpHandlerFunc) {
	app.handlers[name] = handler
}

func (app *AmqpApp) Start() error {
	for name, handler := range app.handlers {
		err := app.Client.Consume(name, handler)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *AmqpApp) Stop() error {
	if err := app.Client.Stop(); err != nil {
		return fmt.Errorf("stop rabbitmq client error: %v", err)
	}
	return nil
}
