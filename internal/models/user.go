package models

import "time"

type User struct {
	Username     string    `gorm:"primary_key;type:varchar(20);not_null;unique"`
	Nickname     string    `orm:"type:varchar(25)"`
	Email        string    `gorm:"type:varchar(128);not_null;unique"`
	PasswordHash string    `gorm:"type:varchar(80);not_null"`
	CreatedAt    time.Time `gorm:"column:created_at;not_null"`
	Activated    bool      `gorm:"not_null"`
}

type UserCreateRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email" binding:"required"`
}

type UserLoginRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResponseDTO struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type UserActivationRequestDTO struct {
	Token string `json:"token" binding:"required"`
}

type UserResponseDTO struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}
