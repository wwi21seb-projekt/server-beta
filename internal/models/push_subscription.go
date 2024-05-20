package models

import "github.com/google/uuid"

type PushSubscription struct {
	Id       uuid.UUID `gorm:"column:id;primary_key"`
	Username string    `gorm:"column:username_fk;type:varchar(20)"`
	User     User      `gorm:"foreignKey:username_fk;references:username"`
	Type     string    `gorm:"column:type;type:varchar(4)"` // either "web" or "expo"
	Endpoint string    `gorm:"column:endpoint;type:text"`
	P256dh   string    `gorm:"column:p256dh;type:text"`
	Auth     string    `gorm:"column:auth;type:text"`
}

type VapidKeyResponseDTO struct {
	Key string `json:"key"`
}

type SubscriptionKeys struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}

type SubscriptionInfo struct {
	Endpoint         string           `json:"endpoint"`
	SubscriptionKeys SubscriptionKeys `json:"keys"`
}

type PushSubscriptionRequestDTO struct {
	Type             string           `json:"type" binding:"required"`
	SubscriptionInfo SubscriptionInfo `json:"subscription" binding:"required"`
}

type PushSubscriptionResponseDTO struct {
	SubscriptionId string `json:"subscriptionId"`
}
