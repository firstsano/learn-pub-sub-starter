package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const RabbitConnection = "amqp://guest:guest@localhost:5672/"

type SimpleQueue int

const (
	SimpleQueueDurable SimpleQueue = iota
	SimpleQueueTransient
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueue,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	q, err := ch.QueueDeclare(
		queueName,
		queueType == SimpleQueueDurable,
		queueType == SimpleQueueTransient,
		queueType == SimpleQueueTransient,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	if err = ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}
