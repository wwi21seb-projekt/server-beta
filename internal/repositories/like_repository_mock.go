package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockLikeRepository struct {
	mock.Mock
}

func (m *MockLikeRepository) CreateLike(like *models.Like) error {
	args := m.Called(like)
	return args.Error(0)
}

func (m *MockLikeRepository) DeleteLike(likeId string) error {
	args := m.Called(likeId)
	return args.Error(0)
}

func (m *MockLikeRepository) FindLike(postId string, currentUsername string) (*models.Like, error) {
	args := m.Called(postId, currentUsername)
	return args.Get(0).(*models.Like), args.Error(1)
}

func (m *MockLikeRepository) CountLikes(postId string) (int64, error) {
	args := m.Called(postId)
	return args.Get(0).(int64), args.Error(1)
}
