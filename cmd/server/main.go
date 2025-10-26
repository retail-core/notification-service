package main

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/retail-core/notification-service/internal/channels/email"
	"github.com/retail-core/notification-service/internal/config"
	"github.com/retail-core/notification-service/internal/dispatcher"
	"github.com/retail-core/notification-service/internal/logger"
	"github.com/retail-core/notification-service/internal/queue"
)

func main() {
	_ = godotenv.Load()

	cfg := config.LoadConfig()

	logger.Init(cfg.Environment)

	rabbit := queue.NewRabbitConnection(cfg.RabbitMQUrl)
	defer rabbit.Close()

	rabbit.DeclareQueue("notifications")

	messages, _ := rabbit.Consume("notifications")

	d := dispatcher.New()

	d.Register(dispatcher.EmailChannel, email.NewPostmarkHandler(cfg))

	logger.L().Info("Notification Service is started")

	go func() {
		for msg := range messages {
			d.Dispatch(msg.Body)
		}
	}()

	for {
		time.Sleep(10 * time.Second)
	}
}
