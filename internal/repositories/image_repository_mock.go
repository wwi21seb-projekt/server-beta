package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockImageRepository struct {
	mock.Mock
}

func (m *MockImageRepository) CreateImage(image *models.Image) error {
	args := m.Called(image)
	return args.Error(0)
}

func (m *MockImageRepository) DeleteImage(imageUrl string) error {
	args := m.Called(imageUrl)
	return args.Error(0)
}

func (m *MockImageRepository) GetImageMetadata(imageUrl string) (*models.Image, error) {
	args := m.Called(imageUrl)
	return args.Get(0).(*models.Image), args.Error(1)
}
