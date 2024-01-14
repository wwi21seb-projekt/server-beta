package models

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	Id                uuid.UUID `gorm:"column:id;primary_key"`
	SubscriptionDate  time.Time `gorm:"column:subscription_date;not null"`
	FollowerUsername  string    `gorm:"column:follower;type:varchar(20)"`
	Follower          User      `gorm:"foreignKey:username"`               // Person who follows
	FollowingUsername string    `gorm:"column:following;type:varchar(20)"` // Person who is being followed
	Following         User      `gorm:"foreignKey:username"`
}

type SubscriptionPostRequestDTO struct {
	Following string `json:"following" binding:"required"`
}

type SubscriptionPostResponseDTO struct {
	SubscriptionId   uuid.UUID `json:"subscriptionId"`
	SubscriptionDate time.Time `json:"subscriptionDate"`
	Follower         string    `json:"username"`
	Following        string    `json:"following"`
}
