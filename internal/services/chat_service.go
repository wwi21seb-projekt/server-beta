package services

import (
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"net/http"
)

type ChatServiceInterface interface {
	GetChatsByUsername(username string) (*models.ChatsResponseDTO, *customerrors.CustomError, int)
}

type ChatService struct {
	chatRepository repositories.ChatRepositoryInterface
}

// NewChatService creates a new instance of the ChatService
func NewChatService(chatRepository repositories.ChatRepositoryInterface) *ChatService {
	return &ChatService{chatRepository: chatRepository}
}

// GetChatsByUsername retrieves all chats of a user by its username
func (service *ChatService) GetChatsByUsername(username string) (*models.ChatsResponseDTO, *customerrors.CustomError, int) {
	// Get Chats by username
	chats, err := service.chatRepository.GetChatsByUsername(username)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response
	chatDTOs := make([]models.ChatRecordDTO, 0)
	for _, chat := range chats {

		// Currently, we only have two users in a chat --> find the other user
		var chatUserDto models.ChatUserDTO
		for _, user := range chat.Users {
			if user.Username != username {
				chatUserDto = models.ChatUserDTO{
					Username:          user.Username,
					Nickname:          user.Nickname,
					ProfilePictureUrl: user.ProfilePictureUrl,
				}
				break
			}
		}

		chatDTO := models.ChatRecordDTO{
			ChatId: chat.Id.String(),
			User:   &chatUserDto,
		}

		chatDTOs = append(chatDTOs, chatDTO)
	}

	response := models.ChatsResponseDTO{
		Records: chatDTOs,
	}

	return &response, nil, http.StatusOK
}
