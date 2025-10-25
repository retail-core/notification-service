package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/retail-core/notification-service/internal/config"
	"github.com/retail-core/notification-service/internal/queue"
)

func main() {
	_ = godotenv.Load()

	cfg := config.LoadConfig()

	rabbit := queue.NewRabbitConnection(cfg.RabbitMQUrl)
	defer rabbit.Close()

	rabbit.DeclareQueue("notifications")

	fmt.Printf("Notification service starting on port %s\n", cfg.Port)

	http.HandleFunc("/api/not-s/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}