package main

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/retail-core/notification-service/internal/api"
	"github.com/retail-core/notification-service/internal/channels"
	"github.com/retail-core/notification-service/internal/channels/email"
	"github.com/retail-core/notification-service/internal/config"
	"github.com/retail-core/notification-service/internal/db"
	"github.com/retail-core/notification-service/internal/dispatcher"
	"github.com/retail-core/notification-service/internal/logger"
	"github.com/retail-core/notification-service/internal/queue"
	"github.com/retail-core/notification-service/internal/repository"
	"github.com/retail-core/notification-service/internal/ws"
	"go.uber.org/zap"
)

func main() {
	// Load .env if exists
	_ = godotenv.Load()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	logger.Init(cfg.Environment)
	log := logger.L()

	log.Info("Starting Notification Service")

	// Initialize database
	database, err := db.InitDB(*cfg)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Start WebSocket hub
	hub := ws.NewHub()
	go hub.Run()
	log.Info("WebSocket hub started")

	// Configure HTTP routes
	r := api.ConfigureRoutes(database, hub)

	// Setup RabbitMQ and dispatcher
	rabbit := queue.NewRabbitConnection(cfg.RabbitMQUrl)
	defer rabbit.Close()

	_dispatcher := dispatcher.New()
	_dispatcher.Register(dispatcher.EmailChannel, email.NewPostmarkHandler(cfg))
	_dispatcher.Register(dispatcher.InAppChannel, channels.NewInAppHandler(hub, repository.NewGormNotificationRepository(database)))
	_dispatcher.Register(dispatcher.InventoryUpdateChannel, channels.NewInventoryUpdateChannel(repository.NewGormNotificationRepository(database)))

	err =rabbit.SetupNotificationBindings()
	if err != nil {
		log.Fatal("Failed to setup RabbitMQ notification bindings", zap.Error(err))
	}

	err = rabbit.ConsumeNotifications(_dispatcher)
	if err != nil {
		log.Fatal("Failed to consume notifications", zap.Error(err))
	}

	log.Info("Notification Service initialized, starting HTTP server...", zap.String("port", cfg.Port))
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}
