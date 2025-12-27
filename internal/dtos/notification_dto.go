package dtos

import (
	"github.com/gofrs/uuid"
)

type NotificationResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Message     string    `json:"message"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	IsRead      bool      `json:"is_read"`
	Data        string    `json:"data"`
	InventoryID *uuid.UUID `json:"inventory_id,omitempty"`
	CreatedAt   string    `json:"created_at"`
}

type CreateNotificationRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	Message     string    `json:"message"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	IsRead      bool      `json:"is_read"` // default false
	IsActedUpon bool      `json:"is_acted_upon"`
	Data        string    `json:"data"`
}
