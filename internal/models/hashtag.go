package models

import "github.com/google/uuid"

type Hashtag struct {
	Id    uuid.UUID `gorm:"primary_key"`
	Name  string    `gorm:"uniqueIndex"`
	Posts []Post    `gorm:"many2many:post_hashtags;"` // gorm handles the join table
}
