package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

const DLQ = "peril_dlx"

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	table amqp.Table,
) (*amqp.Channel, amqp.Queue, error) {

	tChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	durable := false
	autoDelete := false
	exclusive := false
	noWait := false
	args := table
	switch queueType {
	case Durable:
		durable = true
	case Transient:
		autoDelete = true
		exclusive = true
	}
	queue, err := tChan.QueueDeclare(queueName, durable, autoDelete, exclusive, noWait, args)
	if err := tChan.QueueBind(queueName, key, exchange, noWait, args); err != nil {
		log.Printf("Error binding queue: %v", err)
		return nil, amqp.Queue{}, err
	}
	return tChan, queue, nil
}
