package services

import (
	"beam/data/models"
	"beam/data/repositories"
	"errors"
)

type CustomerService interface {
	AddCustomer(customer models.Customer) error
	GetCustomerByID(id int) (*models.Customer, error)
	UpdateCustomer(customer models.Customer) error
	DeleteCustomer(id int) error
	GetCustomerAndContacts(customerID int) (*models.Customer, []*models.Contact, error)
	GetPaymentMethodsCust(customerID int) ([]models.PaymentMethodStripe, error)
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

func (s *customerService) GetCustomerAndContacts(customerID int) (*models.Customer, []*models.Contact, error) {
	cust, cont, err := s.customerRepo.GetCustomerAndContacts(customerID)
	if err != nil {
		return nil, nil, err
	}

	if cust.Status != "Active" {
		return nil, nil, errors.New("deleted customer")
	}

	return cust, cont, err
}

func (s *customerService) GetPaymentMethodsCust(customerID int) ([]models.PaymentMethodStripe, error) {
	return s.customerRepo.GetPaymentMethodsCust(customerID)
}

// Create new address with mild validation (required fields + ex. state, country, list...) -> use repo func and repl draft order use of repl with this
// BUT have one version that just updates and one version that returns all addresses

// Set address as default: Confirm that address id exists, update cust, then grab addresses (use repo func to grab them, or edit/make new one)

// Update an address: run same validation then update it, if set as default then go through ^ process, else re-query all on own

// Delete an address: if !default then just delete and re-query, else grab next one and set that as default, then requery

// Update basic user info = default name, email, phone number**

// Create a customer with basic info ^ AND firebase provided, incl ensuring that firebase NOT already in.

// General check that firebase ID NOT already in.

// "Delete" a customer aka mark them as Removed
