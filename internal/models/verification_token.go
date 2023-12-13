package models

import (
	"github.com/google/uuid"
	"time"
)

type VerificationToken struct {
	Id             uuid.UUID `gorm:"column:id,primary_key"`
	Username       string    `gorm:"column:username"`
	Token          string    `gorm:"column:token;not_null"`
	ExpirationTime time.Time `gorm:"column:expiration_time;not_null"`
}

type VerificationTokenRequestDTO struct {
	Token string `json:"token" binding:"required"`
}
