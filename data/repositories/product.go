package repositories

import (
	"beam/data/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product models.Product) error
	FindByID(id uint) (*models.Product, error)
	Update(product models.Product) error
	Delete(id uint) error
}

type productRepo struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewProductRepository(db *gorm.DB, redis *redis.Client) ProductRepository {
	return &productRepo{db: db, redis: redis}
}

func (r *productRepo) Create(product models.Product) error {
	return r.db.Create(&product).Error
}

func (r *productRepo) FindByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	return &product, err
}

func (r *productRepo) Update(product models.Product) error {
	return r.db.Save(&product).Error
}

func (r *productRepo) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}
