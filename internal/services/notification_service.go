package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type NotificationServiceInterface interface {
	CreateNotification(notificationType string, forUsername string, fromUsername string) error
	GetNotifications(username string) ([]models.NotificationResponseDTO, error, int)
	DeleteNotificationById(notificationId string, currentUsername string) (error, int)
}

type NotificationService struct {
	notificationRepository  repositories.NotificationRepositoryInterface
	PushSubscriptionService PushSubscriptionServiceInterface
}

// NewNotificationService can be used as a constructor to create a NotificationService "object"
func NewNotificationService(
	notificationRepository repositories.NotificationRepositoryInterface,
	puhSubscriptionService PushSubscriptionServiceInterface) *NotificationService {
	return &NotificationService{notificationRepository: notificationRepository, PushSubscriptionService: puhSubscriptionService}
}

// CreateNotification is a service function that creates a notification and pushes it to client if push service is registered
func (service *NotificationService) CreateNotification(notificationType string, forUsername string, fromUsername string) error {
	// Create notification and save to database
	newNotification := models.Notification{
		Id:               uuid.New(),
		NotificationType: notificationType,
		Timestamp:        time.Now(),
		ForUsername:      forUsername,
		FromUsername:     fromUsername,
	}
	err := service.notificationRepository.CreateNotification(&newNotification)

	// Send push message to client if push service is registered
	service.PushSubscriptionService.SendPushMessages(newNotification, forUsername)

	return err
}

// GetNotifications is a service function that gets all notifications for the current user
func (service *NotificationService) GetNotifications(username string) ([]models.NotificationResponseDTO, error, int) {
	notifications, err := service.notificationRepository.GetNotificationsByUsername(username)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	notificationResponseDTOs := make([]models.NotificationResponseDTO, 0)
	for _, notification := range notifications {
		notificationResponseDTO := models.NotificationResponseDTO{
			NotificationId:   notification.Id.String(),
			Timestamp:        notification.Timestamp,
			NotificationType: notification.NotificationType,
			User: &models.NotificationUserDTO{
				Username:          notification.FromUsername,
				Nickname:          notification.FromUser.Nickname,
				ProfilePictureUrl: notification.FromUser.ProfilePictureUrl,
			},
		}
		notificationResponseDTOs = append(notificationResponseDTOs, notificationResponseDTO)
	}

	return notificationResponseDTOs, nil, http.StatusOK
}

// DeleteNotificationById is a service function that deletes a notification by its id
func (service *NotificationService) DeleteNotificationById(notificationId string, currentUsername string) (error, int) {
	notification, err := service.notificationRepository.GetNotificationById(notificationId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.NotificationNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	if notification.ForUsername != currentUsername {
		return customerrors.DeleteNotificationForbidden, http.StatusForbidden
	}

	err = service.notificationRepository.DeleteNotificationById(notificationId)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}
