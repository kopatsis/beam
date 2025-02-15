package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/custhelp"
	"beam/data/services/orderhelp"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CustomerService interface {
	GetCustomerByID(id int) (*models.Customer, error)
	GetCustomerAndContacts(dpi *DataPassIn) (*models.Customer, []*models.Contact, error)
	GetPaymentMethodsCust(dpi *DataPassIn) ([]models.PaymentMethodStripe, error)
	AddAddressToCustomer(dpi *DataPassIn, contact *models.Contact, mutex *config.AllMutexes) error
	AddAddressAndRender(dpi *DataPassIn, contact *models.Contact, mutex *config.AllMutexes, isDefault bool) ([]*models.Contact, error, error)
	MakeAddressDefault(dpi *DataPassIn, contactID int) error
	MakeAddressDefaultAndRender(dpi *DataPassIn, contactID int) ([]*models.Contact, error, error)
	UpdateContact(dpi *DataPassIn, contactID int, newContact *models.Contact, mutex *config.AllMutexes) error
	UpdateContactAndRender(dpi *DataPassIn, contactID int, newContact *models.Contact, mutex *config.AllMutexes, isDefault bool) ([]*models.Contact, error, error)
	DeleteContact(dpi *DataPassIn, contactID int) (int, error)
	DeleteContactAndRender(dpi *DataPassIn, contactID int) ([]*models.Contact, error, error)
	CreateCustomer(dpi *DataPassIn, customer *models.CustomerPost, firebaseID string) (*models.Customer, *models.ServerCookie, error)
	DeleteCustomer(dpi *DataPassIn) (*models.Customer, error)
	UpdateCustomer(dpi *DataPassIn, customer *models.CustomerPost) (*models.Customer, error)

	LoginCookie(dpi *DataPassIn, firebaseID string) (*models.ClientCookie, error)
	ResetPass(dpi *DataPassIn, firebaseID string) error
	CustomerMiddleware(cookie *models.ClientCookie)
	GuestMiddleware(cookie *models.ClientCookie, store string)
	FullMiddleware(cookie *models.ClientCookie, store string)
	LogoutCookie(cookie *models.ClientCookie)

	GetCookieCurrencies(mutex *config.AllMutexes) ([]models.CodeBlock, []models.CodeBlock)
	SetCookieCurrency(c *models.ClientCookie, mutex *config.AllMutexes, choice string) error

	GetContactsWithDefault(customerID int) ([]*models.Contact, error)
	Update(cust *models.Customer) error
	AddContactToCustomer(contact *models.Contact) error
}

type customerService struct {
	customerRepo repositories.CustomerRepository
}

func NewCustomerService(customerRepo repositories.CustomerRepository) CustomerService {
	return &customerService{customerRepo: customerRepo}
}

func (s *customerService) GetCustomerByID(id int) (*models.Customer, error) {
	return s.customerRepo.Read(id)
}

func (s *customerService) GetCustomerAndContacts(dpi *DataPassIn) (*models.Customer, []*models.Contact, error) {
	cust, cont, err := s.customerRepo.GetCustomerAndContacts(dpi.CustomerID)
	if err != nil {
		return nil, nil, err
	}

	if cust.Status != "Active" {
		return nil, nil, errors.New("deleted customer")
	}

	return cust, cont, err
}

func (s *customerService) GetPaymentMethodsCust(dpi *DataPassIn) ([]models.PaymentMethodStripe, error) {
	return s.customerRepo.GetPaymentMethodsCust(dpi.CustomerID)
}

func (s *customerService) AddAddressToCustomer(dpi *DataPassIn, contact *models.Contact, mutex *config.AllMutexes) error {
	contact.CustomerID = dpi.CustomerID

	if err := custhelp.VerifyContact(contact, mutex); err != nil {
		return err
	}

	return s.customerRepo.AddContactToCustomer(contact)
}

// Actual contacts, contact update error, contact retrieval error
func (s *customerService) AddAddressAndRender(dpi *DataPassIn, contact *models.Contact, mutex *config.AllMutexes, isDefault bool) ([]*models.Contact, error, error) {
	updateErr := s.AddAddressToCustomer(dpi, contact, mutex)

	if updateErr == nil && isDefault {
		updateErr = s.customerRepo.UpdateCustomerDefault(dpi.CustomerID, contact.ID)
	}

	list, getErr := s.customerRepo.GetContactsWithDefault(dpi.CustomerID)

	return list, updateErr, getErr
}

func (s *customerService) MakeAddressDefault(dpi *DataPassIn, contactID int) error {
	c, err := s.customerRepo.GetSingleContact(contactID)
	if err != nil {
		return err
	} else if c == nil {
		return fmt.Errorf("unable to find non-nil contact for id: %d", contactID)
	} else if c.CustomerID != dpi.CustomerID {
		return fmt.Errorf("customer id %d doesn't match contact customer id of %d for contact id: %d", dpi.CustomerID, c.CustomerID, contactID)
	}

	return s.customerRepo.UpdateCustomerDefault(dpi.CustomerID, contactID)
}

