package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](
	ch *amqp.Channel,
	exchange,
	key string,
	val T,
) error {
	var dat bytes.Buffer
	enc := gob.NewEncoder(&dat)
	if err := enc.Encode(val); err != nil {
		return fmt.Errorf("Error encoding value: %v", err)
	}
	if err := ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/gob", Body: dat.Bytes()}); err != nil {
		log.Printf("error publishing gob: %v", err)
		return fmt.Errorf("error publishing gob: %v", err)
	}
	return nil
}
