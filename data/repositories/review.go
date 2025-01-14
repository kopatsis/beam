package repositories

import (
	"beam/data/models"

	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(review *models.Review) error
}

type reviewRepo struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepo{db: db}
}

func (r *reviewRepo) Create(review *models.Review) error {
	return r.db.Create(&review).Error
}
