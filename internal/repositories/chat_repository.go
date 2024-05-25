package repositories

import (
	"errors"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type ChatRepositoryInterface interface {
	CreateChat(chat *models.Chat) error
	GetChatMessages(username string, offset int, limit int) ([]models.Chat, error)
	GetAllChats(username string) ([]models.Chat, error)
}

type ChatRepository struct {
	DB *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{DB: db}
}

func (repo *ChatRepository) CreateChat(chat *models.Chat) error {
	err := repo.DB.Create(chat).Error
	return err
}

func (repo *ChatRepository) GetChatMessages(username string, offset int, limit int) ([]models.Chat, error) {
	var chats []models.Chat
	err := repo.DB.Where("sender = ? OR receiver = ?", username, username).Order("created_at desc").Offset(offset).Limit(limit).Find(&chats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.BadRequest
		}
		return nil, err
	}
	return chats, nil
}

func (repo *ChatRepository) GetAllChats(username string) ([]models.Chat, error) {
	var chats []models.Chat
	err := repo.DB.Where("sender = ? OR receiver = ?", username, username).Find(&chats).Error
	return chats, err
}
