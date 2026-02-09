package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func HandlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return (pubsub.Ack)
	}
}

func HandlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		mvRes := gs.HandleMove(move)
		log.Printf("detected move...")
		switch mvRes {
		case gamelogic.MoveOutComeSafe, gamelogic.MoveOutcomeMakeWar:
			return (pubsub.Ack)
		default:
			return (pubsub.NackDiscard)
		}
	}
}
