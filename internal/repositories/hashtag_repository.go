package repositories

import (
	"errors"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type HashtagRepositoryInterface interface {
	FindOrCreateHashtag(name string) (models.Hashtag, error)
}

type HashtagRepository struct {
	db *gorm.DB
}

// NewHashtagRepository can be used as a constructor to return a new HashtagRepository "object"
func NewHashtagRepository(db *gorm.DB) *HashtagRepository {
	return &HashtagRepository{db: db}
}

// FindOrCreateHashtag finds a hashtag by name or creates it if it doesn't exist
func (repo *HashtagRepository) FindOrCreateHashtag(name string) (models.Hashtag, error) {
	var hashtag models.Hashtag
	err := repo.db.Where("name = ?", name).First(&hashtag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// hashtag doesn't exist yet, create it
			hashtag = models.Hashtag{
				Id:   uuid.New(),
				Name: name,
			}
			err = repo.db.Create(&hashtag).Error
		}
		if err != nil {
			return models.Hashtag{}, err
		}
	}
	return hashtag, nil
}
