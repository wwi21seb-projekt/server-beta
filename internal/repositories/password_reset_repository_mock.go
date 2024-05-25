package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockPasswordResetRepository struct {
	mock.Mock
}

func (m *MockPasswordResetRepository) CreatePasswordResetToken(token *models.PasswordResetToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockPasswordResetRepository) FindPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	args := m.Called(token)
	return args.Get(0).(*models.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetRepository) DeletePasswordResetToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}
