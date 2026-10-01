package mapping

import (
	"github.com/retail-core/notification-service/internal/dtos"
	"github.com/retail-core/notification-service/internal/models"
)

func ToNotificationResponse(notification models.Notification) dtos.NotificationResponse {
	return dtos.NotificationResponse{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Message:   notification.Message,
		Type:      notification.Type,
		Title:     notification.Title,
		IsRead:    notification.IsRead,
		Data:      string(notification.Data),
		InventoryID: notification.InventoryID,
		StoreID: notification.StoreID,
		CreatedAt: notification.CreatedAt.String(),
	}
}
