package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/reviewhelp"
	"errors"
	"fmt"
	"log"
	"net/url"
)

type ReviewService interface {
	AddReview(dpi *DataPassIn, productID int, store string, stars int, justStar, useDefaultName bool, displayName, subject, body string, ps ProductService, cs CustomerService, tools *config.Tools) (*models.Review, error)
	UpdateReview(dpi *DataPassIn, productID int, store string, stars int, justStar, useDefaultName bool, displayName, subject, body string, ps ProductService, cs CustomerService, tools *config.Tools) (*models.Review, error)
	DeleteReview(dpi *DataPassIn, productID int, store string, ps ProductService, tools *config.Tools) (*models.Review, error)
	GetReview(dpi *DataPassIn, productID int) (*models.Review, error)
	FirstThreeForProduct(dpi *DataPassIn, productID int) (firstThree []*models.Review, existingReview *models.Review, singleErr error, multiErr error)
	ReviewsByProduct(dpi *DataPassIn, productID int, fromURL url.Values) (ret models.ReviewPageRender, singleErr error, multiErr error)
	ReviewsByCustomer(dpi *DataPassIn, fromURL url.Values) (models.ReviewPageRender, error)
	GetReviewIDOnly(dpi *DataPassIn, ID int) (*models.Review, error)
	GetReviewsForOrder(dpi *DataPassIn, order *models.Order) (map[int]*models.Review, error)
}

type reviewService struct {
	reviewRepo repositories.ReviewRepository
}

func NewReviewService(reviewRepo repositories.ReviewRepository) ReviewService {
	return &reviewService{reviewRepo: reviewRepo}
}

func (s *reviewService) AddReview(dpi *DataPassIn, productID int, store string, stars int, justStar, useDefaultName bool, displayName, subject, body string, ps ProductService, cs CustomerService, tools *config.Tools) (*models.Review, error) {
	if len(subject) > 280 {
		subject = subject[:277] + "..."
	}
	if len(body) > 1400 {
		body = body[:1397] + "..."
	}

	if useDefaultName {
		cust, err := cs.GetCustomerByID(dpi.CustomerID)
		if err == nil {
			displayName = cust.FirstName
		} else {
			log.Printf("Unable to get customer by customerid: %d", dpi.CustomerID)

		}
	}

	if displayName == "" {
		displayName = "Anonymous"
	}

	if len(displayName) > 140 {
		displayName = displayName[:140]
	}

	existingReview, err := s.reviewRepo.GetSingle(dpi.CustomerID, productID)
	if err != nil {
		return nil, err
	}
	if existingReview != nil {
		return nil, fmt.Errorf("review already exists for customerID %d and productID %d", dpi.CustomerID, productID)
	}

	if stars > 5 {
		stars = 5
	} else if stars < 1 {
		stars = 1
	}

	review := &models.Review{
		CustomerID:  dpi.CustomerID,
		ProductID:   productID,
		DisplayName: displayName,
		Stars:       stars,
		JustStar:    justStar,
		Subject:     subject,
		Body:        body,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}

	go ps.UpdateRatings(dpi, productID, stars, 0, 1, tools)

	return review, nil
}

