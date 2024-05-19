package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) CreateNotification(notification *models.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetNotificationsByUsername(username string) ([]models.Notification, error) {
	args := m.Called(username)
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetNotificationById(notificationId string) (models.Notification, error) {
	args := m.Called(notificationId)
	return args.Get(0).(models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) DeleteNotificationById(notificationId string) error {
	args := m.Called(notificationId)
	return args.Error(0)
}
