package rabbitmq

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"os"
)

// RabbitMQ broker client
type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	log  *slog.Logger
}

//func RabbitURL() (string, error) {
//	u := &url.URL{
//		Scheme: "amqp",
//		User:   url.UserPassword(os.Getenv("RABBITMQ_DEFAULT_USER"), os.Getenv("RABBITMQ_DEFAULT_PASS")),
//		Host:   fmt.Sprintf("%s:%s", os.Getenv("RABBITMQ_HOST"), os.Getenv("RABBITMQ_PORT")),
//		Path:   "/", // Go сам при String() превращает "/" в "%2F"
//	}
//	return u.String(), nil
//}

func NewRabbitMQ(log *slog.Logger) (*RabbitMQ, error) {
	rabbitConnUrl := os.Getenv("RABBITMQ_URL")
	conn, err := amqp.Dial(rabbitConnUrl)
	if err != nil {
		log.Error("Failed to connect to RabbitMQ: " + err.Error())
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel: " + err.Error())
		return nil, err
	}
	return &RabbitMQ{conn: conn, ch: ch, log: log}, nil
}

// Send enqueues order as json object to specified queue in RabbitMQ
func (r *RabbitMQ) Send(ctx context.Context, order interface{}, queueName string) error {
	// Ensures queue exists
	q, err := r.ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.log.Error("Failed to declare a queue: " + err.Error())
		return err
	}
	body, err := json.Marshal(order)
	if err != nil {
		r.log.Error("Failed to marshal order: " + err.Error())
		return err
	}
	err = r.ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		r.log.Error("Failed to publish a message: " + err.Error())
		return err
	}
	return nil
}
