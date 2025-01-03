package repositories

import (
	"beam/config"
	"beam/data/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ListRepository interface {
	Create(list models.List) error
	Read(id int) (*models.List, error)
	Update(list models.List) error
	Delete(id int) error
	CheckFavesLine(customerID int, variantID int) (bool, *models.FavesLine, error)
	CheckSavesList(customerID int, variantID int) (bool, *models.SavesList, error)
	CheckLastOrdersList(customerID int, variantID int) (bool, *models.LastOrdersList, error)
	AddFavesLine(customerID, productID, variantID int) error
	AddSavesList(customerID, productID, variantID int) error
	AddLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error
	CheckLastOrdersListMultiVar(customerID int, variantIDs []int) (map[int]bool, error)
	DeleteLastOrdersListVariants(customerID int, variantIDs []int) error
	DeleteFavesLine(customerID, variantID int) error
	DeleteSavesList(customerID, variantID int) error
	UpdateLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error
	GetFavesLineByPage(customerID, page int) ([]*models.FavesLine, bool, bool, error)
	GetSavesListByPage(customerID, page int) ([]*models.SavesList, bool, bool, error)
	GetLastOrdersListByPage(customerID, page int) ([]*models.LastOrdersList, bool, bool, error)
}

type listRepo struct {
	db *gorm.DB
}

func NewListRepository(db *gorm.DB) ListRepository {
	return &listRepo{db: db}
}

func (r *listRepo) Create(list models.List) error {
	return r.db.Create(&list).Error
}

func (r *listRepo) Read(id int) (*models.List, error) {
	var list models.List
	err := r.db.First(&list, id).Error
	return &list, err
}

func (r *listRepo) Update(list models.List) error {
	return r.db.Save(&list).Error
}

func (r *listRepo) Delete(id int) error {
	return r.db.Delete(&models.List{}, id).Error
}

func (r *listRepo) CheckFavesLine(customerID int, variantID int) (bool, *models.FavesLine, error) {
	var favesLine models.FavesLine
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).First(&favesLine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil, nil
		}
		return false, nil, err
	}
	return false, &favesLine, nil
}

func (r *listRepo) CheckSavesList(customerID int, variantID int) (bool, *models.SavesList, error) {
	var savesList models.SavesList
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).First(&savesList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil, nil
		}
		return false, nil, err
	}
	return false, &savesList, nil
}

func (r *listRepo) CheckLastOrdersList(customerID int, variantID int) (bool, *models.LastOrdersList, error) {
	var lastOrdersList models.LastOrdersList
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).
		Order("last_order DESC").
		First(&lastOrdersList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil, nil
		}
		return false, nil, err
	}
	return false, &lastOrdersList, nil
}

func (r *listRepo) CheckLastOrdersListProd(customerID int, productID int) (bool, *models.LastOrdersList, error) {
	var lastOrdersList models.LastOrdersList
	if err := r.db.Where("customer_id = ? AND product_id = ?", customerID, productID).
		Order("last_order DESC").
		First(&lastOrdersList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil, nil
		}
		return false, nil, err
	}
	return false, &lastOrdersList, nil
}

func (r *listRepo) CheckLastOrdersListMultiVar(customerID int, variantIDs []int) (map[int]bool, error) {
	var existingVariants []int
	result := make(map[int]bool, len(variantIDs))

	if err := r.db.Model(&models.LastOrdersList{}).
		Select("variant_id").
		Where("customer_id = ? AND variant_id IN (?)", customerID, variantIDs).
		Find(&existingVariants).Error; err != nil {
		return nil, err
	}

	for _, id := range variantIDs {
		result[id] = false
	}
	for _, id := range existingVariants {
		result[id] = true
	}

	return result, nil
}

func (r *listRepo) AddFavesLine(customerID, productID, variantID int) error {
	favesLine := models.FavesLine{
		CustomerID: customerID,
		ProductID:  productID,
		VariantID:  variantID,
		AddDate:    time.Now(),
	}
	return r.db.Create(&favesLine).Error
}

