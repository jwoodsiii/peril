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

	rChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error opening channel for connection: %v", err)
	}

	fmt.Println("Pause message sent!")

	gamelogic.PrintServerHelp()

	if err := pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, fmt.Sprintf("%s", routing.GameLogSlug), fmt.Sprintf("%s.*", routing.GameLogSlug), pubsub.Durable, handlerLogs(), amqp.Table{"x-dead-letter-exchange": pubsub.DLQ}); err != nil {
		log.Fatalf("could not subscribe to game_logs: %v", err)
	}

	for loop := true; loop; {
		input := gamelogic.GetInput()
		if input == nil {
			continue
		}
		firstWord := input[0]
		switch firstWord {
		case "pause":
			log.Printf("Sending pause message")
			if err := pubsub.PublishJSON(rChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "resume":
			log.Printf("Sending resume message...")
			if err := pubsub.PublishJSON(rChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false}); err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "quit":
			log.Printf("Exiting...\n")
			loop = false
		default:
			fmt.Println("Invalid command")
		}
	}

	//wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nClosing connection to server")
}
