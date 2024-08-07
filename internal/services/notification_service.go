package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type NotificationServiceInterface interface {
	CreateNotification(notificationType string, forUsername string, fromUsername string) error
	GetNotifications(username string) (*models.NotificationsResponseDTO, *customerrors.CustomError, int)
	DeleteNotificationById(notificationId string, currentUsername string) (*customerrors.CustomError, int)
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
	if forUsername == fromUsername { // do not create notification if user is the same
		return nil
	}

	// Create notification and save to database
	newNotification := models.Notification{
		Id:               uuid.New(),
		NotificationType: notificationType,
		Timestamp:        time.Now(),
		ForUsername:      forUsername,
		FromUsername:     fromUsername,
	}
	err := service.notificationRepository.CreateNotification(&newNotification)

	// Get just created notification from database to get user metadata
	createdNotification, err := service.notificationRepository.GetNotificationById(newNotification.Id.String())
	if err != nil {
		return err
	}

	// Send push message to client if push service is registered
	notificationDto := models.NotificationRecordDTO{
		NotificationId:   createdNotification.Id.String(),
		Timestamp:        createdNotification.Timestamp,
		NotificationType: createdNotification.NotificationType,
		User:             utils.GenerateUserDTOFromUser(&createdNotification.FromUser),
	}
	service.PushSubscriptionService.SendPushMessages(&notificationDto, forUsername) // send push message in background

	return err
}

// GetNotifications is a service function that gets all notifications for the current user
func (service *NotificationService) GetNotifications(username string) (*models.NotificationsResponseDTO, *customerrors.CustomError, int) {
	// Retrieve notifications from database
	notifications, err := service.notificationRepository.GetNotificationsByUsername(username)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto
	notificationResponseDTOs := make([]models.NotificationRecordDTO, 0)
	for _, notification := range notifications {
		notificationResponseDTO := models.NotificationRecordDTO{
			NotificationId:   notification.Id.String(),
			Timestamp:        notification.Timestamp,
			NotificationType: notification.NotificationType,
			User:             utils.GenerateUserDTOFromUser(&notification.FromUser),
		}
		notificationResponseDTOs = append(notificationResponseDTOs, notificationResponseDTO)
	}

	responseDto := models.NotificationsResponseDTO{
		Records: notificationResponseDTOs,
	}

	return &responseDto, nil, http.StatusOK
}

// DeleteNotificationById is a service function that deletes a notification by its id
func (service *NotificationService) DeleteNotificationById(notificationId string, currentUsername string) (*customerrors.CustomError, int) {
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
