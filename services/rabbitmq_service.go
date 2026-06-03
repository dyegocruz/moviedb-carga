package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type RabbitMQService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQService(config Config) (*RabbitMQService, error) {
	if config == nil {
		config = DefaultConfig()
	}

	rabbitmqConfig := config.RabbitMQ()
	rabbitmqString := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitmqConfig.User, rabbitmqConfig.Password, rabbitmqConfig.Host, rabbitmqConfig.Port)

	conn, err := amqp.Dial(rabbitmqString)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
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

	stopChan := make(chan bool)
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

	log.Println("Waiting for messages...")
	<-stopChan
	return nil
}
