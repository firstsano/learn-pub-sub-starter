package main

import (
	"time"

	"github.com/firstsano/learn-pub-sub-starter/internal/pubsub"
	"github.com/firstsano/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func publishLog(ch *amqp.Channel, username, message string) error {
	return pubsub.PublishGOB(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			CurrentTime: time.Now(),
			Message:     message,
			Username:    username,
		},
	)
}
