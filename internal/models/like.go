package models

import (
	"github.com/google/uuid"
)

type Like struct {
	Id           uuid.UUID `gorm:"column:id;primary_key"`
	LikedPostId  string    `gorm:"foreignKey:id;references:LikedPostId"`
	LikeUsername string    `gorm:"foreignKey:username;references:LikeUserId"`
}

type LikePostRequestDTO struct {
	LikedPostId string `json:"postId" binding:"required"`
}

type LikePostResponseDTO struct {
}

type LikeDeleteRequestDTO struct {
	LikedPostId string `json:"postId" binding:"required"`
}

type LikeDeleteResponseDTO struct {
}
