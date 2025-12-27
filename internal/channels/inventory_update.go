package channels

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/retail-core/notification-service/internal/dispatcher"
	"github.com/retail-core/notification-service/internal/logger"
	"github.com/retail-core/notification-service/internal/repository"
	"go.uber.org/zap"
)

type InventoryUpdateChannel struct{
	repository	 repository.NotificationRepository
}

func NewInventoryUpdateChannel(repository repository.NotificationRepository) *InventoryUpdateChannel {
	return &InventoryUpdateChannel{
		repository: repository,
	}
}

func (p *InventoryUpdateChannel) Send(notification dispatcher.Notification) error {
	log := logger.L()

	var inventoryID *uuid.UUID
	if s := notification.Data["inventory_id"]; s != "" {
		u := uuid.FromStringOrNil(s)
		inventoryID = &u
	}

	if err := p.repository.DeleteLowStockNotification(context.Background(), *inventoryID); err != nil {
		log.Error("Failed to delete low stock notification", zap.Error(err))
		return err
	}

	logger.L().Info("Low stock notification deleted successfully", zap.String("inventory_id", inventoryID.String()))
	return nil
}
