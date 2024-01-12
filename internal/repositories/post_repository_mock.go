package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreatePost(post *models.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockPostRepository) FindPostsByUser(username string, offset, limit int) ([]models.Post, error) {
	args := m.Called(username, offset, limit)
	return args.Get(0).([]models.Post), args.Error(1)
}
