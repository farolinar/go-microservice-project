package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         //durable
		false,        // auto deleted
		false,        // exchange internally
		false,        // no wait
		nil,          // args
	)
}

func DeclareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"normandy_queue", // name
		false,            // durable
		false,            // delete when unuse
		true,             // exclusive
		false,            // no wait
		nil,              // args
	)
}
