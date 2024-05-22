package models

import "github.com/google/uuid"

type PushSubscription struct {
	Id       uuid.UUID `gorm:"column:id;primary_key"`
	Username string    `gorm:"column:username_fk;type:varchar(20)"`
	User     User      `gorm:"foreignKey:username_fk;references:username"`
	Type     string    `gorm:"column:type;type:varchar(4)"` // either "web" or "expo"
	Endpoint string    `gorm:"column:endpoint;type:text"`   // for web only
	P256dh   string    `gorm:"column:p256dh;type:text"`     // for web only
	Auth     string    `gorm:"column:auth;type:text"`       // for web only
	Token    string    `gorm:"column:token;type:text"`      // for expo only
}

type VapidKeyResponseDTO struct {
	Key string `json:"key"`
}

type SubscriptionKeys struct {
	P256dh string `json:"p256dh" binding:"required"`
	Auth   string `json:"auth" binding:"required"`
}

type SubscriptionInfo struct {
	Endpoint         string           `json:"endpoint" binding:"required"`
	SubscriptionKeys SubscriptionKeys `json:"keys" binding:"required"`
}

type PushSubscriptionRequestDTO struct {
	Type             string            `json:"type" binding:"required"`
	SubscriptionInfo *SubscriptionInfo `json:"subscription"` // subscription info for web push notifications
	Token            string            `json:"token"`        // token for expo push notifications
}

type PushSubscriptionResponseDTO struct {
	SubscriptionId string `json:"subscriptionId"`
}
