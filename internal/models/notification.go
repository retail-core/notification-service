package models

import (
	"github.com/gofrs/uuid"
	"gorm.io/datatypes"
)

type NotificationType string
type NotificationChannel string

const (
	NotificationEmail NotificationChannel = "EMAIL"
	NotificationSMS   NotificationChannel = "SMS"
	NotificationPush  NotificationChannel = "PUSH"
)

const (
	NotificationLowStock    NotificationType = "INVENTORY_LOWSTOCK"
	NotificationExpiredItem NotificationType = "INVENTORY_EXPIRED"
	NotificationOther       NotificationType = "OTHER"
)

type Notification struct {
	Base

	UserID uuid.UUID `gorm:"type:varchar(200);not null"`

	Type string `gorm:"type:varchar(50);not null"`

	Title string `gorm:"type:varchar(255);not null"`

	Message string `gorm:"type:text;not null"`

	IsRead bool `gorm:"default:false"`

	IsActedUpon bool `gorm:"default:false"`

	InventoryID *uuid.UUID `gorm:"type:varchar(200);default:null"`

	Data datatypes.JSON `gorm:"type:jsonb"`
}

func (Notification) TableName() string {
	return "notifications"
}
