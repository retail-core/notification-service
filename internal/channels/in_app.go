package channels

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/retail-core/notification-service/internal/dispatcher"
	"github.com/retail-core/notification-service/internal/logger"
	"github.com/retail-core/notification-service/internal/models"
	"github.com/retail-core/notification-service/internal/repository"
	"github.com/retail-core/notification-service/internal/ws"
	"go.uber.org/zap"
)

type NotificationInApp struct {
	UserID      string     `json:"user_id"`
	Message     string     `json:"message"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	InventoryID *uuid.UUID `json:"inventory_id,omitempty"`
	StoreID     *uuid.UUID `json:"store_id,omitempty"`
}

type InAppHandler struct {
	hub        *ws.Hub
	repository repository.NotificationRepository
}

func NewInAppHandler(hub *ws.Hub, repository repository.NotificationRepository) *InAppHandler {
	return &InAppHandler{hub: hub, repository: repository}
}

func (h *InAppHandler) Send(notification dispatcher.Notification) error {
	log := logger.L()

	var inventoryID *uuid.UUID
	var storeID *uuid.UUID

	if s := notification.Data["inventory_id"]; s != "" {
		u := uuid.FromStringOrNil(s)
		inventoryID = &u
	}

	if s := notification.Data["store_id"]; s != "" {
		u := uuid.FromStringOrNil(s)
		storeID = &u
	}

	req := NotificationInApp{
		UserID:      notification.Data["user_id"],
		Message:     notification.Data["message"],
		Title:       notification.Data["title"],
		Type:        notification.Data["type"],
		InventoryID: inventoryID,
		StoreID:     storeID,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Error("Failed to unmarshal in-app notification data", zap.Error(err))
		return err
	}

	client, ok := h.hub.Clients[req.UserID]
	if !ok {
		log.Warn("No active WebSocket client for user", zap.String("user_id", req.UserID))
		_ = __CreateNotificationInDB(context.Background(), h.repository, req)
		return nil
	}

	client.Send <- payload

	_ = __CreateNotificationInDB(context.Background(), h.repository, req)
	return nil
}

func __CreateNotificationInDB(ctx context.Context, repo repository.NotificationRepository, req NotificationInApp) error {
	notificationModel := &models.Notification{
		UserID:      uuid.FromStringOrNil(req.UserID),
		Message:     req.Message,
		Title:       req.Title,
		Type:        req.Type,
		IsRead:      false,
		IsActedUpon: false,
		InventoryID: req.InventoryID,
		Data:        nil,
		StoreID: req.StoreID,
	}

	if _, err := repo.CreateNotification(ctx, notificationModel); err != nil {
		logger.L().Error("Failed to create notification in repository", zap.Error(err))
		return err
	}

	logger.L().Info("In-app notification (LowStock) sent", zap.String("user_id", req.UserID))
	return nil
}
