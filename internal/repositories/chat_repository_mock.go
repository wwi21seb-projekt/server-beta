package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) GetChatMessages(chatId string, offset int, limit int) ([]models.Message, error) {
	args := m.Called(chatId, offset, limit)
	return args.Get(0).([]models.Message), args.Error(1)
}

func (m *MockChatRepository) GetAllChats(username string) ([]models.Chat, error) {
	args := m.Called(username)
	return args.Get(0).([]models.Chat), args.Error(1)
}
