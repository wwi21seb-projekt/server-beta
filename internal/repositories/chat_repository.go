package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type ChatRepositoryInterface interface {
	CreateChatWithFirstMessage(chat models.Chat, message models.Message) error
	GetChatByUsernames(currentUsername, otherUsername string) (models.Chat, error)
}

type ChatRepository struct {
	DB *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository { return &ChatRepository{DB: db} }

func (repo *ChatRepository) CreateChatWithFirstMessage(chat models.Chat, message models.Message) error {
	// Create a transaction to ensure that both the chat and the message are created
	return repo.DB.Transaction(func(tx *gorm.DB) error {
		// Save chat
		if err := tx.Create(chat).Error; err != nil {
			return err
		}
		message.ChatId = chat.Id
		// Save message
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *ChatRepository) GetChatByUsernames(currentUsername, otherUsername string) (models.Chat, error) {
	var chat models.Chat
	query := repo.DB.Joins("JOIN chat_users cu1 ON cu1.chat_id = chats.id").
		Joins("JOIN chat_users cu2 ON cu2.chat_id = chats.id").
		Joins("JOIN users u1 ON cu1.user_id = u1.id").
		Joins("JOIN users u2 ON cu2.user_id = u2.id").
		Where("u1.username = ? AND u2.username = ?", currentUsername, otherUsername).
		First(&chat)
	return chat, query.Error
}
