package models

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	Id         uuid.UUID  `gorm:"column:id;primary_key"`
	Username   string     `gorm:"column:username"`
	User       User       `gorm:"foreignKey:username;references:username"`
	Content    string     `gorm:"column:content;type:varchar(256);null"`
	ImageUrl   string     `gorm:"column:image_url;type:varchar(128);null"`
	Hashtags   []Hashtag  `gorm:"many2many:post_hashtags;onDelete:CASCADE"` // gorm handles the join table, onDelete:CASCADE deletes the hashtags if the post is deleted
	CreatedAt  time.Time  `gorm:"column:created_at;not_null"`
	LocationId *uuid.UUID `gorm:"column:location_id;null"`
	Location   Location   `gorm:"foreignKey:location_id;references:id"`
}

type PostCreateRequestDTO struct {
	Content  string       `json:"content" binding:"required"`
	Location *LocationDTO `json:"location" `
}

type PostResponseDTO struct {
	PostId       uuid.UUID    `json:"postId"`
	Author       *AuthorDTO   `json:"author"`
	CreationDate time.Time    `json:"creationDate"`
	Content      string       `json:"content"`
	Likes        int64        `json:"likes"`
	Liked        bool         `json:"liked"`
	Location     *LocationDTO `json:"location"`
}

type AuthorDTO struct { // to be used in post response dto
	Username          string `json:"username"`
	Nickname          string `json:"nickname"`
	ProfilePictureUrl string `json:"profilePictureUrl"`
}

type UserFeedDTO struct { // to be used for response to user feed request
	Records    []UserFeedRecordDTO    `json:"records"`
	Pagination *UserFeedPaginationDTO `json:"pagination"`
}

type UserFeedRecordDTO struct {
	PostId       string       `json:"postId"`
	CreationDate time.Time    `json:"creationDate"`
	Content      string       `json:"content"`
	Likes        int64        `json:"likes"`
	Liked        bool         `json:"liked"`
	Location     *LocationDTO `json:"location"`
}

type UserFeedPaginationDTO struct {
	Offset  int   `json:"offset"`
	Limit   int   `json:"limit"`
	Records int64 `json:"records"`
}

type GeneralFeedDTO struct { // to be used for response to general feed request
	Records    []PostResponseDTO         `json:"records"`
	Pagination *GeneralFeedPaginationDTO `json:"pagination"`
}

type GeneralFeedPaginationDTO struct {
	LastPostId string `json:"lastPostId"`
	Limit      int    `json:"limit"`
	Records    int64  `json:"records"`
}
