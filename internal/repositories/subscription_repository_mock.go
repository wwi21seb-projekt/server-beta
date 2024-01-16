package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) CreateSubscription(subscription *models.Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) DeleteSubscription(subscriptionId string) error {
	args := m.Called(subscriptionId)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetSubscriptionByUsernames(follower, following string) (*models.Subscription, error) {
	args := m.Called(follower, following)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetSubscriptionById(subscriptionId string) (*models.Subscription, error) {
	args := m.Called(subscriptionId)
	return args.Get(0).(*models.Subscription), args.Error(1)
}