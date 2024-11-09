package repositories

import (
	"beam/data/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product models.Product) error
	Read(id int) (*models.Product, error)
	Update(product models.Product) error
	Delete(id int) error
}

type productRepo struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewProductRepository(db *gorm.DB, rdb *redis.Client) ProductRepository {
	return &productRepo{db: db, rdb: rdb}
}

func (r *productRepo) Create(product models.Product) error {
	return r.db.Create(&product).Error
}

func (r *productRepo) Read(id int) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	return &product, err
}

func (r *productRepo) Update(product models.Product) error {
	return r.db.Save(&product).Error
}

func (r *productRepo) Delete(id int) error {
	return r.db.Delete(&models.Product{}, id).Error
}
