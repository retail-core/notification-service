package queue

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/retail-core/notification-service/internal/dispatcher"
	"github.com/retail-core/notification-service/internal/logger"
)

type RabbitConnection struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewRabbitConnection creates and returns a connected RabbitMQ instance.
func NewRabbitConnection(url string) *RabbitConnection {
	var err error
	var conn *amqp.Connection

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		logger.L().Info("failed to connect to RabbitMQ, Retrying to connect...")
		time.Sleep(5 * time.Second)
	} 

	failOnError(err, "failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "failed to open RabbitMQ channel")

	logger.L().Info("✅ connected to RabbitMQ successfully")

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


func (r *RabbitConnection) ConsumeNotifications(d *dispatcher.Dispatcher) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}

	channels := []dispatcher.NotChannel{
		dispatcher.EmailChannel,
		dispatcher.SMSChannel,
		dispatcher.PushChannel,
		dispatcher.InAppChannel,
		dispatcher.InventoryUpdateChannel,
	}

	for _, c := range channels {
		queueName := "notifications." + string(c)
		msgs, err := ch.Consume(queueName, "", false, false, false, false, nil) // auto-ack
		if err != nil {
			return err
		}

		go func(c string, deliveries <-chan amqp.Delivery) {
			for msg := range deliveries {
				d.Dispatch(msg.Body) // dispatcher figures out which handler to call
				msg.Ack(false) // TODO: handle error
			}
		}(string(c), msgs)
	}

	return nil
}

func (r *RabbitConnection) SetupNotificationBindings() error {
	exchange := "notifications"

	// Declare the exchange
	if err := r.ch.ExchangeDeclare(
		exchange,
		"direct", // simple routing by key: "email", "sms", "inapp"
		true,     // durable
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	// Channels your service listens to
	channels := []dispatcher.NotChannel{
		dispatcher.EmailChannel,
		dispatcher.SMSChannel,
		dispatcher.PushChannel,
		dispatcher.InAppChannel,
		dispatcher.InventoryUpdateChannel,
	}

	for _, channel := range channels {
		queueName := "notifications." + channel

		// Declare queue
		q, err := r.ch.QueueDeclare(
			string(queueName),
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			nil,
		)
		if err != nil {
			return err
		}

		// Bind queue to exchange using routing key = channel
		if err := r.ch.QueueBind(
			q.Name,
			string(channel),   // routing key (email/sms/inapp)
			exchange,  // exchange
			false,
			nil,
		); err != nil {
			return err
		}
	}

	return nil
}
