package repositories

import (
	"beam/data/models"
	"fmt"
	"time"

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

	SetReviewFeedback(customerID, reviewID int, helpful bool) error
	UnsetReviewFeedback(customerID, reviewID int) error

	// INT: 0 = none found, 1+ = helpful, -1- = unhelpful; error
	GetReviewFeedback(reviewID, customerID int) (int, error)
	GetReviewFeedbackMulti(reviewIDs []int, customerID int) (map[int]int, error)
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

func (r *reviewRepo) SetReviewFeedback(customerID, reviewID int, helpful bool) error {
	var existingFeedback models.ReviewFeedback
	err := r.db.Where("review_id = ? AND customer_id = ?", reviewID, customerID).First(&existingFeedback).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == nil && existingFeedback.Helpful == helpful {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		feedback := models.ReviewFeedback{
			ReviewID:   reviewID,
			CustomerID: customerID,
			Assigned:   time.Now(),
			Helpful:    helpful,
		}

		if err := tx.Save(&feedback).Error; err != nil {
			return err
		}

		var helpfulChange, unhelpfulChange int
		if err == gorm.ErrRecordNotFound {
			if helpful {
				helpfulChange = 1
			} else {
				unhelpfulChange = 1
			}
		} else {
			if helpful {
				if !existingFeedback.Helpful {
					helpfulChange = 1
					unhelpfulChange = -1
				}
			} else {
				if existingFeedback.Helpful {
					helpfulChange = -1
					unhelpfulChange = 1
				}
			}
		}

		if helpfulChange != 0 || unhelpfulChange != 0 {
			if err := tx.Model(&models.Review{}).
				Where("id = ?", reviewID).
				Updates(map[string]interface{}{
					"helpful":   gorm.Expr("helpful + ?", helpfulChange),
					"unhelpful": gorm.Expr("unhelpful + ?", unhelpfulChange),
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *reviewRepo) UnsetReviewFeedback(customerID, reviewID int) error {
	var existingFeedback models.ReviewFeedback
	err := r.db.Where("customer_id = ? AND review_id = ?", customerID, reviewID).First(&existingFeedback).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == gorm.ErrRecordNotFound {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var helpfulChange, unhelpfulChange int

		if existingFeedback.Helpful {
			helpfulChange = -1
		} else {
			unhelpfulChange = -1
		}

		if helpfulChange != 0 || unhelpfulChange != 0 {
			if err := tx.Model(&models.Review{}).
				Where("id = ?", reviewID).
				Updates(map[string]interface{}{"helpful": gorm.Expr("helpful + ?", helpfulChange), "unhelpful": gorm.Expr("unhelpful + ?", unhelpfulChange)}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("customer_id = ? AND review_id = ?", customerID, reviewID).Delete(&models.ReviewFeedback{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *reviewRepo) GetReviewFeedback(reviewID, customerID int) (int, error) {
	var feedback models.ReviewFeedback
	err := r.db.Where("review_id = ? AND customer_id = ?", reviewID, customerID).First(&feedback).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if feedback.Helpful {
		return 1, nil
	}
	return -1, nil
}

func (r *reviewRepo) GetReviewFeedbackMulti(reviewIDs []int, customerID int) (map[int]int, error) {
	var feedbacks []models.ReviewFeedback
	err := r.db.Where("review_id IN ? AND customer_id = ?", reviewIDs, customerID).Find(&feedbacks).Error
	if err != nil {
		return nil, err
	}

	result := map[int]int{}
	for _, feedback := range feedbacks {
		if feedback.Helpful {
			result[feedback.ReviewID] = 1
		} else {
			result[feedback.ReviewID] = -1
		}
	}

	return result, nil
}
