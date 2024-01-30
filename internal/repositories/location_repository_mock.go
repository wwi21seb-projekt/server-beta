package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockLocationRepository struct {
	mock.Mock
}

func (m *MockLocationRepository) CreateLocation(location *models.Location) error {
	args := m.Called(location)
	return args.Error(0)
}
