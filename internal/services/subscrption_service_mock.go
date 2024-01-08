package services

import (
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionService ist eine Mock-Implementierung des SubscriptionServiceInterface
type MockSubscriptionService struct {
	mock.Mock
}

func (m *MockSubscriptionService) PostSubscription(follower, following string) (*models.SubscriptionPostResponseDTO, *customerrors.CustomError, int) {
	args := m.Called(follower, following)
	return args.Get(0).(*models.SubscriptionPostResponseDTO), args.Get(1).(*customerrors.CustomError), args.Int(2)
}

func (m *MockSubscriptionService) DeleteSubscription(subscriptionId uuid.UUID) (*models.SubscriptionDeleteResponseDTO, *customerrors.CustomError, int) {
	args := m.Called(subscriptionId)
	return args.Get(0).(*models.SubscriptionDeleteResponseDTO), args.Get(1).(*customerrors.CustomError), args.Int(2)
}
