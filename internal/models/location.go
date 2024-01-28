package models

import (
	"github.com/google/uuid"
)

type Location struct {
	LocationId uuid.UUID `gorm:"column:id,primary_key"`
	Username   string    `gorm:"column:username"`
	User       User      `gorm:"foreignKey:username;references:username"`
	Longitude  string    `gorm:"type:varchar(25)"`
	Latitude   string    `gorm:"type:varchar(25)"`
	Accuracy   int       `gorm:"type:integer"`
}

type LocationResponseDTO struct {
	Longitude string `json:"author"`
	Latitude  string `json:"creationDate"`
	Accuracy  int    `json:"content"`
}
