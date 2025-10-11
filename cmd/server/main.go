package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitConn := "amqp://guest:guest@localhost:5672/"
	rabbit, err := amqp.Dial(rabbitConn)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbit.Close()

	channel, err := rabbit.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer channel.Close()

	err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to RabbitMQ")
	fmt.Println("Starting Peril server...")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Received an interrupt, stopping services...")
}
