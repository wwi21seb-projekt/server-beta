package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type ChatRepositoryInterface interface {
	CreateChatWithFirstMessage(chat models.Chat, message models.Message) error
	GetChatByUsernames(currentUsername, otherUsername string) (models.Chat, error)
	GetChatsByUsername(username string) ([]models.Chat, error)
	GetChatById(chatId string) (models.Chat, error)
}

type ChatRepository struct {
	DB *gorm.DB
}

// NewChatRepository can be used as a constructor to create a ChatRepository "object"
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
		Joins("JOIN users u1 ON cu1.user_username = u1.username").
		Joins("JOIN users u2 ON cu2.user_username = u2.username").
		Where("u1.username = ? AND u2.username = ?", currentUsername, otherUsername).
		First(&chat)
	return chat, query.Error
}

func (repo *ChatRepository) GetChatsByUsername(username string) ([]models.Chat, error) {
	var chats []models.Chat

	subQuery := repo.DB.Table("messages").
		Select("chat_id", "MAX(created_at) as last_message_date").
		Group("chat_id") // Sub query to get the latest message date for each chat

	err := repo.DB.
		Joins("JOIN chat_users ON chats.id = chat_users.chat_id").
		Joins("LEFT JOIN (?) as latest_messages ON chats.id = latest_messages.chat_id", subQuery).
		Where("chat_users.user_username = ?", username).
		Preload("Users").
		Preload("Users.Image").
		Order("latest_messages.last_message_date DESC"). // Order chats by latest message date
		Find(&chats).Error
	return chats, err
}

func (repo *ChatRepository) GetChatById(chatId string) (models.Chat, error) {
	var chat models.Chat
	err := repo.DB.Where("id = ?", chatId).Preload("Users").Preload("Users.Image").First(&chat).Error
	return chat, err
}
