package models

import (
	"github.com/google/uuid"
)

type Location struct {
	Id        uuid.UUID `gorm:"column:id;primary_key"`
	Longitude string    `gorm:"type:varchar(25)"`
	Latitude  string    `gorm:"type:varchar(25)"`
	Accuracy  uint      `gorm:"type:integer"`
}

type LocationDTO struct {
	Longitude string `json:"longitude" binding:"required"`
	Latitude  string `json:"latitude" binding:"required"`
	Accuracy  uint   `json:"accuracy" binding:"required"`
}
