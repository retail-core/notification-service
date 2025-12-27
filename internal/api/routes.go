package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/retail-core/notification-service/internal/handler"
	"github.com/retail-core/notification-service/internal/repository"
	"github.com/retail-core/notification-service/internal/service"
	"github.com/retail-core/notification-service/internal/ws"
	"gorm.io/gorm"
)

func ConfigureRoutes(database *gorm.DB, hub *ws.Hub) http.Handler {

	notificationRepo := repository.NewGormNotificationRepository(database)

	notificationService := service.NewNotificationServiceImpl(notificationRepo)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Route("/v1", func(v1 chi.Router) {
		v1.Get("/users/{user_id}/notifications", notificationHandler.GetNotificationsByUserID)
		v1.Get("/users/{user_id}/notifications/unread-count", notificationHandler.GetUnreadCount)
		v1.Post("/users/{user_id}/notifications/{notification_id}/read", notificationHandler.MarkRead)
		v1.Post("/users/{user_id}/notifications/read-all", notificationHandler.MarkAllRead)
		v1.Get("/ws", ws.WebSocketHandler(hub))
	})

	return r
}
