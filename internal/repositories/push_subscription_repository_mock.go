package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockPushSubscriptionRepository struct {
	mock.Mock
}

func (m *MockPushSubscriptionRepository) CreatePushSubscription(subscription *models.PushSubscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func (m *MockPushSubscriptionRepository) GetPushSubscriptionsByUsername(username string) ([]models.PushSubscription, error) {
	args := m.Called(username)
	return args.Get(0).([]models.PushSubscription), args.Error(1)
}

func (m *MockPushSubscriptionRepository) DeletePushSubscriptionById(subscriptionId string) error {
	args := m.Called(subscriptionId)
	return args.Error(0)
}
