package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type MessageRepositoryInterface interface {
	CreateMessage(message *models.Message) error
	DeleteMessage(messageId string) error
	GetChatMessages(chatId string) ([]*models.Message, error)
}

type MessageRepository struct {
	DB *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository { return &MessageRepository{DB: db} }

func (repo *MessageRepository) CreateMessage(message *models.Message) error {
	return repo.DB.Create(message).Error
}

func (repo *MessageRepository) DeleteMessage(messageId string) error {
	return repo.DB.Delete(&models.Message{}, "id = ?", messageId).Error
}

func (repo *MessageRepository) GetChatMessages(chatId string) ([]*models.Message, error) {
	var messages []*models.Message
	query := repo.DB.Model(&models.Message{}).Where("chat_id = ?", chatId).Find(&messages)
	return messages, query.Error
}
