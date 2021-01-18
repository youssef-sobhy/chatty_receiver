package internal

import (
	"os"

	"github.com/streadway/amqp"
)

const rabbitMQQueueName = "chatty_receiver.messages"

var rabbitMQURL = os.Getenv("RABBITMQ_URL")

// Publish func
func Publish(m []byte) error {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	// init RabbitMQ channel
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Init RabbitMQ queue
	args := make(amqp.Table)
	args["x-dead-letter-exchange"] = "chatty_receiver.messages-retry"

	q, err := ch.QueueDeclare(
		rabbitMQQueueName, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		args,              // arguments
	)
	if err != nil {
		return err
	}

	// Publish to RabbitMQ
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         m,
		})
	if err != nil {
		return err
	}

	return nil
}
