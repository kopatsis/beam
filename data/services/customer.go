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
	"math/rand"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
	CreateCustomer(dpi *DataPassIn, customer *models.CustomerPost, tools *config.Tools) (*models.ClientCookie, *models.TwoFactorCookie, *models.Customer, *models.ServerCookie, error)
	DeleteCustomer(dpi *DataPassIn) (*models.Customer, error)
	UpdateCustomer(dpi *DataPassIn, customer *models.CustomerPost) (*models.Customer, error)

	LoginCookie(dpi *DataPassIn, email string, addEmailSub bool) (*models.ClientCookie, *models.TwoFactorCookie, error)
	ResetPass(dpi *DataPassIn, email string) error
	CustomerMiddleware(cookie *models.ClientCookie, device *models.DeviceCookie)
	GuestMiddleware(cookie *models.ClientCookie, store string)
	FullMiddleware(cookie *models.ClientCookie, device *models.DeviceCookie, store string)
	TwoFAMiddleware(cookie *models.ClientCookie, twofa *models.TwoFactorCookie)
	LogoutCookie(dpi *DataPassIn, cookie *models.ClientCookie) error

	GetCookieCurrencies(mutex *config.AllMutexes) ([]models.CodeBlock, []models.CodeBlock)
	SetCookieCurrency(c *models.ClientCookie, mutex *config.AllMutexes, choice string) error

	GetContactsWithDefault(customerID int) ([]*models.Contact, error)
	Update(cust *models.Customer) error
	AddContactToCustomer(contact *models.Contact) error

	ToggleEmailVerified(dpi *DataPassIn, verified bool) error
	ToggleEmailSubbed(dpi *DataPassIn, subbed bool) error
	ToggleTwoFactor(dpi *DataPassIn, uses bool) error

	WatchEmailVerification(dpi *DataPassIn, conn *websocket.Conn)
	SendVerificationEmail(dpi *DataPassIn, tools *config.Tools) (string, error)
	ProcessVerificationEmail(dpi *DataPassIn, param string) error
	SendSignInEmail(dpi *DataPassIn, email string, emailSubbed bool, tools *config.Tools) (string, error)
	ProcessSignInEmail(dpi *DataPassIn, param string, tools *config.Tools) (*models.ClientCookie, error)
	CreateTwoFACode(cust *models.Customer, store string) (*models.TwoFactorCookie, error)
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

