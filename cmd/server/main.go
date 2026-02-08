package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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

	if err := pubsub.PublishJSON(rChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
		log.Printf("could not publish time: %v", err)
	}
	fmt.Println("Pause message sent!")

	//wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nClosing connection to server")
}
