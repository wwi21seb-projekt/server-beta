package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreatePost(post *models.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockPostRepository) GetPostsByUsername(username string, offset, limit int) ([]models.Post, int64, error) {
	args := m.Called(username, offset, limit)
	return args.Get(0).([]models.Post), args.Get(1).(int64), args.Error(2)
}

func (m *MockPostRepository) GetPostCountByUsername(username string) (int64, error) {
	args := m.Called(username)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostRepository) GetPostById(postId string) (models.Post, error) {
	args := m.Called(postId)
	return args.Get(0).(models.Post), args.Error(1)
}

func (m *MockPostRepository) GetPostsGlobalFeed(lastPost *models.Post, limit int) ([]models.Post, int64, error) {
	args := m.Called(lastPost, limit)
	return args.Get(0).([]models.Post), args.Get(1).(int64), args.Error(2)
}

func (m *MockPostRepository) GetPostsPersonalFeed(username string, lastPost *models.Post, limit int) ([]models.Post, int64, error) {
	args := m.Called(username, lastPost, limit)
	return args.Get(0).([]models.Post), args.Get(1).(int64), args.Error(2)
}

func (m *MockPostRepository) DeletePostById(postId string) error {
	args := m.Called(postId)
	return args.Error(0)
}