func (s *customerService) CreateCustomer(dpi *DataPassIn, customer *models.CustomerPost, tools *config.Tools) (*models.ClientCookie, *models.TwoFactorCookie, *models.Customer, *models.ServerCookie, error) {
	validate := validator.New()
	err := validate.Struct(customer)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if customer.IsPassword && customer.Password != customer.PasswordConf {
		return nil, nil, nil, nil, errors.New("passwords don't match")
	}

	email := strings.ToLower(customer.Email)
	if !custhelp.VerifyEmail(email, tools) {
		return nil, nil, nil, nil, errors.New("invalid email")
	}

	id, oldArch, err := s.customerRepo.GetCustomerIDByEmail(email)
	if err != nil {
		return nil, nil, nil, nil, err
	} else if id != 0 {
		return nil, nil, nil, nil, errors.New("existing customer that wasn't an old archival")
	}

	if oldArch {
		if err := s.customerRepo.ArchiveCustomerEmail(id, customer.Email); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	newCust := &models.Customer{
		FirstName:     customer.FirstName,
		LastName:      customer.LastName,
		Email:         email,
		EmailSubbed:   customer.EmailSubbed,
		Status:        "Active",
		Created:       time.Now(),
		EmailVerified: customer.IsEmailVerified,
	}

	if customer.PhoneNumber != nil {
		newCust.PhoneNumber = orderhelp.CopyString(customer.PhoneNumber)
	}

	if customer.IsPassword {
		password, err := custhelp.EncryptPassword(customer.Password)
		if err != nil {
			// send an error to me as this is major
			return nil, nil, nil, nil, errors.New("unable to encrypt password: " + err.Error())
		}
		newCust.PasswordHash = password
	}

	if err := s.customerRepo.Create(*newCust); err != nil {
		return nil, nil, nil, nil, err
	}

	c, err := s.customerRepo.CreateServerCookie(newCust.ID, dpi.Store, customer.IsEmailVerified, false)
	if err != nil {
		return nil, nil, newCust, nil, err
	}

	if err := s.customerRepo.SetDeviceMapping(newCust.ID, dpi.DeviceID, dpi.Store); err != nil {
		return nil, nil, newCust, c, err
	}

	var twofa *models.TwoFactorCookie
	if customer.Uses2FA {
		twofa, err = s.CreateTwoFACode(newCust, dpi.Store)
		if err != nil {
			return nil, nil, newCust, c, err
		}
	}

	return &models.ClientCookie{
		Store:         dpi.Store,
		CustomerID:    newCust.ID,
		CustomerSet:   time.Now(),
		GuestID:       dpi.GuestID,
		OtherCurrency: false,
		Currency:      "USD",
	}, twofa, newCust, c, nil
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
	cust.Archived = time.Now()
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
	cust.FirstName = customer.FirstName
	cust.LastName = customer.LastName
	cust.PhoneNumber = customer.PhoneNumber

	if err := s.customerRepo.Update(*cust); err != nil {
		return nil, err
	}

	return cust, nil
}

func (s *customerService) LoginCookie(dpi *DataPassIn, email string, addEmailSub bool) (*models.ClientCookie, *models.TwoFactorCookie, error) {
	customer, err := s.customerRepo.GetActiveCustomerByEmail(email)
	if err != nil {
		return nil, nil, err
	} else if customer == nil {
		return nil, nil, fmt.Errorf("no active customer for email:  %s", email)
	} else if customer.Status == "Archived" {
		return nil, nil, fmt.Errorf("archived customer for email:  %s", email)
	}

	serverCookie, err := s.customerRepo.GetServerCookie(customer.ID, dpi.Store)
	if err != nil {
		return nil, nil, err
	} else if serverCookie == nil {
		return nil, nil, fmt.Errorf("no active server cookie for email:  %s; customer id: %d; store: %s", email, customer.ID, dpi.Store)
	} else if customer.Status == "Archived" {
		return nil, nil, fmt.Errorf("archived server cookie for email:  %s; customer id: %d; store: %s", email, customer.ID, dpi.Store)
	}

	if addEmailSub {
		if err := s.customerRepo.SetEmailSubbed(customer.ID, true); err != nil {
			log.Printf("failure to set user's email sub to true")
		}
	}

	if err := s.customerRepo.SetDeviceMapping(customer.ID, dpi.DeviceID, dpi.Store); err != nil {
		return nil, nil, err
	}

	var twofa *models.TwoFactorCookie
	if customer.Uses2FA {
		twofa, err = s.CreateTwoFACode(customer, dpi.Store)
		if err != nil {
			return nil, nil, err
		}
	}

	return &models.ClientCookie{
		Store:         dpi.Store,
		CustomerID:    customer.ID,
		CustomerSet:   time.Now(),
		GuestID:       dpi.GuestID,
		OtherCurrency: customer.UsesOtherCurrency,
		Currency:      customer.OtherCurrency,
	}, twofa, nil
}

func (s *customerService) ResetPass(dpi *DataPassIn, email string) error {
	customer, err := s.customerRepo.GetActiveCustomerByEmail(email)
	if err != nil {
		return err
	} else if customer == nil {
		return fmt.Errorf("no active customer for email:  %s", email)
	} else if customer.Status == "Archived" {
		return fmt.Errorf("archived customer for email:  %s", email)
	}

	serverCookie, err := s.customerRepo.GetServerCookie(customer.ID, dpi.Store)
	if err != nil {
		return err
	} else if serverCookie == nil {
		return fmt.Errorf("no active server cookie for email:  %s; customer id: %d; store: %s", email, customer.ID, dpi.Store)
	} else if customer.Status == "Archived" {
		return fmt.Errorf("archived server cookie for email:  %s; customer id: %d; store: %s", email, customer.ID, dpi.Store)
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

func (s *customerService) CustomerMiddleware(cookie *models.ClientCookie, device *models.DeviceCookie) {

	if cookie == nil {
		return
	}

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
			}
			// Notify me somehow server cookie not set
			if _, err := s.customerRepo.CreateServerCookie(customer.ID, cookie.Store, customer.EmailVerified, customer.Status == "Archived"); err != nil {
				// Notify me server cookie can't be set
				log.Printf("Unable to set server cookie for valid customer\n")
			}
		} else if serverCookie == nil {
			log.Printf("No active server cookie for customer id: %d; store: %s\n", cookie.CustomerID, cookie.Store)
			cookie.CustomerID = 0
			cookie.CustomerSet = time.Time{}
			return
		} else if serverCookie.Archived || serverCookie.LastForcedLogout.After(cookie.CustomerSet) {
			cookie.CustomerID = 0
			cookie.CustomerSet = time.Time{}
		}

		if cookie.CustomerID > 0 {
			id, err := s.customerRepo.GetDeviceMapping(device.DeviceID, cookie.Store)
			if err != nil || cookie.CustomerID != id {
				cookie.CustomerID = 0
				cookie.CustomerSet = time.Time{}
			}
		}

	}
}

