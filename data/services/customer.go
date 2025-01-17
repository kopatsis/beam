package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/custhelp"
	"beam/data/services/orderhelp"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type CustomerService interface {
	AddCustomer(customer models.Customer) error
	GetCustomerByID(id int) (*models.Customer, error)
	UpdateCustomer(customer models.Customer) error
	DeleteCustomer(id int) error
	GetCustomerAndContacts(customerID int) (*models.Customer, []*models.Contact, error)
	GetPaymentMethodsCust(customerID int) ([]models.PaymentMethodStripe, error)
	AddAddressToCustomer(customerID int, contact *models.Contact, mutex *config.AllMutexes) error
	AddAddressAndRender(customerID int, contact *models.Contact, mutex *config.AllMutexes, isDefault bool) ([]*models.Contact, error, error)
	MakeAddressDefault(customerID, contactID int) error
	MakeAddressDefaultAndRender(customerID, contactID int) ([]*models.Contact, error, error)
	UpdateContact(customerID, contactID int, newContact *models.Contact, mutex *config.AllMutexes) error
	UpdateContactAndRender(customerID, contactID int, newContact *models.Contact, mutex *config.AllMutexes, isDefault bool) ([]*models.Contact, error, error)
	DeleteContact(customerID, contactID int) (int, error)
	DeleteContactAndRender(customerID, contactID int) ([]*models.Contact, error, error)
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

func (s *customerService) AddAddressToCustomer(customerID int, contact *models.Contact, mutex *config.AllMutexes) error {
	contact.CustomerID = customerID

	if err := custhelp.VerifyContact(contact, mutex); err != nil {
		return err
	}

	return s.customerRepo.AddContactToCustomer(contact)
}

// Actual contacts, contact update error, contact retrieval error
func (s *customerService) AddAddressAndRender(customerID int, contact *models.Contact, mutex *config.AllMutexes, isDefault bool) ([]*models.Contact, error, error) {
	updateErr := s.AddAddressToCustomer(customerID, contact, mutex)

	if updateErr == nil && isDefault {
		updateErr = s.customerRepo.UpdateCustomerDefault(customerID, contact.ID)
	}

	list, getErr := s.customerRepo.GetContactsWithDefault(customerID)

	return list, updateErr, getErr
}

func (s *customerService) MakeAddressDefault(customerID, contactID int) error {
	c, err := s.customerRepo.GetSingleContact(contactID)
	if err != nil {
		return err
	} else if c == nil {
		return fmt.Errorf("unable to find non-nil contact for id: %d", contactID)
	} else if c.CustomerID != customerID {
		return fmt.Errorf("customer id %d doesn't match contact customer id of %d for contact id: %d", customerID, c.CustomerID, contactID)
	}

	return s.customerRepo.UpdateCustomerDefault(customerID, contactID)
}

// Actual contacts, contact update error, contact retrieval error
func (s *customerService) MakeAddressDefaultAndRender(customerID, contactID int) ([]*models.Contact, error, error) {
	updateErr := s.MakeAddressDefault(customerID, contactID)

	list, getErr := s.customerRepo.GetContactsWithDefault(customerID)

	return list, updateErr, getErr
}

func (s *customerService) UpdateContact(customerID, contactID int, newContact *models.Contact, mutex *config.AllMutexes) error {
	newContact.CustomerID = customerID

	if err := custhelp.VerifyContact(newContact, mutex); err != nil {
		return err
	}

	oldContact, err := s.customerRepo.GetSingleContact(contactID)
	if err != nil {
		return err
	} else if oldContact.CustomerID != customerID {
		return fmt.Errorf("non matching customer id for update contacts, id provided: %d; id on contact: %d", customerID, oldContact.CustomerID)
	}

	newContact.ID = contactID

	return s.customerRepo.UpdateContact(*newContact)

}

// Actual contacts, contact update error, contact retrieval error
func (s *customerService) UpdateContactAndRender(customerID, contactID int, newContact *models.Contact, mutex *config.AllMutexes, isDefault bool) ([]*models.Contact, error, error) {
	updateErr := s.UpdateContact(customerID, contactID, newContact, mutex)

	if updateErr == nil && isDefault {
		updateErr = s.customerRepo.UpdateCustomerDefault(customerID, contactID)
	}

	list, getErr := s.customerRepo.GetContactsWithDefault(customerID)

	return list, updateErr, getErr
}

func (s *customerService) DeleteContact(customerID, contactID int) (int, error) {
	newDefaultID := 0

	list, err := s.customerRepo.GetContactsWithDefault(customerID)
	if err != nil {
		return newDefaultID, err
	}

	index := -1
	for i, c := range list {
		if c.ID == contactID {
			index = i
			break
		}
	}

	if index == -1 {
		return newDefaultID, fmt.Errorf("no contact of id: %d for customer: %d", contactID, customerID)
	}

	if index == 0 && len(list) > 1 {
		newDefaultID = list[1].ID
	}

	return newDefaultID, s.customerRepo.DeleteContact(contactID)
}

// Actual contacts, contact update error, contact retrieval error
func (s *customerService) DeleteContactAndRender(customerID, contactID int) ([]*models.Contact, error, error) {
	newDefault, updateErr := s.DeleteContact(customerID, contactID)

	if updateErr == nil && newDefault > 0 {
		updateErr = s.customerRepo.UpdateCustomerDefault(customerID, newDefault)
	}

	list, getErr := s.customerRepo.GetContactsWithDefault(customerID)

	return list, updateErr, getErr
}

func (s *customerService) CreateCustomer(customer *models.CustomerPost, firebaseID string) (*models.Customer, error) {
	validate := validator.New()
	err := validate.Struct(customer)
	if err != nil {
		return nil, err
	}

	cid, status, err := s.customerRepo.CheckFirebaseUID(firebaseID)
	if err != nil {
		return nil, err
	}

	if status == "Archived" {
		return nil, fmt.Errorf("inactive customer for firebase id: %s, with id: %d", firebaseID, cid)
	} else if status == "Active" {
		return nil, fmt.Errorf("active customer for firebase id: %s, with id: %d", firebaseID, cid)
	}

	newCust := &models.Customer{
		FirebaseUID: firebaseID,
		DefaultName: customer.DefaultName,
		Email:       customer.Email,
		EmailSubbed: customer.EmailSubbed,
		Status:      "Active",
	}

	if customer.PhoneNumber != nil {
		newCust.PhoneNumber = orderhelp.CopyString(customer.PhoneNumber)
	}

	if err := s.customerRepo.Create(*newCust); err != nil {
		return nil, err
	}

	return newCust, nil
}

// Update basic user info = default name, email, phone number**

// "Delete" a customer aka mark them as Removed
