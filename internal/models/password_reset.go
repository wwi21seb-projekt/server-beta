package models

import (
	"github.com/google/uuid"
	"time"
)

type PasswordResetResponseDTO struct {
	CensoredEmail string `json:"censoredEmail"`
}

type SetNewPasswordDTO struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type PasswordResetRequestDTO struct {
	Username string `json:"username" binding:"required"`
}

type PasswordResetToken struct {
	Id             uuid.UUID `gorm:"type:uuid;primary_key;"`
	Username       string    `gorm:"column:username;type:varchar(20);not_null"`
	User           User      `gorm:"foreignKey:username_fk;references:username"`
	Token          string    `gorm:"column:token;type:varchar(6);not_null"`
	ExpirationTime time.Time `gorm:"column:expiration_time;not_null"`
}
