package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbit, err := amqp.Dial(pubsub.RabbitConnection)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbit.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		rabbit,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully declare and bind: ", queue.Name)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel

	fmt.Println("Starting Peril client...")
}
