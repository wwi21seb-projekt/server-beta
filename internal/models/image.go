package models

import (
	"github.com/google/uuid"
	"time"
)

type Image struct {
	Id        uuid.UUID `gorm:"column:id;primary_key"`
	Format    string    `gorm:"column:format;type:varchar(10)"` // e.g. ".jpeg", ".png",...
	ImageData []byte    `gorm:"column:image_data;type:bytea"`
	Width     int       `gorm:"column:width;type:int"`
	Height    int       `gorm:"column:height;type:int"`
	Tag       time.Time `gorm:"column:tag;not_null"` // timestamp of creation/last edit to enable cache invalidation for clients
}

type ImageMetadataDTO struct {
	Url    string    `json:"url"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
	Tag    time.Time `json:"tag"`
}

type ImageDTO struct {
	Format string `json:"format"`
	Data   []byte `json:"data"`
}
