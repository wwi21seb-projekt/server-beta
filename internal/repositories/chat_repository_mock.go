package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) GetChatsByUsername(username string) ([]models.Chat, error) {
	args := m.Called(username)
	return args.Get(0).([]models.Chat), args.Error(1)
}

func (m *MockChatRepository) GetChatById(chatId string) (models.Chat, error) {
	args := m.Called(chatId)
	return args.Get(0).(models.Chat), args.Error(1)
}
