package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/retail-core/notification-service/internal/dtos"
	"github.com/retail-core/notification-service/internal/logger"
	"github.com/retail-core/notification-service/internal/models"
	"github.com/retail-core/notification-service/internal/repository"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type NotificationServiceImpl struct {
	repo repository.NotificationRepository
}

func NewNotificationServiceImpl(repo repository.NotificationRepository) *NotificationServiceImpl {
	return &NotificationServiceImpl{
		repo: repo,
	}
}

func (s *NotificationServiceImpl) Create(ctx context.Context, storeID uuid.UUID, req dtos.CreateNotificationRequest) (*models.Notification, error) {
	notification := &models.Notification{
		UserID:      req.UserID,
		Message:     req.Message,
		Type:        req.Type,
		Title:       req.Title,
		IsRead:      req.IsRead,
		IsActedUpon: req.IsActedUpon,
		Data:        datatypes.JSON([]byte(req.Data)),
	}

	createdNotification, err := s.repo.CreateNotification(ctx, notification)
	if err != nil {
		logger.L().Error("Failed to Create Notification", zap.Error(err))
		return nil, err
	}

	return createdNotification, nil
}

func (s *NotificationServiceImpl) GetNotificationByUserID(ctx context.Context, userID uuid.UUID) ([]models.Notification, error) {
	notifications, err := s.repo.GetNotificationByUserID(ctx, userID)

	if err != nil {
		logger.L().Error("Failed to Get Notification by UserID", zap.Error(err))
		return nil, err
	}

	return notifications, nil
}

func (s *NotificationServiceImpl) GetUnreadCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repo.GetUnreadCountByUserID(ctx, userID)
	if err != nil {
		logger.L().Error("Failed to Get Unread Notification Count by UserID", zap.Error(err))
		return 0, err
	}

	return count, nil
}

func (s *NotificationServiceImpl) MarkRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	err := s.repo.MarkRead(ctx, userID, notificationID)

	if err != nil {
		logger.L().Error("Failed to Mark Notification as Read", zap.Error(err))
		return err
	}

	return err
}

func (s *NotificationServiceImpl) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	err := s.repo.MarkAllRead(ctx, userID)

	if err != nil {
		logger.L().Error("Failed to Mark All Notifications as Read", zap.Error(err))
		return err
	}

	return nil
}