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
	Content string `json:"content" binding:"required"` // MARC: das attribut heißt hier nicht content
}

type SubscriptionDeleteRequestDTO struct {
	Content string `json:"content" binding:"required"` // MARC: es gibt kein subscription delete request body
}
type SubscriptionPostResponseDTO struct { // MARC: hier fehlt der Zeitstempel
	SubscriptionId uuid.UUID `json:"postId"`
	Follower       string    `json:"username"`
	Following      string    `json:"following"`
}
type SubscriptionDeleteResponseDTO struct { // MARC: Auch den gibt und braucht es nicht
	SubscriptionId uuid.UUID `json:"postId"`
	Follower       string    `json:"username"`
	Following      string    `json:"following"`
}
