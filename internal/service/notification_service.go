package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/retail-core/notification-service/internal/dtos"
	"github.com/retail-core/notification-service/internal/models"
)

type NotificationService interface {
	Create(ctx context.Context, storeID uuid.UUID, req dtos.CreateNotificationRequest) (*models.Notification, error)
	GetNotificationByUserID(ctx context.Context, userID uuid.UUID) ([]models.Notification, error)
	GetUnreadCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
}