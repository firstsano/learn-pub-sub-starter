package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueue, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("failed to subscribe to queue %s: %w", queueName, err)
	}

	subs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}

	go readDeliveries(channel, subs, handler)

	return nil
}

func readDeliveries[T any](channel *amqp.Channel, deliveries <-chan amqp.Delivery, handler func(T)) {
	defer channel.Close()

	var message T
	fmt.Println("Reading messages...")

	for delivery := range deliveries {
		err := json.Unmarshal(delivery.Body, &message)
		if err != nil {
			log.Println("failed to unmarshal message:", err)
			continue
		}

		handler(message)
		_ = delivery.Ack(false)
	}
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
