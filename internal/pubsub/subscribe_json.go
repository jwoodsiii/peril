package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
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
			handler(message)
			delivery.Ack(false)
		}
	}()
	return nil
}
