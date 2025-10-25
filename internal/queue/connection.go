package queue

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitConnection struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewRabbitConnection creates and returns a connected RabbitMQ instance.
func NewRabbitConnection(url string) *RabbitConnection {
	conn, err := amqp.Dial(url)
	failOnError(err, "failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "failed to open RabbitMQ channel")

	log.Println("✅ connected to RabbitMQ successfully")

	return &RabbitConnection{conn: conn, ch: ch}
}

func (r *RabbitConnection) DeclareQueue(name string) amqp.Queue {
	q, err := r.ch.QueueDeclare(
		name,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare queue %s: %v", name, err)
	}

	log.Printf("📦 Queue declared: %s", q.Name)
	return q
}

func (r *RabbitConnection) Consume(queue string) (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		queue,
		"",
		true,  // auto-ack for now
		false, // exclusive
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// Close closes the connection cleanly
func (r *RabbitConnection) Close() {
	log.Println("🧹 Closing RabbitMQ connection...")
	r.ch.Close()
	r.conn.Close()
}


func failOnError(err error, msg string) {
  if err != nil {
    log.Panicf("%s: %s", msg, err)
  }
}