package queue

import (
	"errors"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/retail-core/notification-service/internal/dispatcher"
	"github.com/retail-core/notification-service/internal/logger"
	"go.uber.org/zap"
)

type RabbitConnection struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	url  string
}

func NewRabbitConnection(url string) *RabbitConnection {
	r := &RabbitConnection{url: url}
	r.connect()
	return r
}

func (r *RabbitConnection) connect() {
	var err error
	for i := 0; i < 10; i++ {
		r.conn, err = amqp.Dial(r.url)
		if err == nil {
			break
		}
		logger.L().Info("failed to connect to RabbitMQ, retrying...")
		time.Sleep(5 * time.Second)
	}
	failOnError(err, "failed to connect to RabbitMQ")

	r.ch, err = r.conn.Channel()
	failOnError(err, "failed to open RabbitMQ channel")

	logger.L().Info("✅ connected to RabbitMQ successfully")
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
				if err := dispatchWithRetry(d, msg.Body); err != nil {
					logger.L().Error("dispatch failed, dead-lettering", zap.String("channel", c), zap.Error(err))
					msg.Nack(false, false) // no requeue → DLX/DLQ
					continue
				}
				msg.Ack(false)
			}

		}(string(c), msgs)
	}

	return nil
}

func (r *RabbitConnection) SetupNotificationBindings() error {
	exchange := "notifications"

	if err := r.ch.ExchangeDeclare(
		exchange,
		"topic", // simple routing by key: "email", "sms", "inapp"
		true,    // durable
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := r.ch.ExchangeDeclare("notifications.dlx", "fanout", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := r.ch.QueueDeclare("notifications.dlq", true, false, false, false, nil); err != nil {
		return err
	}
	if err := r.ch.QueueBind("notifications.dlq", "", "notifications.dlx", false, nil); err != nil {
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
			logger.L().Error("Failed to declare queue", zap.String("queue", string(queueName)), zap.Error(err))
			return err
		}

		// Bind queue to exchange using routing key = channel
		if err := r.ch.QueueBind(
			q.Name,
			string(channel), // routing key (email/sms/inapp)
			exchange,        // exchange
			false,
			nil,
		); err != nil {
			logger.L().Error("Failed to bind queue to exchange", zap.String("queue", q.Name), zap.String("exchange", exchange), zap.String("routing_key", string(channel)), zap.Error(err))
			return err
		}
	}

	logger.L().Info("✅ RabbitMQ notification bindings set up successfully DDL")
	return nil
}

// call this AFTER SetupNotificationBindings, wraps ConsumeNotifications with
// auto-reconnect whenever the connection drops
func (r *RabbitConnection) ConsumeNotificationsWithReconnect(d *dispatcher.Dispatcher) {
	for {
		closeCh := make(chan *amqp.Error)
		r.conn.NotifyClose(closeCh)

		if err := r.ConsumeNotifications(d); err != nil {
			logger.L().Error("failed to start consuming, retrying in 5s", zap.Error(err))
			time.Sleep(5 * time.Second)
			r.connect()
			if err := r.SetupNotificationBindings(); err != nil {
				logger.L().Error("failed to rebind after reconnect", zap.Error(err))
			}
			continue
		}

		// block here until the connection dies
		reason := <-closeCh
		logger.L().Error("🔌 RabbitMQ connection lost, reconnecting...", zap.Any("reason", reason))

		time.Sleep(5 * time.Second)
		r.connect()
		if err := r.SetupNotificationBindings(); err != nil {
			logger.L().Error("failed to rebind after reconnect", zap.Error(err))
		}
	}
}

const (
	maxAttempts = 3
	baseBackoff = 2 * time.Second
)

func dispatchWithRetry(d *dispatcher.Dispatcher, body []byte) error {
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = d.Dispatch(body); err == nil {
			return nil
		}
		var perm *dispatcher.PermanentError
		if errors.As(err, &perm) {
			return err
		}
		if attempt < maxAttempts {
			logger.L().Warn("dispatch failed, retrying", zap.Int("attempt", attempt), zap.Error(err))
			time.Sleep(baseBackoff * time.Duration(1<<(attempt-1)))
		}
	}
	return err
}
