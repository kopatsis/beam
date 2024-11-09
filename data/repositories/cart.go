package repositories

import (
	"beam/data/models"

	"gorm.io/gorm"
)

type CartRepository interface {
	Create(cart models.Cart) error
	Read(id int) (*models.Cart, error)
	Update(cart models.Cart) error
	Delete(id int) error
}

type cartRepo struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepo{db: db}
}

func (r *cartRepo) Create(cart models.Cart) error {
	return r.db.Create(&cart).Error
}

func (r *cartRepo) Read(id int) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.First(&cart, id).Error
	return &cart, err
}

func (r *cartRepo) Update(cart models.Cart) error {
	return r.db.Save(&cart).Error
}

func (r *cartRepo) Delete(id int) error {
	return r.db.Delete(&models.Cart{}, id).Error
}
