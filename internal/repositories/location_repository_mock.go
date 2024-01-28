package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockLocationRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreateLocation(location *models.Location) error {
	args := m.Called(location)
	return args.Error(0)
}

func (m *MockPostRepository) GetLocationById(locationId string) (models.Location, error) {
	args := m.Called(locationId)
	return args.Get(0).(models.Location), args.Error(1)
}

func (m *MockPostRepository) DeleteLocationById(locationId string) error {
	args := m.Called(locationId)
	return args.Error(0)
}
