package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"moviedb/configs"
	"os"

	"github.com/streadway/amqp"
)

// RabbitMQ represents a connection to RabbitMQ
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQ creates a new RabbitMQ connection and channel
func NewRabbitMQ() (*RabbitMQ, error) {
  rabbitmqConfig := configs.GetRabbitMQEnv()
  rabbitmqString := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitmqConfig.User, rabbitmqConfig.Password, rabbitmqConfig.Host, rabbitmqConfig.Port)

  log.Println("RabbitMQ connection string: ", rabbitmqString)

	conn, err := amqp.Dial(rabbitmqString)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
	}, nil
}

// Close closes the RabbitMQ connection and channel
func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.conn.Close()
}

// SetPrefetch sets the prefetch count for the channel
func (r *RabbitMQ) SetPrefetch(count int) error {
	return r.channel.Qos(
		count, // prefetch count
		0,     // prefetch size (0 means no limit)
		false, // global (false applies to this channel only)
	)
}

// Publish sends a message to a queue
func (r *RabbitMQ) PublishJSON(queueName string, data interface{}) error {

  // Serialize the struct to JSON
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = r.channel.QueueDeclare(
		queueName, // queue name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	err = r.channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	return err
}

// ConsumeJSON consumes JSON messages from a queue and deserializes them into a struct
func (r *RabbitMQ) ConsumeJSON(queueName string, handler func([]byte) error) error {
	// Declare the queue
	_, err := r.channel.QueueDeclare(
		queueName, // queue name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	// Consume messages
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}

  stopChan := make(chan bool)

	// Start consuming messages
	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
      if err := handler(d.Body); err != nil {
        log.Printf("Error processing message: %s", err)
      }
			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message : %s", err)
			} else {
				log.Printf("Acknowledged message")
			}
		}
	}()

  fmt.Println("Waiting for messages...")
  <-stopChan

	return nil
}
