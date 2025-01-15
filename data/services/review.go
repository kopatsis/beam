package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"fmt"
)

type ReviewService interface {
	AddReview(customerID int, productID int, store string, stars int, justStar bool, subject string, body string, ps *productService, tools *config.Tools) (*models.Review, error)
}

type reviewService struct {
	reviewRepo repositories.ReviewRepository
}

func NewReviewService(reviewRepo repositories.ReviewRepository) ReviewService {
	return &reviewService{reviewRepo: reviewRepo}
}

func (s *reviewService) AddReview(customerID int, productID int, store string, stars int, justStar bool, subject string, body string, ps *productService, tools *config.Tools) (*models.Review, error) {
	if len(subject) > 280 {
		subject = subject[:277] + "..."
	}
	if len(body) > 1400 {
		body = body[:1397] + "..."
	}

	existingReview, err := s.reviewRepo.GetSingle(customerID, productID)
	if err != nil {
		return nil, err
	}
	if existingReview != nil {
		return nil, fmt.Errorf("review already exists for customerID %d and productID %d", customerID, productID)
	}

	if stars > 5 {
		stars = 5
	} else if stars < 1 {
		stars = 1
	}

	review := &models.Review{
		CustomerID: customerID,
		ProductID:  productID,
		Stars:      stars,
		JustStar:   justStar,
		Subject:    subject,
		Body:       body,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}

	go ps.UpdateRatings(productID, store, stars, 0, 1, tools)

	return review, nil
}
