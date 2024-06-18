package models

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	Id         uuid.UUID  `gorm:"column:id;primary_key"`
	Username   string     `gorm:"column:username_fk;type:varchar(20)"`
	User       User       `gorm:"foreignKey:username_fk;references:username"`
	Content    string     `gorm:"column:content;type:varchar(256);null"`
	ImageId    *uuid.UUID `gorm:"column:image_id;null"`
	Image      Image      `gorm:"foreignKey:image_id;references:id"`
	Hashtags   []Hashtag  `gorm:"many2many:post_hashtags;onDelete:CASCADE"` // gorm handles the join table, onDelete:CASCADE deletes the hashtags if the post is deleted
	CreatedAt  time.Time  `gorm:"column:created_at;not_null"`
	LocationId *uuid.UUID `gorm:"column:location_id;null"`
	Location   Location   `gorm:"foreignKey:location_id;references:id"`
	RepostId   *uuid.UUID `gorm:"column:repost_id;null"` // no foreign key constraint, original post may be deleted without affecting repost
}

type PostCreateRequestDTO struct {
	Content        string       `json:"content"`
	Picture        string       `json:"picture"`
	Location       *LocationDTO `json:"location"`
	RepostedPostId string       `json:"repostedPostId"`
}

type PostResponseDTO struct {
	PostId       uuid.UUID         `json:"postId"`
	Author       *UserDTO          `json:"author"`
	CreationDate time.Time         `json:"creationDate"`
	Content      string            `json:"content"`
	Picture      *ImageMetadataDTO `json:"picture"`
	Comments     int64             `json:"comments"`
	Likes        int64             `json:"likes"`
	Liked        bool              `json:"liked"`
	Location     *LocationDTO      `json:"location"`
	Repost       *PostResponseDTO  `json:"repost"`
}

type GeneralFeedDTO struct { // to be used for response to general feed request
	Records    []PostResponseDTO        `json:"records"`
	Pagination *PostCursorPaginationDTO `json:"pagination"`
}

type UserFeedDTO struct { // to be used for response to user feed request
	Records    []UserFeedRecordDTO  `json:"records"`
	Pagination *OffsetPaginationDTO `json:"pagination"`
}

type UserFeedRecordDTO struct { // Post response dto without author for user feed
	PostId       string            `json:"postId"`
	CreationDate time.Time         `json:"creationDate"`
	Content      string            `json:"content"`
	Picture      *ImageMetadataDTO `json:"picture"`
	Comments     int64             `json:"comments"`
	Likes        int64             `json:"likes"`
	Liked        bool              `json:"liked"`
	Location     *LocationDTO      `json:"location"`
	Repost       *PostResponseDTO  `json:"repost"`
}
