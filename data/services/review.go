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
	"strings"

	"github.com/lib/pq"
)

type ReviewService interface {
	AddReview(dpi *DataPassIn, productID int, store string, stars int, justStar, useDefaultName, public bool, displayName, subject, body string, imgs []models.IntermImage, ps ProductService, cs CustomerService, tools *config.Tools) (*models.Review, error)
	UpdateReview(dpi *DataPassIn, productID int, store string, stars int, justStar, useDefaultName, public bool, displayName, subject, body string, ps ProductService, cs CustomerService, tools *config.Tools) (*models.Review, error)
	DeleteReview(dpi *DataPassIn, productID int, store string, ps ProductService, tools *config.Tools) (*models.Review, error)

	GetReview(dpi *DataPassIn, productID int) (*models.Review, error)
	FirstThreeForProduct(dpi *DataPassIn, productID int) (firstThree []*models.Review, existingReview *models.Review, singleErr error, multiErr error)
	ReviewsByProduct(dpi *DataPassIn, productID int, fromURL url.Values) (ret models.ReviewPageRender, singleErr error, multiErr error)
	ReviewsByCustomer(dpi *DataPassIn, fromURL url.Values) (models.ReviewPageRender, error)
	GetReviewIDOnly(dpi *DataPassIn, ID int) (*models.Review, error)
	GetReviewsForOrder(dpi *DataPassIn, order *models.Order) (map[int]*models.Review, error)

	RateReviewHelpful(dpi *DataPassIn, reviewID int) (*models.Review, error)
	RateReviewUnelpful(dpi *DataPassIn, reviewID int) (*models.Review, error)
	UnrateReview(dpi *DataPassIn, reviewID int) (*models.Review, error)

	AddNewImage(dpi *DataPassIn, reviewID int, img models.IntermImage, tools *config.Tools) (*models.Review, error)
	RemoveImage(dpi *DataPassIn, reviewID int, imgID string, tools *config.Tools) (*models.Review, error)

	RetrieveDraftReviews(dpi *DataPassIn) ([]models.Review, error)
	SetReviewStatus(dpi *DataPassIn, activeIDs, inactiveIDs []int) error
}

type reviewService struct {
	reviewRepo repositories.ReviewRepository
}

func NewReviewService(reviewRepo repositories.ReviewRepository) ReviewService {
	return &reviewService{reviewRepo: reviewRepo}
}

