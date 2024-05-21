package models

import (
	"github.com/google/uuid"
)

type Like struct {
	Id       uuid.UUID `gorm:"column:id;primary_key"`
	PostId   uuid.UUID `gorm:"column:post_id"`
	Post     Post      `gorm:"foreignKey:post_id;references:id"`
	Username string    `gorm:"column:username_fk"`
	User     User      `gorm:"foreignKey:username_fk;references:username"`
}
