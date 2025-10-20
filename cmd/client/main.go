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
	err = pubsub.SubscribeJSON(
		rabbit,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatal(err)
	}

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
			if _, err = gs.CommandMove(userInput); err != nil {
				fmt.Printf("error moving unit: %v", err)
			}
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
