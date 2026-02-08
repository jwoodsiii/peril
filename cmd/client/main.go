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

// Declare and bind a transient queue by creating and using a new function in the internal/pubsub package.
// I called mine DeclareAndBind with this signature:

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

	pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, uName), routing.PauseKey, "transient")
	//wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nClosing connection to server")

}