func (r *listRepo) AddSavesList(customerID, productID, variantID int) error {
	savesList := models.SavesList{
		CustomerID: customerID,
		ProductID:  productID,
		VariantID:  variantID,
		AddDate:    time.Now(),
	}
	return r.db.Create(&savesList).Error
}

func (r *listRepo) AddLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error {
	var lastOrdersList []models.LastOrdersList
	for variantID, productID := range variants {
		lastOrdersList = append(lastOrdersList, models.LastOrdersList{
			CustomerID:  customerID,
			ProductID:   productID,
			VariantID:   variantID,
			LastOrder:   orderDate,
			LastOrderID: orderID,
		})
	}
	return r.db.Create(&lastOrdersList).Error
}

func (r *listRepo) DeleteFavesLine(customerID, variantID int) error {
	return r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).
		Delete(&models.FavesLine{}).Error
}

func (r *listRepo) DeleteSavesList(customerID, variantID int) error {
	return r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).
		Delete(&models.SavesList{}).Error
}

func (r *listRepo) DeleteLastOrdersListVariants(customerID int, variantIDs []int) error {
	return r.db.Where("customer_id = ? AND variant_id IN (?)", customerID, variantIDs).
		Delete(&models.LastOrdersList{}).Error
}

func (r *listRepo) UpdateLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error {
	variantIDs := make([]int, 0, len(variants))
	for variantID := range variants {
		variantIDs = append(variantIDs, variantID)
	}

	existingMap, err := r.CheckLastOrdersListMultiVar(customerID, variantIDs)
	if err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &listRepo{db: tx}

		toDelete := make([]int, 0)
		for id, exists := range existingMap {
			if exists {
				toDelete = append(toDelete, id)
			}
		}

		if len(toDelete) > 0 {
			if err := txRepo.DeleteLastOrdersListVariants(customerID, toDelete); err != nil {
				return err
			}
		}

		if err := txRepo.AddLastOrdersList(customerID, orderDate, orderID, variants); err != nil {
			return err
		}

		return nil
	})
}

func (r *listRepo) GetFavesLineByPage(customerID, page int) ([]*models.FavesLine, bool, bool, error) {
	limit := config.LIST_LIMIT
	offset := (page - 1)
	if offset < 0 {
		offset = 0
	}
	offset *= limit

	var faves []*models.FavesLine
	if err := r.db.Where("customer_id = ?", customerID).
		Order("add_date DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&faves).Error; err != nil {
		return nil, false, false, err
	}

	hasPrev := offset > 0
	hasNext := len(faves) > limit
	if hasNext {
		faves = faves[:limit]
	}

	return faves, hasPrev, hasNext, nil
}

func (r *listRepo) GetSavesListByPage(customerID, page int) ([]*models.SavesList, bool, bool, error) {
	limit := config.LIST_LIMIT
	offset := (page - 1)
	if offset < 0 {
		offset = 0
	}
	offset *= limit

	var saves []*models.SavesList
	if err := r.db.Where("customer_id = ?", customerID).
		Order("add_date DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&saves).Error; err != nil {
		return nil, false, false, err
	}

	hasPrev := offset > 0
	hasNext := len(saves) > limit
	if hasNext {
		saves = saves[:limit]
	}

	return saves, hasPrev, hasNext, nil
}

func (r *listRepo) GetLastOrdersListByPage(customerID, page int) ([]*models.LastOrdersList, bool, bool, error) {
	limit := config.LIST_LIMIT
	offset := (page - 1)
	if offset < 0 {
		offset = 0
	}
	offset *= limit

	var orders []*models.LastOrdersList
	if err := r.db.Where("customer_id = ?", customerID).
		Order("last_order DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, false, false, err
	}

	hasPrev := offset > 0
	hasNext := len(orders) > limit
	if hasNext {
		orders = orders[:limit]
	}

	return orders, hasPrev, hasNext, nil
}
