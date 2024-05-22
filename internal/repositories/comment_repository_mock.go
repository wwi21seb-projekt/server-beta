package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) CreateComment(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetCommentsByPostId(postId string, offset, limit int) ([]models.Comment, int64, error) {
	args := m.Called(postId, offset, limit)
	return args.Get(0).([]models.Comment), args.Get(1).(int64), args.Error(2)
}

func (m *MockCommentRepository) CountComments(postId string) (int64, error) {
	args := m.Called(postId)
	return args.Get(0).(int64), args.Error(1)
}