func (s *customerService) GuestMiddleware(cookie *models.ClientCookie, store string) {
	cookie.Store = store

	if cookie.GuestID == "" {
		cookie.GuestID = fmt.Sprintf("GI:%s", uuid.New().String())
	}
}

func (s *customerService) FullMiddleware(cookie *models.ClientCookie, device *models.DeviceCookie, store string) {
	if cookie == nil {
		return
	}
	s.GuestMiddleware(cookie, store)
	s.CustomerMiddleware(cookie, device)
}

func (s *customerService) TwoFAMiddleware(cookie *models.ClientCookie, twofa *models.TwoFactorCookie) {
	if cookie == nil || twofa == nil || cookie.CustomerID == 0 || twofa.TwoFactorCode == "" {
		twofa.TwoFactorCode = ""
		twofa.CustomerID = 0
		twofa.Set = time.Time{}
		return
	}
	if time.Since(twofa.Set) > config.TWOFA_EXPIR_MINS*time.Minute || twofa.CustomerID != cookie.CustomerID {
		twofa.TwoFactorCode = ""
		twofa.CustomerID = 0
		twofa.Set = time.Time{}
	}
}

func (s *customerService) LogoutCookie(dpi *DataPassIn, cookie *models.ClientCookie) error {
	cookie.CustomerID = 0
	cookie.CustomerSet = time.Time{}

	return s.customerRepo.SetDeviceMapping(0, dpi.DeviceID, dpi.Store)
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

	go func() {
		if err := s.customerRepo.UpdateCustomerCurrency(c.CustomerID, c.OtherCurrency, c.Currency); err != nil {
			// Email me this error, low priority
		}
	}()

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

func (s *customerService) ToggleEmailVerified(dpi *DataPassIn, verified bool) error {
	if err := s.customerRepo.SetEmailVerified(dpi.CustomerID, verified); err != nil {
		return err
	}

	c, err := s.customerRepo.GetServerCookie(dpi.CustomerID, dpi.Store)
	if err != nil {
		// Notify me since server cookie needs to have correct information
		return err
	}

	if _, err := s.customerRepo.SetServerCookieCompletion(c, verified); err != nil {
		// Notify me since server cookie needs to have correct information
		return err
	}

	return nil
}

func (s *customerService) ToggleEmailSubbed(dpi *DataPassIn, subbed bool) error {
	return s.customerRepo.SetEmailSubbed(dpi.CustomerID, subbed)
}

func (s *customerService) ToggleTwoFactor(dpi *DataPassIn, uses bool) error {
	cust, err := s.customerRepo.Read(dpi.CustomerID)
	if err != nil {
		return err
	}

	if cust.Uses2FA == uses {
		return nil
	}

	cust.Uses2FA = uses
	if err := s.customerRepo.Update(*cust); err != nil {
		return err
	}

	if uses {
		c, err := s.customerRepo.GetServerCookie(cust.ID, dpi.Store)
		if err != nil {
			// Notify me but not return error
			log.Printf("Unable to set forced logout date\n")
			return nil
		} else if c == nil {
			// Notify me but not return error
			log.Printf("Unable to set forced logout date\n")
			return nil
		}

		if _, err := s.customerRepo.SetServerCookieReset(c, time.Now()); err != nil {
			// Notify me but not return error
			log.Printf("Unable to set forced logout date\n")
		}
	}

	return nil
}

func (s *customerService) WatchEmailVerification(dpi *DataPassIn, conn *websocket.Conn) {

	if dpi.CustomerID <= 0 {
		conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
		conn.Close()
		return
	}

	intervals := []time.Duration{2500 * time.Millisecond, 5000 * time.Millisecond, 10000 * time.Millisecond, 25000 * time.Millisecond}
	limits := []time.Duration{15 * time.Second, 45 * time.Second, 105 * time.Second, 240 * time.Second}

	start := time.Now()
	for i, interval := range intervals {
		deadline := start.Add(limits[i])
		for time.Now().Before(deadline) {
			nextCheck := time.Now().Add(interval)
			verified, err := s.customerRepo.IsEmailVerified(dpi.CustomerID)
			if err != nil || verified {
				if err == nil {
					conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
				} else {
					conn.WriteMessage(websocket.TextMessage, []byte("cease"))
				}
				conn.Close()
				return
			}
			time.Sleep(time.Until(nextCheck))
		}
	}
	conn.Close()
}

func (s *customerService) SendVerificationEmail(dpi *DataPassIn, tools *config.Tools) (string, error) {
	cust, err := s.customerRepo.Read(dpi.CustomerID)
	if err != nil {
		return "", err
	} else if cust.Status == "Archived" {
		return "", errors.New("archived customer")
	}

	if cust.EmailVerified {
		return "", nil
	}

	if !custhelp.VerifyEmail(cust.Email, tools) {
		// Notify me an existing customer's email not deliverable
		return "", errors.New("email found to be undeliverable")
	}

	id := "EV-" + uuid.NewString()
	storedParam := models.VerificationEmailParam{Param: id, EmailAtTime: cust.Email, CustomerID: cust.ID, Set: time.Now()}

	return id, s.customerRepo.StoreVerificationEmail(storedParam, dpi.Store)
}

func (s *customerService) ProcessVerificationEmail(dpi *DataPassIn, param string) error {
	verifParams, err := s.customerRepo.GetVerificationEmail(param, dpi.Store)
	if err != nil {
		return err
	}

	if verifParams.Param != param {
		return errors.New("params do not match")
	}

	if time.Since(verifParams.Set) > time.Duration(config.VERIF_EXPIR_MINS)*time.Minute {
		// Alert me about it, since it should have expired on redis too
		return errors.New("expired code, not deleted in system")
	}

	cust, err := s.customerRepo.Read(verifParams.CustomerID)
	if err != nil {
		return err
	}

	if cust == nil || cust.ID != verifParams.CustomerID {
		return errors.New("unable to retrieve correct customer")
	}

	if cust.Status == "Archived" {
		return errors.New("archived customer")
	}

	if cust.Email != verifParams.EmailAtTime {
		return errors.New("email was changed, cannot verify")
	}

	if cust.EmailVerified {
		return nil
	}

	return s.customerRepo.SetEmailVerified(cust.ID, true)
}

func (s *customerService) SendSignInEmail(dpi *DataPassIn, email string, emailSubbed bool, tools *config.Tools) (string, error) {
	cust, err := s.customerRepo.GetCustomerByEmail(email)
	if err != nil {
		return "", err
	} else if cust != nil && cust.Status == "Archived" && time.Since(cust.Archived) <= 7*24*time.Hour {
		return "", errors.New("recently archived customer")
	}

	if !custhelp.VerifyEmail(cust.Email, tools) {
		// Notify me an existing customer's email not deliverable
		return "", errors.New("email found to be undeliverable")
	}

	id := "SI-" + uuid.NewString()
	storedParam := models.SignInEmailParam{Param: id, EmailAtTime: email, DeviceCookie: dpi.DeviceID, Set: time.Now(), EmailSubbed: emailSubbed}

	return id, s.customerRepo.StoreSignInEmail(storedParam, dpi.Store)
}

func (s *customerService) ProcessSignInEmail(dpi *DataPassIn, param string, tools *config.Tools) (*models.ClientCookie, error) {
	signinParams, err := s.customerRepo.GetSignInEmail(param, dpi.Store)
	if err != nil {
		return nil, err
	}

	if signinParams.Param != param {
		return nil, errors.New("params do not match")
	}

	if dpi.DeviceID != signinParams.DeviceCookie {
		return nil, errors.New("must open link on same device and browser")
	}

	if time.Since(signinParams.Set) > time.Duration(config.SIGNIN_EXPIR_MINS)*time.Minute {
		// Alert me about it, since it should have expired on redis too
		return nil, errors.New("expired code, not deleted in system")
	}

	cust, err := s.customerRepo.GetCustomerByEmail(signinParams.EmailAtTime)
	if err != nil {
		return nil, err
	}

	if cust == nil || cust.Status == "Archived" && time.Since(cust.Archived) > 7*24*time.Hour {
		custPost := models.CustomerPost{
			Email:           signinParams.EmailAtTime,
			EmailSubbed:     signinParams.EmailSubbed,
			IsEmailVerified: true,
		}
		client, _, _, _, err := s.CreateCustomer(dpi, &custPost, tools)
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	client, _, err := s.LoginCookie(dpi, signinParams.EmailAtTime, signinParams.EmailSubbed)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (s *customerService) CreateTwoFACode(cust *models.Customer, store string) (*models.TwoFactorCookie, error) {

	code := "TF-" + uuid.NewString()
	sixdigit := uint(100000 + rand.Intn(900000))
	setTime := time.Now()
	param := models.TwoFactorEmailParam{Param: code, CustomerID: cust.ID, Set: setTime, SixDigitCode: sixdigit, Tries: 0}
	cookie := models.TwoFactorCookie{TwoFactorCode: code, CustomerID: cust.ID, Set: setTime}

	return &cookie, s.customerRepo.StoreTwoFA(param, store)
}
