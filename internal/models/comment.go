package models

import (
	"github.com/google/uuid"
	"time"
)

type Comment struct {
	Id        uuid.UUID `gorm:"column:id;primary_key"`
	PostID    uuid.UUID `gorm:"column:post_id"`
	Post      Post      `gorm:"foreignKey:post_id;references:id"`
	Username  string    `gorm:"column:username_fk;type:varchar(20)"`
	User      User      `gorm:"foreignKey:username_fk;references:username"`
	Content   string    `gorm:"column:content;type:varchar(128);not_null"`
	CreatedAt time.Time `gorm:"column:created_at;not_null"`
}

type CommentCreateRequestDTO struct {
	Content string `json:"content" binding:"required"`
}

type CommentResponseDTO struct {
	CommentId    uuid.UUID `json:"commentId"`
	Content      string    `json:"content"`
	Author       *UserDTO  `json:"author"`
	CreationDate time.Time `json:"creationDate"`
}

type CommentFeedResponseDTO struct {
	Records    []CommentResponseDTO `json:"records"`
	Pagination *OffsetPaginationDTO `json:"pagination"`
}
