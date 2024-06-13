package models

import (
	"github.com/google/uuid"
	"time"
)

type Notification struct {
	Id               uuid.UUID `gorm:"column:id;primaryKey"`
	Timestamp        time.Time `gorm:"column:timestamp"`
	NotificationType string    `gorm:"column:notification_type"`
	ForUsername      string    `gorm:"column:for_username"` // the user that the notification is for
	ForUser          User      `gorm:"foreignKey:for_username;references:username"`
	FromUsername     string    `gorm:"column:from_username"` // the user that created the notification by following or reposting
	FromUser         User      `gorm:"foreignKey:from_username;references:username"`
}

type NotificationUserDTO struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Picture  *Image `json:"picture"`
}

type NotificationRecordDTO struct {
	NotificationId   string               `json:"notificationId"`
	Timestamp        time.Time            `json:"timestamp"`
	NotificationType string               `json:"notificationType"`
	User             *NotificationUserDTO `json:"user"`
}

type NotificationsResponseDTO struct {
	Records []NotificationRecordDTO `json:"records"`
}
