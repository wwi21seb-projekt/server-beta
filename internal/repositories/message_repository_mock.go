package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) GetMessagesByChatId(chatId string, offset int, limit int) ([]models.Message, int64, error) {
	args := m.Called(chatId, offset, limit)
	return args.Get(0).([]models.Message), args.Get(1).(int64), args.Error(2)
}
