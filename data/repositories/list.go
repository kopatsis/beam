package repositories

import (
	"beam/data/models"

	"gorm.io/gorm"
)

type ListRepository interface {
	Create(list models.List) error
	Read(id int) (*models.List, error)
	Update(list models.List) error
	Delete(id int) error
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
