package models

import "time"

type Message struct {
	Id        string    `gorm:"column:id;primary_key"`
	ChatId    string    `gorm:"column:chat_id"`
	Chat      Chat      `gorm:"foreignKey:chat_id;references:id"`
	Username  string    `gorm:"column:username_fk;type:varchar(20)"`
	User      User      `gorm:"foreignKey:username_fk;references:username"`
	Content   string    `gorm:"column:content;type:varchar(256);null"`
	CreatedAt time.Time `gorm:"column:created_at;not_null"`
}
