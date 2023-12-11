package models

type User struct {
	Username     string `json:"username" gorm:"primary_key;type:varchar(20);not null;unique"`
	Nickname     string `json:"nickname" gorm:"type:varchar(25)"`
	Email        string `json:"email" gorm:"type:varchar(128);not null;unique"`
	PasswordHash string `gorm:"type:varchar(80);not null"`
	Activated    bool   `gorm:"not null"`
}

type UserCreateRequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

type UserLoginRequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponseDTO struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}
