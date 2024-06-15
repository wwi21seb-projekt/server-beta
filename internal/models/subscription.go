package models

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	Id                uuid.UUID `gorm:"column:id;primary_key"`
	SubscriptionDate  time.Time `gorm:"column:subscription_date;not null"`
	FollowerUsername  string    `gorm:"column:follower;type:varchar(20)"`
	Follower          User      `gorm:"foreignKey:follower;references:username"` // Person who follows
	FollowingUsername string    `gorm:"column:following;type:varchar(20)"`       // Person who is being followed
	Following         User      `gorm:"foreignKey:following;references:username"`
}

type SubscriptionPostRequestDTO struct {
	Following string `json:"following" binding:"required"`
}

type SubscriptionPostResponseDTO struct {
	SubscriptionId   uuid.UUID `json:"subscriptionId"`
	SubscriptionDate time.Time `json:"subscriptionDate"`
	Follower         string    `json:"follower"`
	Following        string    `json:"following"`
}

type SubscriptionResponseDTO struct {
	Records    []UserSubscriptionRecordDTO `json:"records"`
	Pagination *OffsetPaginationDTO        `json:"pagination"`
}

type UserSubscriptionRecordDTO struct {
	FollowerId  *uuid.UUID        `json:"followerId"`  // SubscriptionId, if user follows me - may be null
	FollowingId *uuid.UUID        `json:"followingId"` // SubscriptionId, if I follow user - may be null
	Username    string            `json:"username"`    // Username of follower/following
	Nickname    string            `json:"nickname"`
	Picture     *ImageMetadataDTO `json:"picture"`
}

type UserSubscriptionSQLRecordDTO struct { // to be used for sql query results
	FollowerId  *uuid.UUID
	FollowingId *uuid.UUID
	Username    string
	Nickname    string
	ImageId     string
	Format      string
	Width       int
	Height      int
	Tag         time.Time
}
