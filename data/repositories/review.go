package repositories

import (
	"beam/data/models"
	"fmt"

	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(review *models.Review) error
	Update(review *models.Review) error
	Delete(ID int) error

	GetSingle(customerID int, productID int) (*models.Review, error)
	GetSingleByID(ID int) (*models.Review, error)
	GetReviewsByCustomer(customerID, offset, limit int, sortColumn string, desc bool) ([]*models.Review, error)
	GetReviewsByProduct(productID, offset, limit int, sortColumn string, desc bool, custBlock int) ([]*models.Review, error)
	GetReviewsMultiProduct(productIDs []int, customerID int) (map[int]*models.Review, error)

	SetReviewFeedback(customerID, reviewID int, helpful bool) (*models.Review, error)
	UnsetReviewFeedback(customerID, reviewID int) (*models.Review, error)
}

type reviewRepo struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepo{db: db}
}

func (r *reviewRepo) Create(review *models.Review) error {
	return r.db.Create(review).Error
}

func (r *reviewRepo) Update(review *models.Review) error {
	return r.db.Save(review).Error
}

func (r *reviewRepo) Delete(ID int) error {
	return r.db.Delete(&models.Review{}, ID).Error
}

func (r *reviewRepo) GetSingle(customerID int, productID int) (*models.Review, error) {
	var reviews []models.Review
	if err := r.db.Where("customer_id = ? AND product_id = ?", customerID, productID).Find(&reviews).Error; err != nil {
		return nil, err
	}
	if len(reviews) > 1 {
		return &reviews[0], fmt.Errorf("more than one review found for customerID %d and productID %d", customerID, productID)
	}
	if len(reviews) == 0 {
		return nil, nil
	}
	return &reviews[0], nil
}

func (r *reviewRepo) GetSingleByID(ID int) (*models.Review, error) {
	var review models.Review
	if err := r.db.First(&review, ID).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepo) GetReviewsByCustomer(customerID, offset, limit int, sortColumn string, desc bool) ([]*models.Review, error) {
	var reviews []*models.Review
	order := sortColumn

	if desc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	if sortColumn != "created_at" {
		order += ", created_at DESC"
	} else {
		order += ", stars DESC"
	}

	if err := r.db.Where("customer_id = ? AND just_star = false", customerID).
		Order(order).
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error; err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *reviewRepo) GetReviewsByProduct(productID, offset, limit int, sortColumn string, desc bool, custBlock int) ([]*models.Review, error) {
	var reviews []*models.Review
	order := sortColumn

	if desc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	if sortColumn != "created_at" {
		order += ", created_at DESC"
	} else {
		order += ", stars DESC"
	}

	query := r.db.Where("product_id = ? AND just_star = false", productID)

	if custBlock > 0 {
		query = query.Where("customer_id != ?", custBlock)
	}

	if err := query.Order(order).
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error; err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *reviewRepo) GetReviewsMultiProduct(productIDs []int, customerID int) (map[int]*models.Review, error) {
	reviews := make(map[int]*models.Review)
	var result []models.Review
	if err := r.db.Where("product_id IN ? AND customer_id = ?", productIDs, customerID).Find(&result).Error; err != nil {
		return nil, err
	}
	for _, review := range result {
		reviews[review.ProductID] = &review
	}
	for _, productID := range productIDs {
		if _, exists := reviews[productID]; !exists {
			reviews[productID] = nil
		}
	}
	return reviews, nil
}

func (r *reviewRepo) SetReviewFeedback(customerID, reviewID int, helpful bool) (*models.Review, error) {
	var review models.Review
	if err := r.db.First(&review, reviewID).Error; err != nil {
		return nil, err
	}

	check := review.CheckCust(customerID)
	if (check != 1 && helpful) || (check != -1 && !helpful) {
		review.SetCust(customerID, helpful)
		return &review, r.db.Save(&review).Error
	}

	return &review, nil
}

func (r *reviewRepo) UnsetReviewFeedback(customerID, reviewID int) (*models.Review, error) {
	var review models.Review
	if err := r.db.First(&review, reviewID).Error; err != nil {
		return nil, err
	}

	if review.CheckCust(customerID) != 0 {
		review.UnsetCust(customerID)
		return &review, r.db.Save(&review).Error
	}

	return &review, nil
}
