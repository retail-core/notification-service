package main

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/retail-core/notification-service/internal/channels/email"
	"github.com/retail-core/notification-service/internal/config"
	"github.com/retail-core/notification-service/internal/dispatcher"
	"github.com/retail-core/notification-service/internal/queue"
)

func main() {
	_ = godotenv.Load()

	cfg := config.LoadConfig()

	rabbit := queue.NewRabbitConnection(cfg.RabbitMQUrl)
	defer rabbit.Close()

	rabbit.DeclareQueue("notifications")

	messages, _ := rabbit.Consume("notifications")

	d := dispatcher.New()

	d.Register(dispatcher.EmailChannel, email.NewEmailHandler())

	go func() {
		for msg := range messages {
			d.Dispatch(msg.Body)
		}
	}()

	for {
		time.Sleep(10 * time.Second)
	}
}