package models

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	Id                uuid.UUID `gorm:"column:id;primary_key"`
	SubscriptionDate  time.Time `gorm:"column:subscription_date;not null"`
	FollowerUsername  string    `gorm:"column:follower;type:varchar(20)"`
	Follower          User      `gorm:"foreignKey:username;references:follower"` // Person who follows
	FollowingUsername string    `gorm:"column:following;type:varchar(20)"`       // Person who is being followed
	Following         User      `gorm:"foreignKey:username;references:following"`
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

type SubscriptionResponseDTO struct {
	Records    []UserSubscriptionRecordDTO `json:"records"`
	Pagination *SubscriptionPaginationDTO  `json:"pagination"`
}

type UserSubscriptionRecordDTO struct {
	FollowerId        uuid.UUID `json:"followerId"`        // SubscriptionID, wenn Nutzer mir folgt - ggf. null
	FollowingId       uuid.UUID `json:"followingId"`       // SubscriptionID, wenn ich Nutzer folge - ggf. null
	Username          string    `json:"username"`          // Der Benutzername des Followers/Following
	Nickname          string    `json:"nickname"`          // Der Spitzname des Followers/Following
	ProfilePictureUrl string    `json:"profilePictureUrl"` // Die URL des Profilbildes des Followers/Following
}

type SubscriptionPaginationDTO struct {
	Offset  int   `json:"offset"`
	Limit   int   `json:"limit"`
	Records int64 `json:"records"`
}
