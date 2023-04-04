package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	conn *amqp.Connection
}

func NewEventEmitter(conn *amqp.Connection) (Emitter, error) {
	e := Emitter{
		conn: conn,
	}

	err := e.setup()
	if err != nil {
		return Emitter{}, err
	}

	return e, nil
}

func (e *Emitter) setup() error {
	ch, err := e.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	return DeclareExchange(ch)
}

func (e *Emitter) Push(event, severity string) error {
	ch, err := e.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	err = ch.Publish(
		"logs_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "plain/text",
			Body:        []byte(event),
		},
	)
	if err != nil {
		return err
	}

	return nil
}