func (s *reviewService) AddReview(dpi *DataPassIn, productID int, store string, stars int, justStar, useDefaultName, public bool, displayName, subject, body string, imgs []models.IntermImage, ps ProductService, cs CustomerService, tools *config.Tools) (*models.Review, error) {
	if len(subject) > 280 {
		subject = subject[:277] + "..."
	}
	if len(body) > 1400 {
		body = body[:1397] + "..."
	}

	if useDefaultName {
		cust, err := cs.GetCustomerByID(dpi, dpi.CustomerID)
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

	existingReview, err := s.reviewRepo.GetSingle(dpi.CustomerID, productID, true)
	if err != nil {
		dpi.AddLog("Review", "AddReview", "Error querying existing review not related to review found", "", err, models.EventPassInFinal{ProductID: productID})
		return nil, err
	}
	if existingReview != nil {
		dpi.AddLog("Review", "AddReview", "", "Review already exists for combination, no issue", nil, models.EventPassInFinal{ProductID: productID})
		return nil, fmt.Errorf("review already exists for customerID %d and productID %d", dpi.CustomerID, productID)
	}

	if stars > 5 {
		stars = 5
	} else if stars < 1 {
		stars = 1
	}

	var urls pq.StringArray
	for i, img := range imgs {
		if i > 2 {
			break
		}
		url, err := config.UploadToS3(tools.S3, img.FileNameNew, img.Data)
		if err != nil {
			/// Notify me S3 failure
			log.Printf("Failed to add image from S3: %s, Error: %v", img.FileNameNew, err)
		} else {
			urls = append(urls, url)
		}

	}

	review := &models.Review{
		CustomerID:  dpi.CustomerID,
		ProductID:   productID,
		DisplayName: displayName,
		Status:      "Draft",
		Public:      public,
		Stars:       stars,
		JustStar:    justStar,
		Subject:     subject,
		Body:        body,
		ImageURLs:   urls,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		dpi.AddLog("Review", "AddReview", "Error creating review", "", err, models.EventPassInFinal{ProductID: productID})
		return nil, err
	}

	go ps.UpdateRatings(dpi, productID, stars, 0, 1, tools)

	dpi.AddLog("Review", "AddReview", "", "", nil, models.EventPassInFinal{ProductID: productID, ReviewID: review.PK})
	return review, nil
}

func (s *reviewService) UpdateReview(dpi *DataPassIn, productID int, store string, stars int, justStar, useDefaultName, public bool, displayName, subject, body string, ps ProductService, cs CustomerService, tools *config.Tools) (*models.Review, error) {
	if len(subject) > 280 {
		subject = subject[:277] + "..."
	}
	if len(body) > 1400 {
		body = body[:1397] + "..."
	}

	if useDefaultName {
		cust, err := cs.GetCustomerByID(dpi, dpi.CustomerID)
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

	existingReview, err := s.reviewRepo.GetSingle(dpi.CustomerID, productID, true)
	if err != nil {
		dpi.AddLog("Review", "UpdateReview", "Unable to retrieve review unrelated to does not exist", "", err, models.EventPassInFinal{ProductID: productID})
		return nil, err
	}
	if existingReview == nil {
		dpi.AddLog("Review", "UpdateReview", "Review doesn't exist", "", fmt.Errorf("review does not exist for customerID %d and productID %d", dpi.CustomerID, productID), models.EventPassInFinal{ProductID: productID})
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
	existingReview.Public = public
	existingReview.Status = "Draft"

	if err := s.reviewRepo.Update(existingReview); err != nil {
		dpi.AddLog("Review", "UpdateReview", "Unable to update review", "", err, models.EventPassInFinal{ProductID: productID, ReviewID: existingReview.PK})
		return nil, err
	}

	go ps.UpdateRatings(dpi, productID, stars, oldStars, 0, tools)

	dpi.AddLog("Review", "UpdateReview", "", "", nil, models.EventPassInFinal{ProductID: productID, ReviewID: existingReview.PK})
	return existingReview, nil
}

func (s *reviewService) DeleteReview(dpi *DataPassIn, productID int, store string, ps ProductService, tools *config.Tools) (*models.Review, error) {

	existingReview, err := s.reviewRepo.GetSingle(dpi.CustomerID, productID, true)
	if err != nil {
		dpi.AddLog("Review", "DeleteReview", "Unable to retrieve review unrelated to does not exist", "", err, models.EventPassInFinal{ProductID: productID})
		return nil, err
	}
	if existingReview == nil {
		dpi.AddLog("Review", "DeleteReview", "Review doesn't exist", "", fmt.Errorf("review does not exist for customerID %d and productID %d", dpi.CustomerID, productID), models.EventPassInFinal{ProductID: productID})
		return nil, fmt.Errorf("review does not exist for customerID %d and productID %d", dpi.CustomerID, productID)
	}

	for _, imgURL := range existingReview.ImageURLs {
		if err := config.DeleteFromS3(tools.S3, imgURL); err != nil {
			// Notify me S3 failure
			log.Printf("Failed to delete image from S3: %s, Error: %v", imgURL, err)
		}
	}

	stars := existingReview.Stars

	if err := s.reviewRepo.Delete(existingReview.PK); err != nil {
		dpi.AddLog("Review", "DeleteReview", "Unable to delete review", "", err, models.EventPassInFinal{ProductID: productID})
		return nil, err
	}

	go ps.UpdateRatings(dpi, productID, stars, 0, -1, tools)

	dpi.AddLog("Review", "DeleteReview", "", "", nil, models.EventPassInFinal{ProductID: productID})
	return existingReview, nil
}

func (s *reviewService) GetReview(dpi *DataPassIn, productID int) (*models.Review, error) {
	if r, err := s.reviewRepo.GetSingle(dpi.CustomerID, productID, true); err != nil {
		dpi.AddLog("Review", "GetReview", "Error getting review", "", err, models.EventPassInFinal{ProductID: productID})
		return r, err
	} else {
		dpi.AddLog("Review", "GetReview", "", "", nil, models.EventPassInFinal{ProductID: productID})
		return r, err
	}
}

// First 3 featured, customer review, error for featured, error for customer
func (s *reviewService) FirstThreeForProduct(dpi *DataPassIn, productID int) (firstThree []*models.Review, existingReview *models.Review, singleErr error, multiErr error) {

	existingReview, singleErr = s.reviewRepo.GetSingle(dpi.CustomerID, productID, true)

	if existingReview != nil {
		firstThree, multiErr = s.reviewRepo.GetReviewsByProduct(productID, 0, 3, "stars", true, existingReview.CustomerID, false)
	} else {
		firstThree, multiErr = s.reviewRepo.GetReviewsByProduct(productID, 0, 3, "stars", true, 0, false)
	}

	if singleErr != nil {
		dpi.AddLog("Review", "FirstThreeForProduct", "Error getting single review", "", singleErr, models.EventPassInFinal{ProductID: productID})
	}

	if multiErr != nil {
		dpi.AddLog("Review", "FirstThreeForProduct", "Error getting first 3 reviews", "", multiErr, models.EventPassInFinal{ProductID: productID})
	}

	if multiErr == nil && singleErr == nil {
		dpi.AddLog("Review", "FirstThreeForProduct", "", "", nil, models.EventPassInFinal{ProductID: productID})
	}

	return firstThree, existingReview, multiErr, singleErr
}

func (s *reviewService) ReviewsByProduct(dpi *DataPassIn, productID int, fromURL url.Values) (ret models.ReviewPageRender, singleErr error, multiErr error) {

	var existingReview *models.Review
	var allReviews []*models.Review

	sort, desc, page := reviewhelp.ParseQueryParams(fromURL)

	perPage := config.REVIEWLEN

	offset := (perPage * page) - perPage

	existingReview, singleErr = s.reviewRepo.GetSingle(dpi.CustomerID, productID, true)

	allReviews, multiErr = s.reviewRepo.GetReviewsByProduct(productID, offset, perPage+1, sort, desc, 0, false)

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

	if singleErr != nil {
		dpi.AddLog("Review", "ReviewsByProduct", "Error getting single review", "", singleErr, models.EventPassInFinal{ProductID: productID})
	}

	if multiErr != nil {
		dpi.AddLog("Review", "ReviewsByProduct", "Error getting multiple reviews", "", multiErr, models.EventPassInFinal{ProductID: productID})
	}

	if multiErr == nil && singleErr == nil {
		dpi.AddLog("Review", "ReviewsByProduct", "", "", nil, models.EventPassInFinal{ProductID: productID})
	}

	return ret, multiErr, singleErr
}

func (s *reviewService) ReviewsByCustomer(dpi *DataPassIn, fromURL url.Values) (models.ReviewPageRender, error) {

	var ret models.ReviewPageRender

	sort, desc, page := reviewhelp.ParseQueryParams(fromURL)

	perPage := config.REVIEWLEN

	offset := (perPage * page) - perPage

	allReviews, err := s.reviewRepo.GetReviewsByCustomer(dpi.CustomerID, offset, perPage+1, sort, desc, true)

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

	if err != nil {
		dpi.AddLog("Review", "ReviewsByCustomer", "Error getting reviews", "", err, models.EventPassInFinal{})
	} else {
		dpi.AddLog("Review", "ReviewsByCustomer", "", "", nil, models.EventPassInFinal{})
	}

	return ret, err
}

func (s *reviewService) GetReviewIDOnly(dpi *DataPassIn, ID int) (*models.Review, error) {
	r, err := s.reviewRepo.GetSingleByID(ID, true)
	if err != nil {
		dpi.AddLog("Review", "GetReviewIDOnly", "Unable to retrieve review unrelated to does not exist", "", err, models.EventPassInFinal{ReviewID: ID})
		return nil, err
	} else if r == nil {
		dpi.AddLog("Review", "GetReviewIDOnly", "Review doesn't exist", "", errors.New("empty review"), models.EventPassInFinal{ReviewID: ID})
		return nil, errors.New("empty review")
	}

	if r.CustomerID != dpi.CustomerID && (!r.Public || r.Status != "Active") {
		dpi.AddLog("Review", "GetReviewIDOnly", "Review not accessible to requesting party", "", fmt.Errorf("review doesn't belong to customer: %d", dpi.CustomerID), models.EventPassInFinal{ReviewID: ID})
		return nil, fmt.Errorf("review doesn't belong to customer: %d", dpi.CustomerID)
	}

	dpi.AddLog("Review", "GetReviewIDOnly", "", "", nil, models.EventPassInFinal{ReviewID: ID})
	return r, nil
}

func (s *reviewService) GetReviewsForOrder(dpi *DataPassIn, order *models.Order) (map[int]*models.Review, error) {
	if order == nil {
		dpi.AddLog("Review", "GetReviewsForOrder", "No order to get reviews for", "", errors.New("nil order"), models.EventPassInFinal{})
		return nil, errors.New("nil order")
	}

	pids := map[int]struct{}{}
	for _, l := range order.Lines {
		pids[l.ProductID] = struct{}{}
	}

	list := []int{}
	for id := range pids {
		list = append(list, id)
	}

	if rs, err := s.reviewRepo.GetReviewsMultiProduct(list, order.CustomerID, true); err != nil {
		dpi.AddLog("Review", "GetReviewsForOrder", "Unable to get review map", "", err, models.EventPassInFinal{OrderID: order.ID.Hex()})
		return rs, err
	} else {
		dpi.AddLog("Review", "GetReviewsForOrder", "", "", nil, models.EventPassInFinal{OrderID: order.ID.Hex()})
		return rs, err
	}
}

func (s *reviewService) RateReviewHelpful(dpi *DataPassIn, reviewID int) (*models.Review, error) {
	return s.reviewRepo.SetReviewFeedback(dpi.CustomerID, reviewID, true)
}

func (s *reviewService) RateReviewUnelpful(dpi *DataPassIn, reviewID int) (*models.Review, error) {
	return s.reviewRepo.SetReviewFeedback(dpi.CustomerID, reviewID, false)
}

func (s *reviewService) UnrateReview(dpi *DataPassIn, reviewID int) (*models.Review, error) {
	return s.reviewRepo.UnsetReviewFeedback(dpi.CustomerID, reviewID)
}

func (s *reviewService) AddNewImage(dpi *DataPassIn, reviewID int, img models.IntermImage, tools *config.Tools) (*models.Review, error) {
	r, err := s.reviewRepo.GetSingleByID(reviewID, true)
	if err != nil {
		return nil, err
	} else if r == nil {
		return nil, errors.New("empty review")
	} else if r.CustomerID != dpi.CustomerID {
		return nil, fmt.Errorf("review doesn't belong to customer: %d", dpi.CustomerID)
	}

	fileName, err := config.UploadToS3(tools.S3, img.FileNameNew, img.Data)
	if err != nil {
		return nil, err
	}

	if len(r.ImageURLs) >= 3 {
		if err := config.DeleteFromS3(tools.S3, r.ImageURLs[2]); err != nil {
			// Notify me S3 failure
			log.Printf("Failed to delete image from S3: %s, Error: %v", r.ImageURLs[2], err)
		}
		r.ImageURLs = r.ImageURLs[:2]
	}

	r.ImageURLs = append(r.ImageURLs, fileName)
	r.Status = "Draft"

	return r, s.reviewRepo.Update(r)
}

func (s *reviewService) RemoveImage(dpi *DataPassIn, reviewID int, imgID string, tools *config.Tools) (*models.Review, error) {
	r, err := s.reviewRepo.GetSingleByID(reviewID, true)
	if err != nil {
		return nil, err
	} else if r == nil {
		return nil, errors.New("empty review")
	} else if r.CustomerID != dpi.CustomerID {
		return nil, fmt.Errorf("review doesn't belong to customer: %d", dpi.CustomerID)
	}

	index := -1
	for i, listedImg := range r.ImageURLs {
		if strings.Contains(listedImg, imgID) {
			index = i
		}
	}

	if index == -1 {
		return r, errors.New("no image with matching file name to id")
	}

	var newImageURLs pq.StringArray
	for i, listedImg := range r.ImageURLs {
		if i != index {
			newImageURLs = append(newImageURLs, listedImg)
		} else {
			if err := config.DeleteFromS3(tools.S3, r.ImageURLs[i]); err != nil {
				// Notify me S3 failure
				log.Printf("Failed to delete image from S3: %s, Error: %v", r.ImageURLs[2], err)
			}
		}
	}
	r.ImageURLs = newImageURLs

	return r, s.reviewRepo.Update(r)
}

func (s *reviewService) RetrieveDraftReviews(dpi *DataPassIn) ([]models.Review, error) {
	return s.reviewRepo.GetDraftReviews()
}

func (s *reviewService) SetReviewStatus(dpi *DataPassIn, activeIDs []int, inactiveIDs []int) error {
	return s.reviewRepo.UpdateReviewStatus(activeIDs, inactiveIDs)
}
