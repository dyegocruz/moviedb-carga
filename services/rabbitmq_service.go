package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

var rabbitDialFn = amqp.Dial
var rabbitChannelFactoryFn = func(conn *amqp.Connection) (amqpChanneler, error) {
	return conn.Channel()
}

// amqpChanneler abstracts the amqp.Channel methods used by RabbitMQService so
// a fake can be injected in tests.
type amqpChanneler interface {
	Qos(prefetchCount, prefetchSize int, global bool) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

type RabbitMQService struct {
	conn    *amqp.Connection
	channel amqpChanneler
}

func NewRabbitMQService(config Config) (*RabbitMQService, error) {
	if config == nil {
		config = DefaultConfig()
	}

	rabbitmqConfig := config.RabbitMQ()
	rabbitmqString := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitmqConfig.User, rabbitmqConfig.Password, rabbitmqConfig.Host, rabbitmqConfig.Port)

	conn, err := rabbitDialFn(rabbitmqString)
	if err != nil {
		return nil, err
	}

	channel, err := rabbitChannelFactoryFn(conn)
	if err != nil {
		return nil, err
	}

	return &RabbitMQService{conn: conn, channel: channel}, nil
}

func (r *RabbitMQService) Close() {
	if r == nil {
		return
	}
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

func (r *RabbitMQService) SetPrefetch(count int) error {
	return r.channel.Qos(count, 0, false)
}

func (r *RabbitMQService) PublishJSON(queueName string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if _, err := r.channel.QueueDeclare(queueName, false, false, false, false, nil); err != nil {
		return err
	}

	return r.channel.Publish("", queueName, false, false, amqp.Publishing{ContentType: "application/json", Body: body})
}

func (r *RabbitMQService) ConsumeJSON(queueName string, handler func([]byte) error) error {
	if _, err := r.channel.QueueDeclare(queueName, false, false, false, false, nil); err != nil {
		return err
	}

	msgs, err := r.channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			processDelivery(d, handler)
		}
		// msgs closes when the broker drops the channel/connection (restart,
		// network blip, heartbeat timeout, protocol error). Surface that as an
		// error instead of blocking forever, so the caller can exit and let the
		// process supervisor (Docker "restart: always") restore the consumer.
		log.Printf("Consumer channel closed unexpectedly, PID: %d", os.Getpid())
		done <- fmt.Errorf("rabbitmq consumer channel closed for queue %q", queueName)
	}()

	log.Println("Waiting for messages...")
	return <-done
}

// processDelivery runs handler for a single delivery, recovering from panics
// so that a single bad message can't take down the whole consumer/process.
func processDelivery(d amqp.Delivery, handler func([]byte) error) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("Recovered from panic while processing message: %v", rec)
			if err := d.Nack(false, true); err != nil {
				log.Printf("Error nacking message after panic: %s", err)
			}
		}
	}()

	if err := handler(d.Body); err != nil {
		log.Printf("Error processing message: %s", err)
	}
	if err := d.Ack(false); err != nil {
		log.Printf("Error acknowledging message : %s", err)
	} else {
		log.Printf("Acknowledged message")
	}
}
