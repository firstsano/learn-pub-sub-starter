package main

import (
	"fmt"

	"github.com/firstsano/learn-pub-sub-starter/internal/gamelogic"
	"github.com/firstsano/learn-pub-sub-starter/internal/pubsub"
	"github.com/firstsano/learn-pub-sub-starter/internal/routing"
)

func handlerLog(gl routing.GameLog) pubsub.AckType {
	defer fmt.Print("> ")

	if err := gamelogic.WriteLog(gl); err != nil {
		return pubsub.NackRequeue
	}

	return pubsub.Ack
}
