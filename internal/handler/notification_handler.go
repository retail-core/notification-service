package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/retail-core/notification-service/internal/dtos"
	"github.com/retail-core/notification-service/internal/errors"
	"github.com/retail-core/notification-service/internal/httpx"
	"github.com/retail-core/notification-service/internal/mapping"
	"github.com/retail-core/notification-service/internal/service"
)

type NotificationHandler struct {
	Service service.NotificationService
}

func NewNotificationHandler(s service.NotificationService) *NotificationHandler {
	return &NotificationHandler{Service: s}
}

func (h *NotificationHandler) GetNotificationsByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.FromString(userIDStr)

	if err != nil {
		httpx.WriteError(w, errors.BadRequest("invalid user ID"))
		return
	}

	notifications, err := h.Service.GetNotificationByUserID(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	var response = make([]dtos.NotificationResponse, 0, len(notifications))

	for _, notification := range notifications {
		response = append(response, mapping.ToNotificationResponse(notification))
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.FromString(userIDStr)

	if err != nil {
		httpx.WriteError(w, errors.BadRequest("invalid user ID"))
		return
	}

	count, err := h.Service.GetUnreadCountByUserID(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	response := map[string]int64{
		"count": count,
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	notificationIDStr := chi.URLParam(r, "notification_id")

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		httpx.WriteError(w, errors.BadRequest("invalid user ID"))
		return
	}

	notificationID, err := uuid.FromString(notificationIDStr)
	if err != nil {
		httpx.WriteError(w, errors.BadRequest("invalid notification ID"))
		return
	}

	err = h.Service.MarkRead(r.Context(), userID, notificationID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.FromString(userIDStr)

	if err != nil {
		httpx.WriteError(w, errors.BadRequest("invalid user ID"))
		return
	}

	err = h.Service.MarkAllRead(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}