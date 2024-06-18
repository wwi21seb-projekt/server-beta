package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Username     string     `gorm:"column:username;primary_key;type:varchar(20)"`
	Nickname     string     `gorm:"column:nickname;type:varchar(25)"`
	Email        string     `gorm:"column:email;type:varchar(128);not_null;unique"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(80);not_null"`
	CreatedAt    time.Time  `gorm:"column:created_at;not_null"`
	Activated    bool       `gorm:"column:activated;not_null"`
	ImageId      *uuid.UUID `gorm:"column:image_id;null"`
	Image        Image      `gorm:"foreignKey:image_id;references:id"`
	Status       string     `gorm:"column:status;type:varchar(128)"`
	Chats        []Chat     `gorm:"many2many:chat_users;"` // gorm handles the join table
}

type UserDTO struct { // General dto for user, also used as author dto
	Username string            `json:"username"`
	Nickname string            `json:"nickname"`
	Picture  *ImageMetadataDTO `json:"picture"`
}

type UserCreateRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Picture  string `json:"picture"`
	Email    string `json:"email" binding:"required"`
}

type UserCreateResponseDTO struct {
	Username string            `json:"username"`
	Nickname string            `json:"nickname"`
	Picture  *ImageMetadataDTO `json:"profilePicture"`
	Email    string            `json:"email"`
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

type UserInformationUpdateRequestDTO struct {
	Nickname string  `json:"nickname" binding:"required"`
	Status   string  `json:"status" binding:"required"`
	Picture  *string `json:"picture"`
}

type UserInformationUpdateResponseDTO struct {
	Nickname string            `json:"nickname"`
	Status   string            `json:"status"`
	Picture  *ImageMetadataDTO `json:"picture"`
}

type ChangePasswordDTO struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type UserSearchResponseDTO struct {
	Records    []UserDTO            `json:"records"`
	Pagination *OffsetPaginationDTO `json:"pagination"`
}

type UserProfileResponseDTO struct {
	Username       string            `json:"username"`
	Nickname       string            `json:"nickname"`
	Status         string            `json:"status"`
	Picture        *ImageMetadataDTO `json:"picture"`
	Follower       int64             `json:"follower"`
	Following      int64             `json:"following"`
	Posts          int64             `json:"posts"`
	SubscriptionId *string           `json:"subscriptionId"`
}
