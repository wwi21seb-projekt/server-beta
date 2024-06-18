package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockImageRepository struct {
	mock.Mock
}

func (m *MockImageRepository) GetImageById(id string) (*models.Image, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Image), args.Error(1)
}

func (m *MockImageRepository) DeleteImageById(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
