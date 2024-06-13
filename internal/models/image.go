package models

type Image struct {
	ImageUrl string `gorm:"column:imageUrl;primary_key;type:varchar(255)"`
	Width    int    `gorm:"column:width;type:int"`
	Height   int    `gorm:"column: height;type:int"`
}
