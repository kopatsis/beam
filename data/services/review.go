package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"fmt"
)

type ReviewService interface {
	AddReview(customerID int, productID int, store string, stars int, justStar bool, subject string, body string, ps *productService, tools *config.Tools) (*models.Review, error)
	UpdateReview(customerID int, productID int, store string, stars int, justStar bool, subject string, body string, ps *productService, tools *config.Tools) (*models.Review, error)
	DeleteReview(customerID int, productID int, store string, ps *productService, tools *config.Tools) (*models.Review, error)
	GetReview(customerID int, productID int) (*models.Review, error)
	FirstThreeForProduct(customerID int, productID int) (firstThree []*models.Review, existingReview *models.Review, singleErr error, multiErr error)
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

func (s *reviewService) UpdateReview(customerID int, productID int, store string, stars int, justStar bool, subject string, body string, ps *productService, tools *config.Tools) (*models.Review, error) {
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
	if existingReview == nil {
		return nil, fmt.Errorf("review does not exist for customerID %d and productID %d", customerID, productID)
	}

	if stars > 5 {
		stars = 5
	} else if stars < 1 {
		stars = 1
	}

	oldStars := existingReview.Stars

	existingReview.Stars = stars
	existingReview.JustStar = justStar
	existingReview.Subject = subject
	existingReview.Body = body

	if err := s.reviewRepo.Update(existingReview); err != nil {
		return nil, err
	}

	go ps.UpdateRatings(productID, store, stars, oldStars, 0, tools)

	return existingReview, nil
}

func (s *reviewService) DeleteReview(customerID int, productID int, store string, ps *productService, tools *config.Tools) (*models.Review, error) {

	existingReview, err := s.reviewRepo.GetSingle(customerID, productID)
	if err != nil {
		return nil, err
	}
	if existingReview == nil {
		return nil, fmt.Errorf("review does not exist for customerID %d and productID %d", customerID, productID)
	}

	stars := existingReview.Stars

	if err := s.reviewRepo.Delete(existingReview.PK); err != nil {
		return nil, err
	}

	go ps.UpdateRatings(productID, store, stars, 0, -1, tools)

	return existingReview, nil
}

func (s *reviewService) GetReview(customerID int, productID int) (*models.Review, error) {
	return s.reviewRepo.GetSingle(customerID, productID)
}

// First 3 featured, customer review, error for featured, error for customer
func (s *reviewService) FirstThreeForProduct(customerID int, productID int) (firstThree []*models.Review, existingReview *models.Review, singleErr error, multiErr error) {

	existingReview, singleErr = s.reviewRepo.GetSingle(customerID, productID)

	if existingReview != nil {
		firstThree, multiErr = s.reviewRepo.GetReviewsByProduct(productID, 0, 3, "stars", true, existingReview.CustomerID)
	} else {
		firstThree, multiErr = s.reviewRepo.GetReviewsByProduct(productID, 0, 3, "stars", true, 0)
	}

	return firstThree, existingReview, multiErr, singleErr
}
