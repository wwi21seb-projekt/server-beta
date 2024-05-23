package models

import "time"

type Chat struct {
	Id        string    `gorm:"column:id;primary_key"`
	Users     []User    `gorm:"many2many:chat_users;"` // gorm handles the join table
	CreatedAt time.Time `gorm:"column:created_at;not_null"`
}
