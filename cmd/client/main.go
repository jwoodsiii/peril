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
	fmt.Println("Connected to Peril server...")
	uName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error defining username: %v", err)
	}
	fmt.Printf("Welcome to Peril, %s\n", uName)

	state := gamelogic.NewGameState(uName)

	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, uName), routing.PauseKey, pubsub.Transient, handlerPause(state))

	// In the cmd/client package's main function, after creating the game state, replace your previous DeclareAndBind call with pubsub.SubscribeJSON. Use the following parameters:
	// The connection
	// The direct exchange (constant can be found in internal/routing)
	// A queue named pause.username where username is the username of the player
	// The routing key pause (constant can be found in internal/routing)
	// Transient queue type
	// The new handler we just created.

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
			state.CommandMove(input[1:3])

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
