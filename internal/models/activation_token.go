package models

import (
	"github.com/google/uuid"
	"time"
)

type ActivationToken struct {
	Id             uuid.UUID `gorm:"column:id;primary_key"`
	Username       string    `gorm:"column:username"`
	User           User      `gorm:"foreignKey:username"`
	Token          string    `gorm:"column:token;not_null"`
	ExpirationTime time.Time `gorm:"column:expiration_time;not_null"`
}

type ActivationTokenRequestDTO struct {
	Token string `json:"token" binding:"required"`
}
