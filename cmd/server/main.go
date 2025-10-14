package main

import (
	"fmt"
	"log"

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

	channel, err := rabbit.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer channel.Close()

	fmt.Println("Connection to RabbitMQ established")
	fmt.Println("Starting Peril server...")
	gamelogic.PrintServerHelp()

gameLoop:
	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}

		switch userInput[0] {
		case "pause":
			fmt.Println("Pausing the game...")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
		case "resume":
			fmt.Println("Resuming the game...")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
		case "quit":
			fmt.Println("Quitting...")
			break gameLoop
		default:
			fmt.Println("Unrecognized command")
		}

	}

	fmt.Println("Received an interrupt, stopping services...")
}
