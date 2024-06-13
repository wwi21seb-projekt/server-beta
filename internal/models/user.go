package models

import (
	"time"
)

type User struct {
	Username     string    `gorm:"column:username;primary_key;type:varchar(20)"`
	Nickname     string    `gorm:"column:nickname;type:varchar(25)"`
	Email        string    `gorm:"column:email;type:varchar(128);not_null;unique"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(80);not_null"`
	CreatedAt    time.Time `gorm:"column:created_at;not_null"`
	Activated    bool      `gorm:"column:activated;not_null"`
	ImageURL     string    `gorm:"column:imageUrl;null"`
	Image        Image     `gorm:"foreignKey:image_fk;references:imageUrl"`
	Status       string    `gorm:"column:status;type:varchar(128)"`
	Chats        []Chat    `gorm:"many2many:chat_users;"` // gorm handles the join table
}

type UserCreateRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Image    string `json:"picture"`
	Email    string `json:"email" binding:"required"`
}

type UserLoginRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResponseDTO struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type UserRefreshTokenRequestDTO struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type UserActivationRequestDTO struct {
	Token string `json:"token" binding:"required"`
}

type UserResponseDTO struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Image    *Image `json:"picture"`
	Email    string `json:"email"`
}

type UserSearchResponseDTO struct {
	Records    []UserSearchRecordDTO    `json:"records"`
	Pagination *UserSearchPaginationDTO `json:"pagination"`
}

type UserSearchRecordDTO struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Image    *Image `json:"picture"`
}

type UserSubscriptionSearchRecordDTO struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Image    *Image `json:"picture"`
}

type UserSearchPaginationDTO struct {
	Offset  int   `json:"offset"`
	Limit   int   `json:"limit"`
	Records int64 `json:"records"`
}

type UserInformationUpdateDTO struct {
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
	Image    string `json:"picture"`
}
type UserInformationUpdateResponseDTO struct {
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
	Image    *Image `json:"picture"`
}
type ChangePasswordDTO struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type UserProfileResponseDTO struct {
	Username       string  `json:"username"`
	Nickname       string  `json:"nickname"`
	Status         string  `json:"status"`
	Image          *Image  `json:"picture"`
	Follower       int64   `json:"follower"`
	Following      int64   `json:"following"`
	Posts          int64   `json:"posts"`
	SubscriptionId *string `json:"subscriptionId"`
}
