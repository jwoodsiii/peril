package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}
	delivery, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to consume queue: %w", err)
	}
	go func() {
		for delivery := range delivery {
			var message T
			if err := json.Unmarshal(delivery.Body, &message); err != nil {
				fmt.Printf("failed to unmarshal message: %v\n", err)
				delivery.Nack(false, false)
				continue
			}
			ack := handler(message)
			switch ack {
			case Ack:
				fmt.Print("Message Ack\n")
				delivery.Ack(false)
			case NackRequeue:
				fmt.Print("Message Nack requeue\n")
				delivery.Nack(false, true)
			case NackDiscard:
				fmt.Print("Message Nack discard\n")
				delivery.Nack(false, false)
			}
		}
	}()
	return nil
}
