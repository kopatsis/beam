package repositories

import (
	"beam/data/models"

	"gorm.io/gorm"
)

type CustomerRepository interface {
	Create(customer models.Customer) error
	Read(id int) (*models.Customer, error)
	Update(customer models.Customer) error
	Delete(id int) error
}

type customerRepo struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepo{db: db}
}

func (r *customerRepo) Create(customer models.Customer) error {
	return r.db.Create(&customer).Error
}

func (r *customerRepo) Read(id int) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.First(&customer, id).Error
	return &customer, err
}

func (r *customerRepo) Update(customer models.Customer) error {
	return r.db.Save(&customer).Error
}

func (r *customerRepo) Delete(id int) error {
	return r.db.Delete(&models.Customer{}, id).Error
}
