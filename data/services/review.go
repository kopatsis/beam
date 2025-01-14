package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type ReviewService interface {
	AddReview(review *models.Review) error
}

type reviewService struct {
	reviewRepo repositories.ReviewRepository
}

func NewReviewService(reviewRepo repositories.ReviewRepository) ReviewService {
	return &reviewService{reviewRepo: reviewRepo}
}

func (s *reviewService) AddReview(review *models.Review) error {
	return s.reviewRepo.Create(review)
}
