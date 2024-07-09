package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type NotificationRepositoryInterface interface {
	CreateNotification(notification *models.Notification) error
	GetNotificationsByUsername(username string) ([]models.Notification, error)
	GetNotificationById(notificationId string) (models.Notification, error)
	DeleteNotificationById(notificationId string) error
}

type NotificationRepository struct {
	DB *gorm.DB
}

// NewNotificationRepository can be used as a constructor to create a NotificationRepository "object"
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{DB: db}
}

func (repo *NotificationRepository) CreateNotification(notification *models.Notification) error {
	err := repo.DB.Create(notification).Error
	return err
}

func (repo *NotificationRepository) GetNotificationsByUsername(username string) ([]models.Notification, error) {
	var notifications []models.Notification
	err := repo.DB.
		Where("for_username = ?", username).
		Order("timestamp desc").
		Preload("FromUser").
		Preload("FromUser.Image").
		Find(&notifications).Error
	return notifications, err
}

func (repo *NotificationRepository) GetNotificationById(notificationId string) (models.Notification, error) {
	var notification models.Notification
	err := repo.DB.
		Where("id = ?", notificationId).
		First(&notification).
		Preload("FromUser").
		Preload("FromUser.Image").Error
	return notification, err
}

func (repo *NotificationRepository) DeleteNotificationById(notificationId string) error {
	err := repo.DB.Where("id = ?", notificationId).Delete(&models.Notification{}).Error
	return err
}
