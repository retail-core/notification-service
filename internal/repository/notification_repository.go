package repository

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/retail-core/notification-service/internal/models"
)

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *models.Notification) (*models.Notification, error)
	GetNotificationByUserID(ctx context.Context, userID uuid.UUID) ([]models.Notification, error)
	DeleteLowStockNotification(ctx context.Context, inventoryID uuid.UUID) error
	GetUnreadCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
}