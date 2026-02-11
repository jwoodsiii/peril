package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGameLog(msg, aggressor string, ch *amqp.Channel) error {
	fmt.Print("publishing game log...\n")
	gl := routing.GameLog{CurrentTime: time.Now().UTC(), Message: msg, Username: aggressor}
	if err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.GameLogSlug, aggressor), gl); err != nil {
		log.Printf("error publishing game logs: %v", err)
		return fmt.Errorf("Error publishing game log: %v", err)
	}
	return nil
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		res, winner, loser := gs.HandleWar(rw)
		var logMsg string
		switch res {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeDraw:
			logMsg = fmt.Sprintf("A war between %s and %s resulted in a draw.", winner, loser)
			if err := PublishGameLog(logMsg, rw.Attacker.Username, ch); err != nil {
				log.Printf("Error publishing game log, requeueing: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			logMsg = fmt.Sprintf("%s won a war against %s.", winner, loser)
			if err := PublishGameLog(logMsg, rw.Attacker.Username, ch); err != nil {
				log.Printf("Error publishing game log, requeueing: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			logMsg = fmt.Sprintf("%s won a war against %s.", loser, winner)
			if err := PublishGameLog(logMsg, rw.Attacker.Username, ch); err != nil {
				log.Printf("Error publishing game log, requeueing: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			log.Printf("Error determining war outcome")
			return pubsub.NackDiscard
		}
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return (pubsub.Ack)
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		mvRes := gs.HandleMove(move)
		log.Printf("detected move...")
		switch mvRes {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetPlayerSnap().Username), gamelogic.RecognitionOfWar{Attacker: move.Player, Defender: gs.GetPlayerSnap()}); err != nil {
				log.Printf("Error publishing message: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}
