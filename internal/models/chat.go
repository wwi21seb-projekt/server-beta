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

type ChatPostRequestDTO struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}
