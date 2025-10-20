package main

import (
	"fmt"

	"github.com/firstsano/learn-pub-sub-starter/internal/gamelogic"
	"github.com/firstsano/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(state routing.PlayingState) {
		defer fmt.Print("> ")

		gs.HandlePause(state)
	}
}
