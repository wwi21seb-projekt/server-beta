package models

import (
	"github.com/google/uuid"
)

type Subscription struct {
	SubscriptionId uuid.UUID `gorm:"column:id;primary_key"`
	Follower       User      `gorm:"foreignKey:username"`
	Following      User      `gorm:"foreignKey:username"`
}

type SubscriptionPostRequestDTO struct {
	Content string `json:"content" binding:"required"`
}

type SubscriptionDeleteRequestDTO struct {
	Content string `json:"content" binding:"required"`
}
type SubscriptionPostResponseDTO struct {
	SubscriptionId uuid.UUID `json:"postId"`
	Follower       string    `json:"username"`
	Following      string    `json:"following"`
}
type SubscriptionDeleteResponseDTO struct {
	SubscriptionId uuid.UUID `json:"postId"`
	Follower       string    `json:"username"`
	Following      string    `json:"following"`
}
