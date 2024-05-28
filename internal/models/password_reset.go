package models

import (
	"github.com/google/uuid"
	"time"
)

type PasswordResetToken struct {
	Id             uuid.UUID `gorm:"type:uuid;primary_key;"`
	Username       string    `gorm:"column:username_fk;type:varchar(20);not_null"`
	User           User      `gorm:"foreignKey:username_fk;references:username"`
	Token          string    `gorm:"column:token;type:varchar(6);not_null"`
	ExpirationTime time.Time `gorm:"column:expiration_time;not_null"`
}

type InitiatePasswordResetResponseDTO struct {
	Email string `json:"email"` // Response contains censored mail
}

type ResetPasswordRequestDTO struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}