// Actual contacts, contact update error, contact retrieval error
func (s *customerService) MakeAddressDefaultAndRender(dpi *DataPassIn, contactID int) ([]*models.Contact, error, error) {
	updateErr := s.MakeAddressDefault(dpi, contactID)

	list, getErr := s.customerRepo.GetContactsWithDefault(dpi.CustomerID)

	return list, updateErr, getErr
}

func (s *customerService) UpdateContact(dpi *DataPassIn, contactID int, newContact *models.Contact, mutex *config.AllMutexes) error {
	newContact.CustomerID = dpi.CustomerID

	if err := custhelp.VerifyContact(newContact, mutex); err != nil {
		return err
	}

	oldContact, err := s.customerRepo.GetSingleContact(contactID)
	if err != nil {
		return err
	} else if oldContact.CustomerID != dpi.CustomerID {
		return fmt.Errorf("non matching customer id for update contacts, id provided: %d; id on contact: %d", dpi.CustomerID, oldContact.CustomerID)
	}

	newContact.ID = contactID

	return s.customerRepo.UpdateContact(*newContact)

}

// Actual contacts, contact update error, contact retrieval error
func (s *customerService) UpdateContactAndRender(dpi *DataPassIn, contactID int, newContact *models.Contact, mutex *config.AllMutexes, isDefault bool) ([]*models.Contact, error, error) {
	updateErr := s.UpdateContact(dpi, contactID, newContact, mutex)

	if updateErr == nil && isDefault {
		updateErr = s.customerRepo.UpdateCustomerDefault(dpi.CustomerID, contactID)
	}

	list, getErr := s.customerRepo.GetContactsWithDefault(dpi.CustomerID)

	return list, updateErr, getErr
}

func (s *customerService) DeleteContact(dpi *DataPassIn, contactID int) (int, error) {
	newDefaultID := 0

	list, err := s.customerRepo.GetContactsWithDefault(dpi.CustomerID)
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
		return newDefaultID, fmt.Errorf("no contact of id: %d for customer: %d", contactID, dpi.CustomerID)
	}

	if index == 0 && len(list) > 1 {
		newDefaultID = list[1].ID
	}

	return newDefaultID, s.customerRepo.DeleteContact(contactID)
}

// Actual contacts, contact update error, contact retrieval error
func (s *customerService) DeleteContactAndRender(dpi *DataPassIn, contactID int) ([]*models.Contact, error, error) {
	newDefault, updateErr := s.DeleteContact(dpi, contactID)

	if updateErr == nil && newDefault > 0 {
		updateErr = s.customerRepo.UpdateCustomerDefault(dpi.CustomerID, newDefault)
	}

	list, getErr := s.customerRepo.GetContactsWithDefault(dpi.CustomerID)

	return list, updateErr, getErr
}

func (s *customerService) CreateCustomer(dpi *DataPassIn, customer *models.CustomerPost, firebaseID string) (*models.Customer, *models.ServerCookie, error) {
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

	c, err := s.customerRepo.CreateServerCookie(newCust.ID, firebaseID, dpi.Store)

	return newCust, c, err
}

func (s *customerService) DeleteCustomer(dpi *DataPassIn) (*models.Customer, error) {
	cust, err := s.customerRepo.Read(dpi.CustomerID)
	if err != nil {
		return nil, err
	}

	if cust.Status == "Archived" {
		return nil, fmt.Errorf("already archived customer for id: %d", dpi.CustomerID)
	}

	cust.Status = "Archived"
	if err := s.customerRepo.Update(*cust); err != nil {
		return nil, err
	}

	cookie, err := s.customerRepo.GetServerCookie(dpi.CustomerID, dpi.Store)
	if err != nil {
		// log it
		return nil, err
	}

	_, err = s.customerRepo.SetServerCookieStatus(cookie, true)

	return cust, err
}

