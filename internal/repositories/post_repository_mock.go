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

func (m *MockPostRepository) FindPostsByUsernameCount(username string) (int64, error) {
	args := m.Called(username)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostRepository) FindPostsByUsername(username string, offset, limit int) ([]models.Post, error) {
	args := m.Called(username, offset, limit)
	return args.Get(0).([]models.Post), args.Error(1)
}
