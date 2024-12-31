package repositories

import (
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
	CheckFavesLine(customerID int, variantID int) (int, time.Time, error)
	CheckSavesList(customerID int, variantID int) (int, time.Time, error)
	CheckLastOrdersList(customerID int, variantID int) (int, time.Time, error)
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

func (r *listRepo) CheckFavesLine(customerID int, variantID int) (int, time.Time, error) {
	var favesLine models.FavesLine
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).First(&favesLine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, time.Time{}, nil
		}
		return 0, time.Time{}, err
	}
	return favesLine.ID, favesLine.AddDate, nil
}

func (r *listRepo) CheckSavesList(customerID int, variantID int) (int, time.Time, error) {
	var savesList models.SavesList
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).First(&savesList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, time.Time{}, nil
		}
		return 0, time.Time{}, err
	}
	return savesList.ID, savesList.AddDate, nil
}

func (r *listRepo) CheckLastOrdersList(customerID int, variantID int) (int, time.Time, error) {
	var lastOrdersList models.LastOrdersList
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).First(&lastOrdersList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, time.Time{}, nil
		}
		return 0, time.Time{}, err
	}
	return lastOrdersList.ID, lastOrdersList.LastOrder, nil
}
