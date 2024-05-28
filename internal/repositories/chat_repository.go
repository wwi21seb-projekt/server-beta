package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type ChatRepositoryInterface interface {
	GetChatsByUsername(username string) ([]models.Chat, error)
	GetChatById(chatId string) (models.Chat, error)
}

type ChatRepository struct {
	DB *gorm.DB
}

// NewChatRepository can be used as a constructor to create a ChatRepository "object"
func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{DB: db}
}

func (repo *ChatRepository) GetChatsByUsername(username string) ([]models.Chat, error) {
	var chats []models.Chat
	err := repo.DB.
		Joins("JOIN chat_users ON chats.id = chat_users.chat_id").
		Where("chat_users.user_username = ?", username).
		Preload("Users").
		Find(&chats).Error
	return chats, err
}

func (repo *ChatRepository) GetChatById(chatId string) (models.Chat, error) {
	var chat models.Chat
	err := repo.DB.Where("id = ?", chatId).Preload("Users").First(&chat).Error
	return chat, err
}
