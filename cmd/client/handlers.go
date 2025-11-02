package main

import (
	"fmt"

	"github.com/firstsano/learn-pub-sub-starter/internal/gamelogic"
	"github.com/firstsano/learn-pub-sub-starter/internal/pubsub"
	"github.com/firstsano/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState, _ *amqp.Channel) func(routing.PlayingState) pubsub.AckType {
	return func(state routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")

		gs.HandlePause(state)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(move gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		switch mo := gs.HandleMove(move); mo {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			_ = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				&gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.Player,
				},
			)
			return pubsub.NackRequeue
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			fmt.Println("error: unknown move outcome")
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, _, _ := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeYouWon, gamelogic.WarOutcomeDraw, gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		default:
			fmt.Println("error: unknown war outcome")
			return pubsub.NackDiscard
		}
	}
}
