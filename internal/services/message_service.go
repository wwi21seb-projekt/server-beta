package services

type MessageServiceInterface interface {
	GetMessagesByChatId(chatId, currentUsername string)
}
