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

type ChatDTO struct {
	Id        uuid.UUID          `json:"id"`
	Users     []string           `json:"users"`
	CreatedAt time.Time          `json:"createdAt"`
	Messages  []MessageRecordDTO `json:"messages"`
}
