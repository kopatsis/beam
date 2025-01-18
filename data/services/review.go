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
	"strconv"
)

type ReviewService interface {
	AddReview(customerID int, productID int, store string, stars int, justStar bool, subject string, body string, ps *productService, tools *config.Tools) (*models.Review, error)
	UpdateReview(customerID int, productID int, store string, stars int, justStar bool, subject string, body string, ps *productService, tools *config.Tools) (*models.Review, error)
	DeleteReview(customerID int, productID int, store string, ps *productService, tools *config.Tools) (*models.Review, error)
	GetReview(customerID int, productID int) (*models.Review, error)
	FirstThreeForProduct(customerID int, productID int) (firstThree []*models.Review, existingReview *models.Review, singleErr error, multiErr error)
	ReviewsByProduct(customerID int, productID int, fromURL url.Values) (ret models.ReviewPageRender, singleErr error, multiErr error)
	ReviewsByCustomer(customerID int, fromURL url.Values) (models.ReviewPageRender, error)
	GetReviewIDOnly(customerID int, ID int) (*models.Review, error)
	GetReviewsForOrder(order *models.Order) (map[int]*models.Review, error)
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

func (s *reviewService) ReviewsByProduct(customerID int, productID int, fromURL url.Values) (ret models.ReviewPageRender, singleErr error, multiErr error) {

	var existingReview *models.Review
	var allReviews []*models.Review

	sort, desc, page := reviewhelp.ParseQueryParams(fromURL)

	perPage := config.REVIEWLEN

	offset := (perPage * page) - perPage

	existingReview, singleErr = s.reviewRepo.GetSingle(customerID, productID)

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

func (s *reviewService) ReviewsByCustomer(customerID int, fromURL url.Values) (models.ReviewPageRender, error) {

	var ret models.ReviewPageRender

	sort, desc, page := reviewhelp.ParseQueryParams(fromURL)

	perPage := config.REVIEWLEN

	offset := (perPage * page) - perPage

	allReviews, err := s.reviewRepo.GetReviewsByCustomer(customerID, offset, perPage+1, sort, desc)

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

func (s *reviewService) GetReviewIDOnly(customerID int, ID int) (*models.Review, error) {
	r, err := s.reviewRepo.GetSingleByID(ID)
	if err != nil {
		return nil, err
	} else if r == nil {
		return nil, errors.New("empty review")
	} else if r.CustomerID != customerID {
		return nil, fmt.Errorf("review doesn't belong to customer: %d", customerID)
	}
	return r, nil
}

func (s *reviewService) GetReviewsForOrder(order *models.Order) (map[int]*models.Review, error) {
	pids := map[int]struct{}{}
	for _, l := range order.Lines {
		idInt, err := strconv.Atoi(l.ProductID)
		if err != nil {
			log.Printf("unable to convert product ID to int: %s\n", l.ProductID)
			continue
		}

		pids[idInt] = struct{}{}
	}

	list := []int{}
	for id := range pids {
		list = append(list, id)
	}

	return s.reviewRepo.GetReviewsMultiProduct(list, order.CustomerID)
}
