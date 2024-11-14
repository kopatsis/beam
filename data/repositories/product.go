package repositories

import (
	"beam/data/models"
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product models.Product) error
	Read(id int) (*models.Product, error)
	Update(product models.Product) error
	Delete(id int) error
	GetAllProductInfo(name string) ([]models.ProductInfo, error)
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

func (r *productRepo) GetAllProductInfo(name string) ([]models.ProductInfo, error) {
	key := name + "::PWC"
	var productInfo []models.ProductInfo
	data, err := r.rdb.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	err = json.Unmarshal([]byte(data), &productInfo)
	if err != nil {
		return nil, err
	}
	return productInfo, nil
}
