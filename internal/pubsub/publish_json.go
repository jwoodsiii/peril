package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	dat, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("Error marshalling value: %v", err)
	}
	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/json", Body: dat})
	return nil
}
