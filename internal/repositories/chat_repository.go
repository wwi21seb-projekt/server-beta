package repositories

import (
	"errors"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type ChatRepositoryInterface interface {
	GetChatMessages(chatId string, offset int, limit int) ([]models.Message, error)
	GetAllChats(username string) ([]models.Chat, error)
}

type ChatRepository struct {
	DB *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{DB: db}
}

func (repo *ChatRepository) GetChatMessages(chatId string, offset int, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := repo.DB.Where("chat_id = ?", chatId).Order("created_at desc").Offset(offset).Limit(limit).Find(&messages).Error
	if err != nil {
		if errors.Is(gorm.ErrRecordNotFound, err) {
			return nil, customerrors.ChatNotFound
		}
		return nil, err
	}
	return messages, nil
}

func (repo *ChatRepository) GetAllChats(username string) ([]models.ChatDTO, error) {
	var chats []models.ChatDTO
	err := repo.DB.
		Joins("JOIN chat_users ON chats.id = chat_users.chat_id").
		Where("chat_users.username = ?", username).
		Find(&chats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ChatNotFound
		}
		return nil, err
	}
	return chats, nil
}
