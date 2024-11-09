package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type CustomerService interface {
	AddCustomer(customer models.Customer) error
	GetCustomerByID(id int) (*models.Customer, error)
	UpdateCustomer(customer models.Customer) error
	DeleteCustomer(id int) error
}

type customerService struct {
	customerRepo repositories.CustomerRepository
}

func NewCustomerService(customerRepo repositories.CustomerRepository) CustomerService {
	return &customerService{customerRepo: customerRepo}
}

func (s *customerService) AddCustomer(customer models.Customer) error {
	return s.customerRepo.Create(customer)
}

func (s *customerService) GetCustomerByID(id int) (*models.Customer, error) {
	return s.customerRepo.Read(id)
}

func (s *customerService) UpdateCustomer(customer models.Customer) error {
	return s.customerRepo.Update(customer)
}

func (s *customerService) DeleteCustomer(id int) error {
	return s.customerRepo.Delete(id)
}
