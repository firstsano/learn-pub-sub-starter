package main

import (
	"fmt"
	"log"

	"github.com/firstsano/learn-pub-sub-starter/internal/gamelogic"
	"github.com/firstsano/learn-pub-sub-starter/internal/pubsub"
	"github.com/firstsano/learn-pub-sub-starter/internal/routing"
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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = pubsub.DeclareAndBind(
		rabbit,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatal(err)
	}

	gs := gamelogic.NewGameState(username)

	fmt.Println("Subscribing to pauses")
	err = pubsub.SubscribeJSON(
		rabbit,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Subscribing to moves")
	err = pubsub.SubscribeJSON(
		rabbit,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gs.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gs),
	)

	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}

		switch userInput[0] {
		case "spawn":
			if err = gs.CommandSpawn(userInput); err != nil {
				fmt.Printf("error spawnin unit: %v", err)
			}
		case "move":
			armyMove, err := gs.CommandMove(userInput)
			if err != nil {
				fmt.Printf("error moving unit: %v", err)
			}

			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+gs.GetUsername(),
				armyMove,
			)
			if err != nil {
				log.Fatalf("failed publishing unit move: %v", err)
			}

			fmt.Println("Move was published successfully")
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unrecognized command")
		}
	}
}
