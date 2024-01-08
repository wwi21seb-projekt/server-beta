package repositories

import (
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionRepository ist eine Mock-Implementierung des SubscriptionRepositoryInterface
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) CreateSubscription(subscription *models.Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) DeleteSubscription(subscriptionId uuid.UUID) error {
	args := m.Called(subscriptionId)
	return args.Error(0)
}
