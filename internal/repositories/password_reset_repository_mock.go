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

func (m *MockPasswordResetRepository) FindPasswordResetToken(username string, token string) (*models.PasswordResetToken, error) {
	args := m.Called(username, token)
	if args.Get(0) != nil {
		return args.Get(0).(*models.PasswordResetToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPasswordResetRepository) DeletePasswordResetToken(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
