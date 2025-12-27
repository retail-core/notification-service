package repository

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/retail-core/notification-service/internal/models"
	"gorm.io/gorm"
)

type GormNotificationRepository struct {
	DB *gorm.DB
}

func NewGormNotificationRepository(db *gorm.DB) *GormNotificationRepository {
	return &GormNotificationRepository{DB: db}
}

func (r *GormNotificationRepository) CreateNotification(ctx context.Context, notification *models.Notification) (*models.Notification, error) {

	// check if there's already a notification with type = INVENTORY_LOWSTOCK for the same user and inventory ID
	// if there is, pls delete it first before creating a new one

	// Delete existing notification with type = INVENTORY_LOWSTOCK for the same user and inventory ID
	if notification.Type == string(models.NotificationLowStock) && notification.InventoryID != nil {
		if err := r.DB.WithContext(ctx).
			Where("user_id = ? AND type = ? AND inventory_id = ?", notification.UserID, notification.Type, notification.InventoryID).
			Delete(&models.Notification{}).Error; err != nil {
			return nil, fmt.Errorf("failed to delete existing notification: %w", err)
		}
	}

	if result := r.DB.WithContext(ctx).Create(notification); result.Error != nil {
		return nil, fmt.Errorf("failed to create notification: %w", result.Error)
	}
	return notification, nil
}

func (r *GormNotificationRepository) GetNotificationByUserID(ctx context.Context, userID uuid.UUID) ([]models.Notification, error) {
	notifications := []models.Notification{}

	if result := r.DB.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&notifications); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return []models.Notification{}, nil
		}
		return nil, fmt.Errorf("failed to find notification by ID %s: %w", userID, result.Error)
	}
	return notifications, nil
}

func (r *GormNotificationRepository) DeleteLowStockNotification(ctx context.Context, inventoryID uuid.UUID) error {
	if err := r.DB.WithContext(ctx).
		Where("type = ? AND inventory_id = ?", string(models.NotificationLowStock), inventoryID).
		Delete(&models.Notification{}).Error; err != nil {
		return fmt.Errorf("failed to delete low stock notifications for inventory ID %s: %w", inventoryID, err)
	}
	return nil
}

func (r *GormNotificationRepository) GetUnreadCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64

    err := r.DB.WithContext(ctx).
        Model(&models.Notification{}).
        Where("user_id = ? AND is_read = false", userID).
        Count(&count).Error

    if err != nil {
        return 0, err
    }

    return count, nil
}

func (r *GormNotificationRepository) MarkRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	result := r.DB.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error

	if result != nil {
		return fmt.Errorf("failed to mark notification %s as read for user %s: %w", notificationID, userID, result)
	}

	return nil
}

func (r *GormNotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	err := r.DB.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error

	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read for user %s: %w", userID, err)
	}

	return nil
}