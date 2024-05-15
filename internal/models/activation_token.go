package models

import (
	"github.com/google/uuid"
	"time"
)

type ActivationToken struct {
	Id             uuid.UUID `gorm:"column:id;primary_key"`
	Username       string    `gorm:"column:username_fk;type:varchar(20)"`
	User           User      `gorm:"foreignKey:username_fk;references:username"`
	Token          string    `gorm:"column:token;not_null;varchar(6)"`
	ExpirationTime time.Time `gorm:"column:expiration_time;not_null"`
}

type ActivationTokenRequestDTO struct {
	Token string `json:"token" binding:"required"`
}
