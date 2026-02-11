package pubsub

import (
	"bytes"
	"encoding/gob"
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

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	table amqp.Table,
) error {
	return subscribe[T](conn, exchange, queueName, key, queueType, handler, table, func(data []byte) (T, error) {
		buf := bytes.NewBuffer(data)
		decoder := gob.NewDecoder(buf)
		var target T
		err := decoder.Decode(&target)
		return target, err
	})
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType, table)
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
			var out bytes.Buffer
			out.Write(delivery.Body)
			if err := gob.NewDecoder(&out).Decode(&message); err != nil {
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
	table amqp.Table,
) error {
	return subscribe[T](conn, exchange, queueName, key, queueType, handler, table, func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	})
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	table amqp.Table,
	unmarshaller func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType, table)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}
	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}
	go func() {
		defer ch.Close()
		for msg := range msgs {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			switch handler(target) {
			case Ack:
				msg.Ack(false)
			case NackDiscard:
				msg.Nack(false, false)
			case NackRequeue:
				msg.Nack(false, true)
			}
		}
	}()
	return nil
}
