package pubsub

import "github.com/firstsano/learn-pub-sub-starter/internal/routing"

// TODO: refactor
const (
	RabbitConnection    = "amqp://guest:guest@localhost:5672/"
	DeadLettersExchange = routing.ExchangePerilDeadLetters
)

type SimpleQueue int
type AckType int

const (
	SimpleQueueDurable SimpleQueue = iota
	SimpleQueueTransient
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)