func (s *reviewService) UpdateReview(dpi *DataPassIn, productID int, store string, stars int, justStar, useDefaultName bool, displayName, subject, body string, ps ProductService, cs CustomerService, tools *config.Tools) (*models.Review, error) {
	if len(subject) > 280 {
		subject = subject[:277] + "..."
	}
	if len(body) > 1400 {
		body = body[:1397] + "..."
	}

	if useDefaultName {
		cust, err := cs.GetCustomerByID(dpi.CustomerID)
		if err == nil {
			displayName = cust.FirstName
		} else {
			log.Printf("Unable to get customer by customerid: %d", dpi.CustomerID)

		}
	}

	if displayName == "" {
		displayName = "Anonymous"
	}

	if len(displayName) > 140 {
		displayName = displayName[:140]
	}

	existingReview, err := s.reviewRepo.GetSingle(dpi.CustomerID, productID)
	if err != nil {
		return nil, err
	}
	if existingReview == nil {
		return nil, fmt.Errorf("review does not exist for customerID %d and productID %d", dpi.CustomerID, productID)
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
	existingReview.DisplayName = displayName

	if err := s.reviewRepo.Update(existingReview); err != nil {
		return nil, err
	}

	go ps.UpdateRatings(dpi, productID, stars, oldStars, 0, tools)

	return existingReview, nil
}

func (s *reviewService) DeleteReview(dpi *DataPassIn, productID int, store string, ps ProductService, tools *config.Tools) (*models.Review, error) {

	existingReview, err := s.reviewRepo.GetSingle(dpi.CustomerID, productID)
	if err != nil {
		return nil, err
	}
	if existingReview == nil {
		return nil, fmt.Errorf("review does not exist for customerID %d and productID %d", dpi.CustomerID, productID)
	}

	stars := existingReview.Stars

	if err := s.reviewRepo.Delete(existingReview.PK); err != nil {
		return nil, err
	}

	go ps.UpdateRatings(dpi, productID, stars, 0, -1, tools)

	return existingReview, nil
}

func (s *reviewService) GetReview(dpi *DataPassIn, productID int) (*models.Review, error) {
	return s.reviewRepo.GetSingle(dpi.CustomerID, productID)
}

// First 3 featured, customer review, error for featured, error for customer
func (s *reviewService) FirstThreeForProduct(dpi *DataPassIn, productID int) (firstThree []*models.Review, existingReview *models.Review, singleErr error, multiErr error) {

	existingReview, singleErr = s.reviewRepo.GetSingle(dpi.CustomerID, productID)

	if existingReview != nil {
		firstThree, multiErr = s.reviewRepo.GetReviewsByProduct(productID, 0, 3, "stars", true, existingReview.CustomerID)
	} else {
		firstThree, multiErr = s.reviewRepo.GetReviewsByProduct(productID, 0, 3, "stars", true, 0)
	}

	return firstThree, existingReview, multiErr, singleErr
}

func (s *reviewService) ReviewsByProduct(dpi *DataPassIn, productID int, fromURL url.Values) (ret models.ReviewPageRender, singleErr error, multiErr error) {

	var existingReview *models.Review
	var allReviews []*models.Review

	sort, desc, page := reviewhelp.ParseQueryParams(fromURL)

	perPage := config.REVIEWLEN

	offset := (perPage * page) - perPage

	existingReview, singleErr = s.reviewRepo.GetSingle(dpi.CustomerID, productID)

	allReviews, multiErr = s.reviewRepo.GetReviewsByProduct(productID, offset, perPage+1, sort, desc, 0)

	more := false
	if len(allReviews) > perPage {
		allReviews = allReviews[:perPage]
		more = true
	}
	less := perPage > 1

	ret.AllReviews = allReviews
	ret.CustReview = existingReview
	ret.Next = more
	ret.Previous = less
	ret.SortColumn = sort
	ret.Descending = desc

	return ret, multiErr, singleErr
}

func (s *reviewService) ReviewsByCustomer(dpi *DataPassIn, fromURL url.Values) (models.ReviewPageRender, error) {

	var ret models.ReviewPageRender

	sort, desc, page := reviewhelp.ParseQueryParams(fromURL)

	perPage := config.REVIEWLEN

	offset := (perPage * page) - perPage

	allReviews, err := s.reviewRepo.GetReviewsByCustomer(dpi.CustomerID, offset, perPage+1, sort, desc)

	more := false
	if len(allReviews) > perPage {
		allReviews = allReviews[:perPage]
		more = true
	}
	less := perPage > 1

	ret.AllReviews = allReviews
	ret.Next = more
	ret.Previous = less
	ret.SortColumn = sort
	ret.Descending = desc

	return ret, err
}

func (s *reviewService) GetReviewIDOnly(dpi *DataPassIn, ID int) (*models.Review, error) {
	r, err := s.reviewRepo.GetSingleByID(ID)
	if err != nil {
		return nil, err
	} else if r == nil {
		return nil, errors.New("empty review")
	} else if r.CustomerID != dpi.CustomerID {
		return nil, fmt.Errorf("review doesn't belong to customer: %d", dpi.CustomerID)
	}
	return r, nil
}

func (s *reviewService) GetReviewsForOrder(dpi *DataPassIn, order *models.Order) (map[int]*models.Review, error) {
	pids := map[int]struct{}{}
	for _, l := range order.Lines {
		pids[l.ProductID] = struct{}{}
	}

	list := []int{}
	for id := range pids {
		list = append(list, id)
	}

	return s.reviewRepo.GetReviewsMultiProduct(list, order.CustomerID)
}
