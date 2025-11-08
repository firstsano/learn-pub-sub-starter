package main

import (
	"fmt"
	"log"
	"strconv"

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
		handlerPause(gs, channel),
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
		handlerMove(gs, channel),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Subscribing to war declarations")
	err = pubsub.SubscribeJSON(
		rabbit,
		routing.ExchangePerilTopic,
		routing.QueueWar,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gs, channel),
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
			if len(userInput[1]) == 0 {
				fmt.Println("Usage: spam <number_of_messages>")
				continue
			}

			msgsNumber, err := strconv.Atoi(userInput[1])
			if err != nil {
				fmt.Printf("Failed getting message number: %v\n", err)
				continue
			}

			for i := 0; i < msgsNumber; i++ {
				_ = publishLog(channel, gs.GetUsername(), gamelogic.GetMaliciousLog())
			}
			fmt.Printf("Spammed %d messages\n", msgsNumber)
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unrecognized command")
		}
	}
}
