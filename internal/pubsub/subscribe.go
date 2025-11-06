package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueue,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler, readDeliveryJSON)
}

func SubscribeGOB[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueue,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler, readDeliveryGOB)
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
		amqp.Table{
			"x-dead-letter-exchange": DeadLettersExchange,
		},
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	if err = ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueue,
	handler func(T) AckType,
	readDelivery func(delivery amqp.Delivery, message *T) error,
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

	go readDeliveries(channel, subs, handler, readDelivery)

	return nil
}

func readDeliveries[T any](
	channel *amqp.Channel,
	deliveries <-chan amqp.Delivery,
	handler func(T) AckType,
	readDelivery func(delivery amqp.Delivery, message *T) error,
) {
	defer channel.Close()
	defer fmt.Print("> ")

	var message T
	fmt.Println("Reading messages...")

	for delivery := range deliveries {
		err := readDelivery(delivery, &message)
		if err != nil {
			log.Println("failed to unmarshal message:", err)
			continue
		}

		switch ack := handler(message); ack {
		case Ack:
			log.Println("Acknowledge message")
			_ = delivery.Ack(false)
		case NackRequeue:
			log.Println("Requeue message")
			_ = delivery.Nack(false, true)
		case NackDiscard:
			log.Println("Discard message")
			_ = delivery.Nack(false, false)
		default:
			log.Println("Unknown ack type:", ack)
			log.Println("Skipping message")
		}
	}
}

func readDeliveryJSON[T any](delivery amqp.Delivery, message *T) error {
	return json.Unmarshal(delivery.Body, message)
}

func readDeliveryGOB[T any](delivery amqp.Delivery, message *T) error {
	buffer := bytes.NewBuffer(delivery.Body)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(message)
}
