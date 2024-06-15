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

type ChatCreateRequestDTO struct {
	Content  string `json:"content" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type ChatCreateResponseDTO struct {
	ChatId  string            `json:"chatId"`
	Message *MessageRecordDTO `json:"message"`
}

type ChatRecordDTO struct {
	ChatId string   `json:"chatId"`
	User   *UserDTO `json:"user"`
}

type ChatsResponseDTO struct {
	Records []ChatRecordDTO `json:"records"`
}
