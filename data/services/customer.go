package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/custhelp"
	"beam/data/services/orderhelp"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CustomerService interface {
	AddCustomer(customer models.Customer) error
	GetCustomerByID(id int) (*models.Customer, error)
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
	CreateCustomer(customer *models.CustomerPost, firebaseID string, store string) (*models.Customer, *models.ServerCookie, error)
	DeleteCustomer(custID int, store string) (*models.Customer, error)
	UpdateCustomer(customer *models.CustomerPost, custID int) (*models.Customer, error)
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

func (s *customerService) CreateCustomer(customer *models.CustomerPost, firebaseID string, store string) (*models.Customer, *models.ServerCookie, error) {
	validate := validator.New()
	err := validate.Struct(customer)
	if err != nil {
		return nil, nil, err
	}

	cid, status, err := s.customerRepo.CheckFirebaseUID(firebaseID)
	if err != nil {
		return nil, nil, err
	}

	if status == "Archived" {
		return nil, nil, fmt.Errorf("inactive customer for firebase id: %s, with id: %d", firebaseID, cid)
	} else if status == "Active" {
		return nil, nil, fmt.Errorf("active customer for firebase id: %s, with id: %d", firebaseID, cid)
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
		return nil, nil, err
	}

	c, err := s.customerRepo.CreateServerCookie(newCust.ID, firebaseID, store)

	return newCust, c, err
}

func (s *customerService) DeleteCustomer(custID int, store string) (*models.Customer, error) {
	cust, err := s.customerRepo.Read(custID)
	if err != nil {
		return nil, err
	}

	if cust.Status == "Archived" {
		return nil, fmt.Errorf("already archived customer for id: %d", custID)
	}

	cust.Status = "Archived"
	if err := s.customerRepo.Update(*cust); err != nil {
		return nil, err
	}

	cookie, err := s.customerRepo.GetServerCookie(custID, store)
	if err != nil {
		// log it
		return nil, err
	}

	_, err = s.customerRepo.SetServerCookieStatus(cookie, true)

	return cust, err
}

func (s *customerService) UpdateCustomer(customer *models.CustomerPost, custID int) (*models.Customer, error) {
	validate := validator.New()
	err := validate.Struct(customer)
	if err != nil {
		return nil, err
	}

	cust, err := s.customerRepo.Read(custID)
	if err != nil {
		return nil, err
	}

	if cust.Status == "Archived" {
		return nil, fmt.Errorf("inactive customer for id: %d", custID)
	}

	cust.Email = customer.Email
	cust.EmailSubbed = customer.EmailSubbed
	cust.DefaultName = customer.DefaultName
	cust.PhoneNumber = customer.PhoneNumber

	if err := s.customerRepo.Update(*cust); err != nil {
		return nil, err
	}

	return cust, nil
}

func (s *customerService) LoginCookie(firebaseID, store string, guestID string) (*models.ClientCookie, error) {
	customer, err := s.customerRepo.GetCustomerByFirebase(firebaseID)
	if err != nil {
		return nil, err
	} else if customer == nil {
		return nil, fmt.Errorf("no active customer for firebase ID: %s", firebaseID)
	} else if customer.Status == "Archived" {
		return nil, fmt.Errorf("archived customer for firebase ID: %s", firebaseID)
	}

	serverCookie, err := s.customerRepo.GetServerCookie(customer.ID, store)
	if err != nil {
		return nil, err
	} else if serverCookie == nil {
		return nil, fmt.Errorf("no active server cookie for firebase ID: %s; customer id: %d; store: %s", firebaseID, customer.ID, store)
	} else if customer.Status == "Archived" {
		return nil, fmt.Errorf("archived server cookie for firebase ID: %s; customer id: %d; store: %s", firebaseID, customer.ID, store)
	}

	return &models.ClientCookie{
		Store:       store,
		CustomerID:  customer.ID,
		CustomerSet: time.Now(),
		GuestID:     guestID,
	}, nil
}

func (s *customerService) ResetPass(firebaseID, store string, guestID string) error {
	customer, err := s.customerRepo.GetCustomerByFirebase(firebaseID)
	if err != nil {
		return err
	} else if customer == nil {
		return fmt.Errorf("no active customer for firebase ID: %s", firebaseID)
	} else if customer.Status == "Archived" {
		return fmt.Errorf("archived customer for firebase ID: %s", firebaseID)
	}

	serverCookie, err := s.customerRepo.GetServerCookie(customer.ID, store)
	if err != nil {
		return err
	} else if serverCookie == nil {
		return fmt.Errorf("no active server cookie for firebase ID: %s; customer id: %d; store: %s", firebaseID, customer.ID, store)
	} else if customer.Status == "Archived" {
		return fmt.Errorf("archived server cookie for firebase ID: %s; customer id: %d; store: %s", firebaseID, customer.ID, store)
	}

	reset := time.Now()
	customer.LastReset = reset
	sqlErr := s.customerRepo.Update(*customer)

	_, redisErr := s.customerRepo.SetServerCookieReset(serverCookie, reset)

	if redisErr != nil {
		return redisErr
	}
	return sqlErr
}

func (s *customerService) CustomerMiddleware(cookie *models.ClientCookie) error {

	if cookie.CustomerID > 0 {
		serverCookie, err := s.customerRepo.GetServerCookie(cookie.CustomerID, cookie.Store)
		if err != nil {
			customer, err := s.customerRepo.Read(cookie.CustomerID)
			if err != nil {
				return err
			} else if customer == nil {
				return fmt.Errorf("no active customer for customer id: %d; store: %s", cookie.CustomerID, cookie.Store)
			}
			if customer.Status == "Archived" || customer.LastReset.After(cookie.CustomerSet) {
				cookie.CustomerID = 0
				cookie.CustomerSet = time.Time{}
			}
		} else if serverCookie == nil {
			return fmt.Errorf("no active server cookie for customer id: %d; store: %s", cookie.CustomerID, cookie.Store)
		}

		if serverCookie.Archived || serverCookie.LastReset.After(cookie.CustomerSet) {
			cookie.CustomerID = 0
			cookie.CustomerSet = time.Time{}
		}
	}
	return nil
}

func (s *customerService) GuestMiddleware(cookie *models.ClientCookie, store string) {
	cookie.Store = store

	if cookie.GuestID == "" {
		cookie.GuestID = fmt.Sprintf("GI:%s", uuid.New().String())
	}
}

func (s *customerService) FullMiddleware(cookie *models.ClientCookie, store string) error {
	s.GuestMiddleware(cookie, store)
	return s.CustomerMiddleware(cookie)
}

func (s *customerService) LogoutCookie(cookie *models.ClientCookie) {
	cookie.CustomerID = 0
	cookie.CustomerSet = time.Time{}
}
