package repositories

import (
	"beam/data/models"

	"gorm.io/gorm"
)

type DiscountRepository interface {
	Create(discount models.Discount) error
	Read(id int) (*models.Discount, error)
	Update(discount models.Discount) error
	Delete(id int) error
}

type discountRepo struct {
	db *gorm.DB
}

func NewDiscountRepository(db *gorm.DB) DiscountRepository {
	return &discountRepo{db: db}
}

func (r *discountRepo) Create(discount models.Discount) error {
	return r.db.Create(&discount).Error
}

func (r *discountRepo) Read(id int) (*models.Discount, error) {
	var discount models.Discount
	err := r.db.First(&discount, id).Error
	return &discount, err
}

func (r *discountRepo) Update(discount models.Discount) error {
	return r.db.Save(&discount).Error
}

func (r *discountRepo) Delete(id int) error {
	return r.db.Delete(&models.Discount{}, id).Error
}
