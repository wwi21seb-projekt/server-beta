package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type MessageRepositoryInterface interface {
	GetMessagesByChatId(chatId string, offset int, limit int) ([]models.Message, int64, error)
}

type MessageRepository struct {
	DB *gorm.DB
}

// NewMessageRepository can be used as a constructor to create a MessageRepository "object"
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{DB: db}
}

func (repo *MessageRepository) GetMessagesByChatId(chatId string, offset int, limit int) ([]models.Message, int64, error) {
	var messages []models.Message
	var count int64

	baseQuery := repo.DB.
		Model(&models.Message{}).
		Where("chat_id = ?", chatId).
		Order("created_at desc")

	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	err = baseQuery.Offset(offset).Limit(limit).Preload("User").Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	return messages, count, err
}