func (s *customerService) UpdateCustomer(dpi *DataPassIn, customer *models.CustomerPost) (*models.Customer, error) {
	validate := validator.New()
	err := validate.Struct(customer)
	if err != nil {
		return nil, err
	}

	cust, err := s.customerRepo.Read(dpi.CustomerID)
	if err != nil {
		return nil, err
	}

	if cust.Status == "Archived" {
		return nil, fmt.Errorf("inactive customer for id: %d", dpi.CustomerID)
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

func (s *customerService) LoginCookie(dpi *DataPassIn, firebaseID string) (*models.ClientCookie, error) {
	customer, err := s.customerRepo.GetCustomerByFirebase(firebaseID)
	if err != nil {
		return nil, err
	} else if customer == nil {
		return nil, fmt.Errorf("no active customer for firebase ID: %s", firebaseID)
	} else if customer.Status == "Archived" {
		return nil, fmt.Errorf("archived customer for firebase ID: %s", firebaseID)
	}

	serverCookie, err := s.customerRepo.GetServerCookie(customer.ID, dpi.Store)
	if err != nil {
		return nil, err
	} else if serverCookie == nil {
		return nil, fmt.Errorf("no active server cookie for firebase ID: %s; customer id: %d; store: %s", firebaseID, customer.ID, dpi.Store)
	} else if customer.Status == "Archived" {
		return nil, fmt.Errorf("archived server cookie for firebase ID: %s; customer id: %d; store: %s", firebaseID, customer.ID, dpi.Store)
	}

	return &models.ClientCookie{
		Store:       dpi.Store,
		CustomerID:  customer.ID,
		CustomerSet: time.Now(),
		GuestID:     dpi.GuestID,
	}, nil
}

func (s *customerService) ResetPass(dpi *DataPassIn, firebaseID string) error {
	customer, err := s.customerRepo.GetCustomerByFirebase(firebaseID)
	if err != nil {
		return err
	} else if customer == nil {
		return fmt.Errorf("no active customer for firebase ID: %s", firebaseID)
	} else if customer.Status == "Archived" {
		return fmt.Errorf("archived customer for firebase ID: %s", firebaseID)
	}

	serverCookie, err := s.customerRepo.GetServerCookie(customer.ID, dpi.Store)
	if err != nil {
		return err
	} else if serverCookie == nil {
		return fmt.Errorf("no active server cookie for firebase ID: %s; customer id: %d; store: %s", firebaseID, customer.ID, dpi.Store)
	} else if customer.Status == "Archived" {
		return fmt.Errorf("archived server cookie for firebase ID: %s; customer id: %d; store: %s", firebaseID, customer.ID, dpi.Store)
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

func (s *customerService) CustomerMiddleware(cookie *models.ClientCookie) {

	if cookie.CustomerID > 0 {
		serverCookie, err := s.customerRepo.GetServerCookie(cookie.CustomerID, cookie.Store)
		if err != nil {
			customer, err := s.customerRepo.Read(cookie.CustomerID)
			if err != nil {
				log.Printf("Unable to query customer for customer id: %d; store: %s; error: %v\n", cookie.CustomerID, cookie.Store, err)
				cookie.CustomerID = 0
				cookie.CustomerSet = time.Time{}
				return
			} else if customer == nil {
				log.Printf("No active customer for customer id: %d; store: %s\n", cookie.CustomerID, cookie.Store)
				cookie.CustomerID = 0
				cookie.CustomerSet = time.Time{}
				return
			}
			if customer.Status == "Archived" || customer.LastReset.After(cookie.CustomerSet) {
				cookie.CustomerID = 0
				cookie.CustomerSet = time.Time{}
				return
			}
		} else if serverCookie == nil {
			log.Printf("No active server cookie for customer id: %d; store: %s\n", cookie.CustomerID, cookie.Store)
			cookie.CustomerID = 0
			cookie.CustomerSet = time.Time{}
			return
		} else if serverCookie.Archived || serverCookie.LastReset.After(cookie.CustomerSet) {
			cookie.CustomerID = 0
			cookie.CustomerSet = time.Time{}
		}
	}
}

func (s *customerService) GuestMiddleware(cookie *models.ClientCookie, store string) {
	cookie.Store = store

	if cookie.GuestID == "" {
		cookie.GuestID = fmt.Sprintf("GI:%s", uuid.New().String())
	}
}

func (s *customerService) FullMiddleware(cookie *models.ClientCookie, store string) {
	if cookie == nil {
		cookie = &models.ClientCookie{}
	}
	s.GuestMiddleware(cookie, store)
	s.CustomerMiddleware(cookie)
}

func (s *customerService) LogoutCookie(cookie *models.ClientCookie) {
	cookie.CustomerID = 0
	cookie.CustomerSet = time.Time{}
}

func (s *customerService) GetCookieCurrencies(mutex *config.AllMutexes) ([]models.CodeBlock, []models.CodeBlock) {
	mainCurrencies, otherCurrencies := []models.CodeBlock{}, []models.CodeBlock{}

	mutex.Currency.Mu.RLock()

	for i, b := range mutex.Currency.List.List {
		if i < 5 {
			mainCurrencies = append(mainCurrencies, b)
		} else {
			otherCurrencies = append(otherCurrencies, b)
		}
	}

	mutex.Currency.Mu.RUnlock()

	return mainCurrencies, otherCurrencies
}

func (s *customerService) SetCookieCurrency(c *models.ClientCookie, mutex *config.AllMutexes, choice string) error {

	if c == nil {
		return errors.New("nil cookie pointer")
	}

	mainCurrencies, otherCurrencies := s.GetCookieCurrencies(mutex)

	in := false
	for _, b := range mainCurrencies {
		if b.Code == choice {
			in = true
			break
		}
	}
	if !in {
		for _, b := range otherCurrencies {
			if b.Code == choice {
				in = true
				break
			}
		}
	}

	if !in {
		return errors.New("not approved currency choice: " + choice)
	}

	c.Currency = choice
	c.OtherCurrency = choice == "USD"

	return nil
}

func (s *customerService) GetContactsWithDefault(customerID int) ([]*models.Contact, error) {
	return s.customerRepo.GetContactsWithDefault(customerID)
}

func (s *customerService) Update(cust *models.Customer) error {
	return s.customerRepo.Update(*cust)
}

func (s *customerService) AddContactToCustomer(contact *models.Contact) error {
	return s.customerRepo.AddContactToCustomer(contact)
}
