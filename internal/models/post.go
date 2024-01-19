package models

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	Id        uuid.UUID `gorm:"column:id;primary_key"`
	Username  string    `gorm:"column:username"`
	User      User      `gorm:"foreignKey:username;references:username"`
	Content   string    `gorm:"column:content;type:varchar(256);null"`
	ImageUrl  string    `gorm:"column:image_url;type:varchar(256);null"`
	Hashtags  []Hashtag `gorm:"many2many:post_hashtags;"` // gorm handles the join table
	CreatedAt time.Time `gorm:"column:created_at;not_null"`
}

type PostCreateRequestDTO struct {
	Content string `json:"content" binding:"required"`
}

type PostCreateResponseDTO struct {
	PostId       uuid.UUID  `json:"postId"`
	Author       *AuthorDTO `json:"author"`
	CreationDate time.Time  `json:"creationDate"`
	Content      string     `json:"content"`
}

type AuthorDTO struct { // to be used in post response dto
	Username          string `json:"username"`
	Nickname          string `json:"nickname"`
	ProfilePictureUrl string `json:"profilePictureUrl"`
}

type UserFeedDTO struct {
	Records    []UserFeedRecordDTO    `json:"records"`
	Pagination *UserFeedPaginationDTO `json:"pagination"`
}

type UserFeedRecordDTO struct {
	PostId       string    `json:"postId"`
	CreationDate time.Time `json:"creationDate"`
	Content      string    `json:"content"`
}

type UserFeedPaginationDTO struct {
	Offset  int   `json:"offset"`
	Limit   int   `json:"limit"`
	Records int64 `json:"records"`
}

type GeneralFeedDTO struct { // to be used for response to feed request
	Records    []PostCreateResponseDTO   `json:"records"`
	Pagination *GeneralFeedPaginationDTO `json:"pagination"`
}

type GeneralFeedPaginationDTO struct {
	LastPostId string `json:"lastPostId"`
	Limit      int    `json:"limit"`
	Records    int64  `json:"records"`
}
