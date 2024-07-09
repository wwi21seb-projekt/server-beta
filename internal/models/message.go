package models

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	Id        uuid.UUID `gorm:"column:id;primary_key"`
	ChatId    uuid.UUID `gorm:"column:chat_id"`
	Chat      Chat      `gorm:"foreignKey:chat_id;references:id"`
	Username  string    `gorm:"column:username_fk;type:varchar(20)"`
	User      User      `gorm:"foreignKey:username_fk;references:username"`
	Content   string    `gorm:"column:content;type:varchar(256);null"`
	CreatedAt time.Time `gorm:"column:created_at;not_null"`
}

type MessageRecordDTO struct {
	Content      string    `json:"content"`
	Username     string    `json:"username"`
	CreationDate time.Time `json:"creationDate"`
}

type MessagesResponseDTO struct {
	Records    []MessageRecordDTO   `json:"records"`
	Pagination *OffsetPaginationDTO `json:"pagination"`
}

type MessageCreateRequestDTO struct {
	Content string `json:"content" binding:"required"`
}
