package models

import (
	"github.com/google/uuid"
	"time"
)

type Chat struct {
	Id        uuid.UUID `gorm:"column:id;primary_key"`
	Users     []User    `gorm:"many2many:chat_users;"` // gorm handles the join table
	CreatedAt time.Time `gorm:"column:created_at;not_null"`
}

type ChatUserDTO struct {
	Username          string `json:"username"`
	Nickname          string `json:"nickname"`
	ProfilePictureUrl string `json:"profilePictureUrl"`
}

type ChatCreateRequestDTO struct {
	Content  string `json:"content" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type FirstMessageResponseDTO struct {
	Content      string    `json:"content"`
	Username     string    `json:"username"`
	CreationDate time.Time `json:"creationDate"`
}

type ChatCreateResponseDTO struct {
	ChatId  string                   `json:"chatId"`
	Message *FirstMessageResponseDTO `json:"message"`
}

type ChatRecordDTO struct {
	ChatId string       `json:"id"`
	User   *ChatUserDTO `json:"user"`
}

type ChatsResponseDTO struct {
	Records []ChatRecordDTO `json:"records"`
}