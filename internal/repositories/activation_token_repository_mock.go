package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockActivationTokenRepository is a mock implementation of the ActivationTokenRepositoryInterface
type MockActivationTokenRepository struct {
	mock.Mock
}

func (m *MockActivationTokenRepository) CreateActivationToken(token *models.ActivationToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockActivationTokenRepository) CreateActivationTokenTx(token *models.ActivationToken, tx *gorm.DB) error {
	args := m.Called(token, tx)
	return args.Error(0)
}

func (m *MockActivationTokenRepository) FindTokenByUsername(username string) ([]models.ActivationToken, error) {
	args := m.Called(username)
	return args.Get(0).([]models.ActivationToken), args.Error(1)
}

func (m *MockActivationTokenRepository) FindActivationToken(username, token string) (*models.ActivationToken, error) {
	args := m.Called(username, token)
	if item, ok := args.Get(0).(*models.ActivationToken); ok {
		return item, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockActivationTokenRepository) DeleteActivationTokenByUsername(username string) error {
	args := m.Called(username)
	return args.Error(0)
}
