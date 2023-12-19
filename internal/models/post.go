package models

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	Id        uuid.UUID `gorm:"column:id;primary_key"`
	Username  string    `gorm:"column:username;type:varchar(20)"`
	User      User      `gorm:"foreignKey:username"`
	Content   string    `gorm:"column:content;type:varchar(256);null"`
	ImageUrl  string    `gorm:"column:image_url;type:varchar(256);null"`
	Hashtags  []string  `gorm:"column:hashtags;null"`
	CreatedAt time.Time `gorm:"column:created_at;not_null"`
}

type PostCreateRequestDTO struct {
	Content string `json:"content" binding:"required"`
}

type PostCreateResponseDTO struct {
	PostId       uuid.UUID  `json:"postId"`
	Author       *AuthorDTO `json:"author"`
	CreationDate time.Time  `json:"creationDate"`
	Content      string     `json:"content"`
}

type AuthorDTO struct { // to be used in post response dto
	Username          string `json:"username"`
	Nickname          string `json:"nickname"`
	ProfilePictureUrl string `json:"profilePictureUrl"`
}
