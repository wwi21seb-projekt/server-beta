package models

import (
	"github.com/google/uuid"
)

type Location struct {
	Id        uuid.UUID `gorm:"column:id;primary_key"`
	Longitude float64   `gorm:"type:float"`
	Latitude  float64   `gorm:"type:float"`
	Accuracy  uint      `gorm:"type:integer"`
}

type LocationDTO struct {
	Longitude *float64 `json:"longitude" binding:"required"` // using pointer to allow values to be zero
	Latitude  *float64 `json:"latitude" binding:"required"`
	Accuracy  *uint    `json:"accuracy" binding:"required"`
}
