package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	godotenv.Load()

	fmt.Println("Starting Peril server...")
	rabbitURL := os.Getenv("AMQP")
	log.Printf("rabbitURL: %s", rabbitURL)

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Error: %v connecting to rabbitq", err)
	}
	defer conn.Close()

	rChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error opening channel for connection: %v", err)
	}

	fmt.Println("Connected to Peril server...")
	uName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error defining username: %v", err)
	}
	fmt.Printf("Welcome to Peril, %s\n", uName)

	state := gamelogic.NewGameState(uName)

	if err := pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, uName), routing.PauseKey, pubsub.Transient, HandlerPause(state), amqp.Table{"x-dead-letter-exchange": pubsub.DLQ}); err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	// each client needs to subscribe to moves from other players
	if err := pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, state.GetUsername()), fmt.Sprintf("%s.*", routing.ArmyMovesPrefix), pubsub.Transient, HandlerMove(state), amqp.Table{"x-dead-letter-exchange": pubsub.DLQ}); err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}

	for loop := true; loop; {
		input := gamelogic.GetInput()
		if input == nil {
			log.Printf("Need to provide a command")
			continue
		}
		firstWord := input[0]
		switch firstWord {
		case "spawn":
			switch input[1] {
			case "americas", "europe", "africa", "asia", "antarctica", "australia":
			default:
				log.Printf("Invalid country input: %s", input[1])
			}
			switch input[2] {
			case "infantry", "cavalry", "artillery":
			default:
				log.Printf("Invalid unit input: %s", input[2])
			}
			log.Printf("Calling command spawn with args: %v", input[1:3])
			if err := state.CommandSpawn(input[1:]); err != nil {
				log.Fatalf("Error spawning unit: %v", err)
			}
		case "move":
			switch input[1] {
			case "americas", " europe", "africa", " asia", "antarctica", "australia":
				continue
			default:
				log.Printf("Invalid country input: %s", input[1])
			}
			armyMove, err := state.CommandMove(input[1:3])
			if err != nil {
				log.Printf("Error moving army: %v", err)
			}
			if err := pubsub.PublishJSON(rChan, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, uName), armyMove); err != nil {
				log.Printf("error: %v", err)
			}
			fmt.Printf("Moved %v units to %s\n", len(armyMove.Units), armyMove.ToLocation)

		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			log.Printf("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			loop = false
		default:
			log.Fatalf("Invalid command\n")
			continue
		}
	}
	//wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nClosing connection to server")

}
