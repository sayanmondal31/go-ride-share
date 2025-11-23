package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trip"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		log.Fatalf("Failed to connect to rabbitmq")
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()

	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %v: ", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	if err = rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to setup exchanges: %v", err)
	}

	return rmq, nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	log.Printf("Publishing message with routing key: %s", routingKey)
	return r.Channel.PublishWithContext(ctx,
		TripExchange, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         jsonMsg,
			DeliveryMode: amqp.Persistent,
		},
	)

}

// interface
type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	// Set prefetch count to 1 for fair dispatch
	// This tells RabbitMQ not to give more than one message to a service at a time
	// The worker will only get the next message after it has acknowledge the previous one

	err := r.Channel.Qos(
		1,     // prefetchCount: Limit to 1 unacknowledge to service at a time.
		0,     // This tells RabbitMQ not to give more than one message to service at a time.
		false, // global: Apply prefetch to each consumer individually
	)

	if err != nil {
		return fmt.Errorf("failed to set Qos: %v", err)
	}

	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	if err != nil {
		return err
	}

	ctx := context.Background()

	go func() {
		for msg := range msgs {
			// log.Printf("Received a message: %s", msg.Body)

			if err := handler(ctx, msg); err != nil {
				log.Fatalf("failed to handle the message : %v", err)

				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("ERROR: failed to Nack message: %v", nackErr)
				}

				continue
			}

			// Only Ack if the handler succeeds
			_ = msg.Ack(false)
		}
	}()

	return nil
}

func (r *RabbitMQ) setupExchangesAndQueues() error {

	err := r.Channel.ExchangeDeclare(
		TripExchange, // name
		"topic",      //type
		true,         // durable
		false,        // auto-deleted
		false,        //internal
		false,        //no-wait
		nil,          //arguments

	)

	if err = r.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{
			contracts.TripEventCreated, contracts.TripEventDriverNotInterested,
		},
		TripExchange,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string) error {
	q, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durable messages will persists
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messageTypes {
		if err := r.Channel.QueueBind(
			q.Name,   // queue name
			msg,      // routing key
			exchange, // exchange
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue to %s: %v", queueName, err)
		}
	}

	return nil

}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}

	if r.Channel != nil {
		r.Channel.Close()
	}
}
