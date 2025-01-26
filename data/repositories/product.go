package repositories

import (
	"beam/data/models"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product models.Product) error
	Read(id int) (*models.Product, error) // U
	Update(product models.Product) error  // U
	Delete(id int) error
	GetAllProductInfo(name string) ([]models.ProductInfo, error)             // U
	GetFullProduct(name, handle string) (models.ProductRedis, string, error) // U
	GetLimVars(name string, vids []int) ([]*models.LimitedVariantRedis, error)
	GetFullProducts(name string, handles []string) ([]*models.ProductRedis, error)
	SaveProductInfoInTransaction(name string, prod *models.ProductRedis, info []models.ProductInfo) error // U
	ReadComparables(id int) ([]*models.Comparable, error)
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

func (r *productRepo) GetFullProduct(name, handle string) (models.ProductRedis, string, error) {
	key := name + "::PRO::" + handle
	var product models.ProductRedis

	data, err := r.rdb.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return product, "", errors.New("not in system at all")
		}
		return product, "", err
	}

	if strings.HasPrefix(data, "RDR::") {
		redirectHandle := strings.TrimPrefix(data, "RDR::")
		return product, redirectHandle, nil
	}

	err = json.Unmarshal([]byte(data), &product)
	if err != nil {
		return product, "", err
	}

	return product, "", nil
}

func (r *productRepo) GetLimVars(name string, vids []int) ([]*models.LimitedVariantRedis, error) {
	var keys []string
	for _, id := range vids {
		keys = append(keys, name+"::LVR::"+strconv.Itoa(id))
	}

	results, err := r.rdb.MGet(context.Background(), keys...).Result()
	if err != nil {
		return nil, err
	}

	var limitedVariants []*models.LimitedVariantRedis
	for _, result := range results {
		if result == nil {
			continue
		}
		var variant models.LimitedVariantRedis
		if err := json.Unmarshal([]byte(result.(string)), &variant); err != nil {
			return nil, err
		}
		limitedVariants = append(limitedVariants, &variant)
	}

	return limitedVariants, nil
}

func (r *productRepo) GetFullProducts(name string, handles []string) ([]*models.ProductRedis, error) {
	var products []*models.ProductRedis
	keys := make([]string, len(handles))

	for i, handle := range handles {
		keys[i] = name + "::PRO::" + handle
	}

	data, err := r.rdb.MGet(context.Background(), keys...).Result()
	if err != nil {
		return nil, err
	}

	for _, d := range data {
		var product models.ProductRedis

		if str, ok := d.(string); ok && strings.HasPrefix(str, "RDR::") {
			continue
		}

		if err := json.Unmarshal([]byte(d.(string)), &product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func (r *productRepo) SaveProductInfoInTransaction(name string, prod *models.ProductRedis, info []models.ProductInfo) error {
	prodKey := name + "::PRO::" + prod.Handle
	infoKey := name + "::PWC"
	prodData, err := json.Marshal(prod)
	if err != nil {
		return err
	}
	infoData, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return r.rdb.Watch(context.Background(), func(tx *redis.Tx) error {
		_, err := tx.Pipelined(context.Background(), func(pipe redis.Pipeliner) error {
			pipe.Set(context.Background(), prodKey, prodData, 0)
			pipe.Set(context.Background(), infoKey, infoData, 0)
			return nil
		})
		return err
	}, prodKey, infoKey)
}

func (r *productRepo) ReadComparables(id int) ([]*models.Comparable, error) {
	var comparables []*models.Comparable
	err := r.db.Where("pkfk_product_id1 = ? OR pkfk_product_id2 = ?", id, id).Find(&comparables).Error
	return comparables, err
}
