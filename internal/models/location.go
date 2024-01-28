package models

import (
	"github.com/google/uuid"
)

type Location struct {
	Id        uuid.UUID `gorm:"column:id,primary_key"`
	Longitude string    `gorm:"type:varchar(25)"`
	Latitude  string    `gorm:"type:varchar(25)"`
	Accuracy  int       `gorm:"type:integer"`
}

type LocationDTO struct {
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
	Accuracy  int    `json:"accuracy"`
}
