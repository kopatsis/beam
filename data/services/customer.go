package services

import (
	"beam/background/emails"
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
	CreateCustomer(dpi *DataPassIn, customer *models.CustomerPost, tools *config.Tools) (*models.ClientCookie, *models.TwoFactorCookie, *models.Customer, *models.ServerCookie, error) // To cart/draft/order IF !2FA
	DeleteCustomer(dpi *DataPassIn) (*models.Customer, error)
	UpdateCustomer(dpi *DataPassIn, customer *models.CustomerPost) (*models.Customer, error)

	LoginCookie(dpi *DataPassIn, email, password string, addEmailSub, usesPassword bool, tools *config.Tools) (*models.ClientCookie, *models.TwoFactorCookie, error) // To cart/draft/order IF !2FA
	ResetPass(dpi *DataPassIn, email string) error
	CustomerMiddleware(cookie *models.ClientCookie, device *models.DeviceCookie)
	GuestMiddleware(cookie *models.ClientCookie, store string)
	FullMiddleware(cookie *models.ClientCookie, device *models.DeviceCookie, store string)
	TwoFAMiddleware(cookie *models.ClientCookie, twofa *models.TwoFactorCookie)
	SignInCodeMiddleware(cookie *models.ClientCookie, si *models.SignInCodeCookie)

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
	ProcessVerificationEmail(dpi *DataPassIn, param string) error // To cart/draft/order ?

	SendSignInCodeEmail(dpi *DataPassIn, email string, tools *config.Tools) (*models.SignInCodeCookie, bool, error)
	ProcessSignInCodeEmail(dpi *DataPassIn, siCookie *models.SignInCodeCookie, sixdigits uint, post *models.CustomerPost, tools *config.Tools) (*models.ClientCookie, error) // To cart/draft/order
	ResendSignInCode(dpi *DataPassIn, siCookie models.SignInCodeCookie, tools *config.Tools) (models.SignInCodeCookie, error)

	CreateTwoFACode(cust *models.Customer, store string, tools *config.Tools) (*models.TwoFactorCookie, error)
	ProcessTwoFactor(dpi *DataPassIn, twofactorcookie *models.TwoFactorCookie, sixdigits uint) (bool, error)
	ResendTwoFactor(dpi *DataPassIn, twofactorcookie models.TwoFactorCookie, tools *config.Tools) (models.TwoFactorCookie, error)

	ChangeCustomerEmail(dpi *DataPassIn, newEmail, password string, tools *config.Tools) error

	ActualEmailVerification(store, param string, customer *models.Customer, tools *config.Tools) error
	DelayedEmailVerification(store, param string, customerID int, wait time.Duration, tools *config.Tools) error
	SendVerificationToEmail(store, param string, customer *models.Customer, tools *config.Tools) error

	UnsubLinkForEmails(store string, customerID int) (string, string, string)
	UnsubCustomerDirect(store, storeEncr, customerEncr, timestamp string) error

	SendResetEmail(dpi *DataPassIn, email string, tools *config.Tools) (string, error)
	ProcessResetEmail(dpi *DataPassIn, param string) (*models.ResetEmailCookie, error)
	ResetPasswordActual(dpi *DataPassIn, resetCookie *models.ResetEmailCookie, password, passwordConfirm string, logAllOut bool) error

	BirthdayEmails(store string, ds DiscountService, tools *config.Tools) error
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

	if customer.IsPassword {
		if !custhelp.PasswordMeetsRequirements(customer.Password, customer.PasswordConf, true) {
			return nil, nil, nil, nil, errors.New("password doesn't meet criteria")
		}
	}

	email := strings.ToLower(customer.Email)
	if !custhelp.VerifyEmail(email, tools) {
		return nil, nil, nil, nil, errors.New("invalid email")
	}

	cust, err := s.customerRepo.GetCustomerByEmail(email)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if cust != nil {
		if cust.Status == "Active" {
			return nil, nil, nil, nil, errors.New("cannot create account for active customer")
		} else if cust.Status == "Archived" && time.Since(cust.Archived) <= 7*24*time.Hour {
			return nil, nil, nil, nil, errors.New("archived and still in cooldown period")
		}
		if cust.Status == "Archived" {
			if err := s.customerRepo.ArchiveCustomerEmail(cust.ID, customer.Email); err != nil {
				return nil, nil, nil, nil, err
			} else {
				cust = nil
			}
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
		BirthdaySet:   customer.HasBirthday,
		BirthDay:      customer.BirthDay,
		BirthMonth:    customer.BirthMonth,
	}

	if cust != nil {
		newCust.ID = cust.ID
	}

	if !customer.IsEmailVerified {
		newCust.Status = "Draft"
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
		twofa, err = s.CreateTwoFACode(newCust, dpi.Store, tools)
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

	cust.EmailSubbed = customer.EmailSubbed
	cust.FirstName = customer.FirstName
	cust.LastName = customer.LastName
	cust.PhoneNumber = customer.PhoneNumber
	cust.BirthdaySet = customer.HasBirthday
	cust.BirthMonth = customer.BirthMonth
	cust.BirthDay = customer.BirthDay

	if err := s.customerRepo.Update(*cust); err != nil {
		return nil, err
	}

	return cust, nil
}

func (s *customerService) LoginCookie(dpi *DataPassIn, email, password string, addEmailSub, usesPassword bool, tools *config.Tools) (*models.ClientCookie, *models.TwoFactorCookie, error) {

	email = strings.ToLower(email)

	customer, err := s.customerRepo.GetCustomerByEmail(email)
	if err != nil {
		return nil, nil, err
	} else if customer == nil {
		return nil, nil, fmt.Errorf("no active customer for email:  %s", email)
	} else if customer.Status == "Archived" {
		return nil, nil, fmt.Errorf("archived customer for email:  %s", email)
	}

	if usesPassword && (!custhelp.PasswordMeetsRequirements(password, "", false) || !custhelp.CheckPassword(customer.PasswordHash, password)) {
		return nil, nil, errors.New("password is incorrect or invalid")
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
		twofa, err = s.CreateTwoFACode(customer, dpi.Store, tools)
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
	email = strings.ToLower(email)

	customer, err := s.customerRepo.GetCustomerByEmail(email)
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
			if customer.Status == "Archived" {
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
	if cookie == nil || twofa == nil {
		return
	}
	if cookie.CustomerID == 0 || twofa.TwoFactorCode == "" {
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

func (s *customerService) SignInCodeMiddleware(cookie *models.ClientCookie, si *models.SignInCodeCookie) {
	if cookie == nil || si == nil {
		return
	}
	if cookie.CustomerID != 0 || si.Param == "" || time.Since(si.Set) > config.SIGNIN_EXPIR_MINS*time.Minute {
		si.Param = ""
		si.CustomerID = 0
		si.Set = time.Time{}
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

	if err := s.customerRepo.StoreVerificationEmail(storedParam, dpi.Store); err != nil {
		return "", err
	}

	return id, s.ActualEmailVerification(dpi.Store, id, cust, tools)
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

// Cookie, whether to render full form, error
func (s *customerService) SendSignInCodeEmail(dpi *DataPassIn, email string, tools *config.Tools) (*models.SignInCodeCookie, bool, error) {

	email = strings.ToLower(email)

	cust, err := s.customerRepo.GetCustomerByEmail(email)
	if err != nil {
		return nil, false, err
	} else if cust != nil && cust.Status == "Archived" && time.Since(cust.Archived) <= 7*24*time.Hour {
		return nil, false, errors.New("recently archived customer")
	} else if cust != nil && cust.Status == "Active" && cust.Uses2FA {
		return nil, false, errors.New("uses 2FA customers not allowed to sign in via pin")
	}

	isCustomer := cust != nil && cust.Status == "Active"
	custID := 0
	if isCustomer {
		custID = cust.ID
	}

	if !custhelp.VerifyEmail(email, tools) {
		if cust != nil {
			// Notify me an existing customer's email not deliverable
		}
		return nil, false, errors.New("email found to be undeliverable")
	}

	id := "SI-" + uuid.NewString()
	sixdigit := uint(100000 + rand.Intn(900000))
	setTime := time.Now()

	storedParam := models.SignInEmailParam{Param: id, EmailAtTime: email, Set: setTime, SixDigitCode: sixdigit, HasCustomer: isCustomer, CustomerID: custID}
	siCookie := models.SignInCodeCookie{Param: id, IsCustomer: isCustomer, CustomerID: custID}

	if err := emails.SignInPin(dpi.Store, storedParam.EmailAtTime, storedParam.SixDigitCode, tools); err != nil {
		return nil, false, err
	}

	return &siCookie, !isCustomer, s.customerRepo.StoreSignInEmail(storedParam, dpi.Store)
}

// nil, nil means everything else right, but 6 digit code wrong
func (s *customerService) ProcessSignInCodeEmail(dpi *DataPassIn, siCookie *models.SignInCodeCookie, sixdigits uint, post *models.CustomerPost, tools *config.Tools) (*models.ClientCookie, error) {
	if time.Since(siCookie.Set) > config.SIGNIN_EXPIR_MINS*time.Minute {
		return nil, errors.New("past expiration")
	}

	if err := s.customerRepo.SetSignInCodeNX(siCookie.Param, dpi.Store); err != nil {
		return nil, err
	}

	removeSI := false
	defer func() {
		if removeSI {
			if err := s.customerRepo.DeleteSignInCode(siCookie.Param, dpi.Store); err != nil {
				// Notify me unable to delete tfa redis which should be
				log.Printf("Unable to delete sign in code: " + err.Error())
			}
		}
		if err := s.customerRepo.UnsetTwoFactorNX(siCookie.Param, dpi.Store); err != nil {
			// Notify me sign in nx not working
			log.Printf("Unable to unset sign in nx: " + err.Error())
		}
	}()

	signinParams, err := s.customerRepo.GetSignInEmail(siCookie.Param, dpi.Store)
	if err != nil {
		return nil, err
	}
	removeSI = true

	if signinParams.Param != siCookie.Param {
		return nil, errors.New("params do not match")
	}

	if time.Since(signinParams.Set) > time.Duration(config.SIGNIN_EXPIR_MINS)*time.Minute {
		// Alert me about it, since it should have expired on redis too
		return nil, errors.New("expired code, not deleted in system")
	}

	if signinParams.Tries >= config.MAX_SICODE_ATTEMPTS {
		return nil, errors.New("two many failed attempts")
	}

	if siCookie.IsCustomer != signinParams.HasCustomer {
		return nil, errors.New("unmatching cookie and redis if customer exists")
	}

	if signinParams.SixDigitCode != sixdigits {
		signinParams.Tries++
		if signinParams.Tries >= config.MAX_SICODE_ATTEMPTS {
			return nil, errors.New("two many failed attempts")
		}
		removeSI = false
		if err := s.customerRepo.StoreSignInEmail(signinParams, dpi.Store); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if siCookie.IsCustomer {
		if siCookie.CustomerID != signinParams.CustomerID {
			return nil, errors.New("incorrect customer between dpi and two factor server side")
		}

		cust, err := s.customerRepo.Read(siCookie.CustomerID)
		if err != nil {
			return nil, err
		}

		if cust.Email != signinParams.EmailAtTime {
			return nil, errors.New("customer email changed")
		}

		if cust == nil || cust.Status == "Archived" && time.Since(cust.Archived) > 7*24*time.Hour {
			return nil, errors.New("no active customer to log in")
		}

		if cust.Uses2FA {
			return nil, errors.New("cannot log in if uses 2FA")
		}

		// other status stuff
		client, _, err := s.LoginCookie(dpi, signinParams.EmailAtTime, "", false, false, tools)
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	post.IsEmailVerified = true
	client, _, _, _, err := s.CreateCustomer(dpi, post, tools)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (s *customerService) CreateTwoFACode(cust *models.Customer, store string, tools *config.Tools) (*models.TwoFactorCookie, error) {

	code := "TF-" + uuid.NewString()
	sixdigit := uint(100000 + rand.Intn(900000))
	setTime := time.Now()
	param := models.TwoFactorEmailParam{Param: code, CustomerID: cust.ID, Set: setTime, SixDigitCode: sixdigit, Tries: 0}
	cookie := models.TwoFactorCookie{TwoFactorCode: code, CustomerID: cust.ID, Set: setTime}

	if err := emails.TwoFactorEmail(store, cust.Email, sixdigit, tools); err != nil {
		return &cookie, err
	}

	return &cookie, s.customerRepo.StoreTwoFactor(param, store)
}

// Is correct six digit (recoverable), other error (not recoverable)
func (s *customerService) ProcessTwoFactor(dpi *DataPassIn, twofactorcookie *models.TwoFactorCookie, sixdigits uint) (bool, error) {
	if time.Since(twofactorcookie.Set) > config.TWOFA_EXPIR_MINS*time.Minute {
		return false, errors.New("past expiration")
	}

	if dpi.CustomerID != twofactorcookie.CustomerID {
		return false, errors.New("incorrect customer between dpi and two factor cookie")
	}

	if err := s.customerRepo.SetTwoFactorNX(twofactorcookie.TwoFactorCode, dpi.Store); err != nil {
		return false, err
	}

	removeTFA := false
	defer func() {
		if removeTFA {
			if err := s.customerRepo.DeleteTwoFactor(twofactorcookie.TwoFactorCode, dpi.Store); err != nil {
				// Notify me unable to delete tfa redis which should be
				log.Printf("Unable to delete two factor: " + err.Error())
			}
		}
		if err := s.customerRepo.UnsetTwoFactorNX(twofactorcookie.TwoFactorCode, dpi.Store); err != nil {
			// Notify me two factor nx not working
			log.Printf("Unable to unset two factor nx: " + err.Error())
		}
	}()

	serverTwoFA, err := s.customerRepo.GetTwoFactor(twofactorcookie.TwoFactorCode, dpi.Store)
	if err != nil {
		return false, err
	}
	removeTFA = true

	if time.Since(serverTwoFA.Set) > config.TWOFA_EXPIR_MINS*time.Minute {
		return false, errors.New("past expiration server side")
	}

	if serverTwoFA.Tries >= config.MAX_TWOFA_ATTEMPTS {
		return false, errors.New("two many failed attempts")
	}

	if serverTwoFA.CustomerID != dpi.CustomerID {
		return false, errors.New("incorrect customer between dpi and two factor server side")
	}

	if serverTwoFA.Param != twofactorcookie.TwoFactorCode {
		return false, errors.New("incorrect two factor code between cookie and server side")
	}

	if serverTwoFA.SixDigitCode != sixdigits {
		serverTwoFA.Tries++
		if serverTwoFA.Tries >= config.MAX_TWOFA_ATTEMPTS {
			return false, errors.New("two many failed attempts")
		}
		removeTFA = false
		if err := s.customerRepo.StoreTwoFactor(serverTwoFA, dpi.Store); err != nil {
			return false, err
		}
		return false, nil
	}

	if err := s.ToggleEmailVerified(dpi, true); err != nil {
		return false, err
	}

	return true, nil
}

func (s *customerService) ChangeCustomerEmail(dpi *DataPassIn, newEmail, password string, tools *config.Tools) error {

	newEmail = strings.ToLower(newEmail)

	if !custhelp.VerifyEmail(newEmail, tools) {
		return errors.New("email not valid")
	}

	customer, err := s.customerRepo.Read(dpi.CustomerID)
	if err != nil {
		return err
	} else if customer == nil {
		return errors.New("nil customer")
	} else if customer.Status == "Archived" {
		return errors.New("archived customer")
	}

	if time.Since(customer.EmailChanged) < config.EMAIL_CHANGE_COOLDOWN*24*time.Hour {
		return errors.New("email changed less than cooldown period ago")
	}

	if customer.PasswordHash == "" {
		return errors.New("customer must have password to change email")
	}

	if !custhelp.CheckPassword(customer.PasswordHash, password) {
		return errors.New("wrong password")
	}

	changedTime := time.Now()
	customer.Email = newEmail
	customer.EmailVerified = false
	customer.EmailChanged = changedTime

	if err := s.customerRepo.Update(*customer); err != nil {
		return errors.New("unable to save customer email change: " + err.Error())
	}

	serverCookie := models.ServerCookie{
		Store:            dpi.Store,
		CustomerID:       customer.ID,
		LastForcedLogout: changedTime,
		Archived:         false,
		Incomplete:       true,
	}

	if err := s.customerRepo.SetServerCookie(&serverCookie); err != nil {
		// Notify me but not return error
		log.Printf("Unable to set forced logout date\n")
	}

	return nil
}

func (s *customerService) ActualEmailVerification(store, param string, customer *models.Customer, tools *config.Tools) error {
	if err := emails.VerificationEmail(store, customer.Email, param, tools); err != nil {
		return err
	}
	customer.LastConfirmSent = time.Now()
	customer.ConfirmInProgress = false
	if time.Since(customer.LastConfirmSent) >= config.CONFIRM_EMAIL_COOLDOWN {
		customer.ConfirmsSent = 1
	} else {
		customer.ConfirmsSent++
	}

	return s.customerRepo.Update(*customer)
}

func (s *customerService) DelayedEmailVerification(store, param string, customerID int, wait time.Duration, tools *config.Tools) error {
	if wait > config.CONFIRM_EMAIL_WAIT*time.Second {
		wait = config.CONFIRM_EMAIL_WAIT * time.Second
	}

	time.Sleep(wait)

	cust, err := s.customerRepo.Read(customerID)
	if err != nil {
		return err
	} else if cust == nil {
		return errors.New("nil customer")
	} else if cust.Status == "Archived" {
		return errors.New("archived customer")
	}

	return s.ActualEmailVerification(store, param, cust, tools)
}

func (s *customerService) SendVerificationToEmail(store, param string, customer *models.Customer, tools *config.Tools) error {
	if customer.ConfirmsSent > config.CONFIRM_EMAIL_MAX && time.Since(customer.LastConfirmSent) < config.CONFIRM_EMAIL_COOLDOWN*time.Hour {
		return errors.New("too many confirms attempted within time period")
	}

	if customer.ConfirmInProgress {
		return nil
	}

	if time.Since(customer.LastConfirmSent) < config.CONFIRM_EMAIL_WAIT*time.Second {
		customer.ConfirmInProgress = true
		if err := s.customerRepo.Update(*customer); err != nil {
			return errors.New("unable to update customer to confirm in progress: " + err.Error())
		}

		go func() {
			duration := time.Until(customer.LastConfirmSent.Add(config.CONFIRM_EMAIL_WAIT * time.Second))
			if err := s.DelayedEmailVerification(store, param, customer.ID, duration, tools); err != nil {
				// Notify me that delay didn't work to send after wait period
				log.Printf("unable to ")
			}
		}()
		return nil
	}

	return s.ActualEmailVerification(store, param, customer, tools)
}

// Customer ID encr, timestamp encr, store check encr
func (s *customerService) UnsubLinkForEmails(store string, customerID int) (string, string, string) {
	return config.EncryptInt(customerID), config.EncodeTime(time.Now()), config.EncryptString(store)
}

func (s *customerService) UnsubCustomerDirect(store, storeEncr, customerEncr, timestamp string) error {

	storeCheck, err := config.DecryptString(storeEncr)
	if err != nil {
		return err
	}

	if storeCheck != store {
		return errors.New("store does not match up with check in id")
	}

	customerID, err := config.DecryptInt(customerEncr)
	if err != nil {
		return err
	}

	actualTime, err := config.DecodeTime(timestamp)
	if err != nil {
		return err
	}

	cust, err := s.customerRepo.Read(customerID)
	if err != nil {
		return err
	}

	if !cust.EmailSubbed {
		return nil
	}

	if cust.Created.Before(actualTime) || time.Now().Before(actualTime) {
		return errors.New("impossible timestamp")
	}

	if cust.EmailChanged.Before(actualTime) {
		return errors.New("email address changed after link was sent")
	}

	return s.customerRepo.SetEmailSubbed(customerID, false)
}

func (s *customerService) SendResetEmail(dpi *DataPassIn, email string, tools *config.Tools) (string, error) {
	email = strings.ToLower(email)

	cust, err := s.customerRepo.GetCustomerByEmail(email)
	if err != nil {
		return "", err
	} else if cust != nil && cust.Status == "Archived" && time.Since(cust.Archived) <= 7*24*time.Hour {
		return "", errors.New("recently archived customer")
	}

	if !cust.EmailVerified {
		return "", errors.New("cannot reset for an unverified email")
	}

	if !custhelp.VerifyEmail(email, tools) {
		// Notify me an existing customer's email not deliverable
		return "", errors.New("email found to be undeliverable")
	}

	if time.Since(cust.LastReset) < config.RESET_PASS_COOLDOWN*24*time.Hour {
		return "", errors.New("password reset too recently to reset again")
	}

	id := "RS-" + uuid.NewString()
	secret := "SC-" + config.GenerateRandomString()
	storedParam := models.ResetEmailParam{Param: id, EmailAtTime: email, CustomerID: cust.ID, Set: time.Now(), SecretCode: secret}

	return id, s.customerRepo.StoreResetEmail(storedParam, dpi.Store)
}

func (s *customerService) ProcessResetEmail(dpi *DataPassIn, param string) (*models.ResetEmailCookie, error) {
	resetParams, err := s.customerRepo.GetResetEmail(param, dpi.Store)
	if err != nil {
		return nil, err
	}

	if resetParams.Param != param {
		return nil, errors.New("params do not match")
	}

	if time.Since(resetParams.Set) > time.Duration(config.RESET_EXPIR_MINS)*time.Minute {
		// Alert me about it, since it should have expired on redis too
		return nil, errors.New("expired code, not deleted in system")
	}

	cust, err := s.customerRepo.Read(resetParams.CustomerID)
	if err != nil {
		return nil, err
	}

	if cust == nil || cust.ID != resetParams.CustomerID {
		return nil, errors.New("unable to retrieve correct customer")
	}

	if cust.Status == "Archived" {
		return nil, errors.New("archived customer")
	}

	if cust.Email != resetParams.EmailAtTime {
		return nil, errors.New("email was changed, cannot verify")
	}

	if !cust.EmailVerified {
		// Alert me about it, since if the email is the same and it was previously verified, something off
		return nil, errors.New("non verified email for customer")
	}

	if time.Since(cust.LastReset) < config.RESET_PASS_COOLDOWN*24*time.Hour {
		return nil, errors.New("password reset too recently to reset again")
	}

	return &models.ResetEmailCookie{
		Param:      resetParams.Param,
		SecretCode: resetParams.SecretCode,
		CustomerID: resetParams.CustomerID,
		Set:        time.Now(),
		Initial:    resetParams.Set,
	}, nil

}

func (s *customerService) ResetPasswordActual(dpi *DataPassIn, resetCookie *models.ResetEmailCookie, password, passwordConfirm string, logAllOut bool) error {

	if !custhelp.PasswordMeetsRequirements(password, passwordConfirm, true) {
		return errors.New("password does not meet requirements")
	}

	resetParams, err := s.customerRepo.GetResetEmail(resetCookie.Param, dpi.Store)
	if err != nil {
		return err
	}

	if resetParams.Param != resetCookie.Param {
		return errors.New("params do not match")
	}

	if time.Since(resetParams.Set) > time.Duration(config.RESET_EXPIR_MINS)*time.Minute {
		// Alert me about it, since it should have expired on redis too
		return errors.New("expired code, not deleted in system")
	}

	if resetParams.SecretCode != resetCookie.SecretCode {
		return errors.New("secret codes do not match")
	}

	if resetParams.CustomerID != resetCookie.CustomerID {
		return errors.New("customer ids do not match")
	}

	cust, err := s.customerRepo.Read(resetParams.CustomerID)
	if err != nil {
		return err
	}

	if cust == nil || cust.ID != resetParams.CustomerID {
		return errors.New("unable to retrieve correct customer")
	}

	if cust.Status == "Archived" {
		return errors.New("archived customer")
	}

	if cust.Email != resetParams.EmailAtTime {
		return errors.New("email was changed, cannot verify")
	}

	if !cust.EmailVerified {
		// Alert me about it, since if the email is the same and it was previously verified, something off
		return errors.New("non verified email for customer")
	}

	if time.Since(cust.LastReset) < config.RESET_PASS_COOLDOWN*24*time.Hour {
		return errors.New("password reset too recently to reset again")
	}

	cust.LastReset = time.Now()

	hash, err := custhelp.EncryptPassword(password)
	if err != nil {
		// Notify me that password unable to be hashed even though met requirements
		return err
	}
	cust.PasswordHash = hash

	if err := s.customerRepo.Update(*cust); err != nil {
		return err
	}

	if logAllOut {
		c, err := s.customerRepo.GetServerCookie(cust.ID, dpi.Store)
		if err != nil {
			// Notify me somehow user can't get server cookie
			return err
		} else if c == nil {
			// Notify me somehow user can't get server cookie
			return errors.New("nil server cookie")
		}

		if _, err := s.customerRepo.SetServerCookieReset(c, cust.LastReset); err != nil {
			return err
		}
	}

	return nil
}

func (s *customerService) BirthdayEmails(store string, ds DiscountService, tools *config.Tools) error {
	currentDate := time.Now()
	day := currentDate.Day()
	month := int(currentDate.Month())

	custs, err := s.customerRepo.ReadByBirthday(month, day)
	if err != nil {
		return err
	}

	var secondCusts []*models.Customer
	if day == 1 && month == 3 {
		secondCusts, err = s.customerRepo.ReadByBirthday(2, 29)
		if err != nil {
			// just notify me
		}
	}

	discCode := fmt.Sprintf("%02d-%02d-%d-%04d", month, day, time.Now().Year(), rand.Intn(10000))
	disc := &models.Discount{
		DiscountCode:    discCode,
		Status:          "Active",
		Created:         time.Now(),
		Expired:         time.Now().AddDate(0, 1, 1),
		IsPercentageOff: true,
		PercentageOff:   0.15,
		HasMinSubtotal:  true,
		MinSubtotal:     9000,
		AppliesToAllAny: true,
		ShortMessage:    "Happy Birthday :P",
	}

	if err := ds.AddDiscount(*disc); err != nil {
		return err
	}

	for _, cust := range custs {
		if cust == nil {
			continue
		}
		emails.CustBirthdayEmail(store, cust.Email, discCode, cust, false, tools)
	}

	for _, cust := range secondCusts {
		if cust == nil {
			continue
		}
		emails.CustBirthdayEmail(store, cust.Email, discCode, cust, true, tools)
	}

	return nil
}

func (s *customerService) ResendSignInCode(dpi *DataPassIn, siCookie models.SignInCodeCookie, tools *config.Tools) (models.SignInCodeCookie, error) {
	if time.Since(siCookie.Set) > config.SIGNIN_EXPIR_MINS*time.Minute {
		return siCookie, errors.New("past expiration")
	}

	if err := s.customerRepo.SetSignInCodeNX(siCookie.Param, dpi.Store); err != nil {
		return siCookie, err
	}

	removeSI := false
	defer func() {
		if removeSI {
			if err := s.customerRepo.DeleteSignInCode(siCookie.Param, dpi.Store); err != nil {
				// Notify me unable to delete tfa redis which should be
				log.Printf("Unable to delete sign in code: " + err.Error())
			}
		}
		if err := s.customerRepo.UnsetTwoFactorNX(siCookie.Param, dpi.Store); err != nil {
			// Notify me sign in nx not working
			log.Printf("Unable to unset sign in nx: " + err.Error())
		}
	}()

	signinParams, err := s.customerRepo.GetSignInEmail(siCookie.Param, dpi.Store)
	if err != nil {
		return siCookie, err
	}
	removeSI = true

	if signinParams.Param != siCookie.Param {
		return siCookie, errors.New("params do not match")
	}

	if time.Since(signinParams.Set) > time.Duration(config.SIGNIN_EXPIR_MINS)*time.Minute {
		// Alert me about it, since it should have expired on redis too
		return siCookie, errors.New("expired code, not deleted in system")
	}

	if signinParams.Tries >= config.MAX_SICODE_ATTEMPTS {
		return siCookie, errors.New("two many failed attempts")
	}

	if siCookie.IsCustomer != signinParams.HasCustomer {
		return siCookie, errors.New("unmatching cookie and redis if customer exists")
	}

	if siCookie.IsCustomer {
		if siCookie.CustomerID != signinParams.CustomerID {
			return siCookie, errors.New("incorrect customer between dpi and two factor server side")
		}

		cust, err := s.customerRepo.Read(siCookie.CustomerID)
		if err != nil {
			return siCookie, err
		}

		if cust.Email != signinParams.EmailAtTime {
			return siCookie, errors.New("customer email changed")
		}

		if cust == nil || cust.Status == "Archived" && time.Since(cust.Archived) > 7*24*time.Hour {
			return siCookie, errors.New("no active customer to log in")
		}

		if cust.Uses2FA {
			return siCookie, errors.New("cannot log in if uses 2FA")
		}
	}

	removeSI = false
	signinParams.SixDigitCode = uint(100000 + rand.Intn(900000))
	signinParams.Set = time.Now()
	siCookie.Set = signinParams.Set

	if err := s.customerRepo.StoreSignInEmail(signinParams, dpi.Store); err != nil {
		return siCookie, err
	}

	if err := emails.SignInPin(dpi.Store, signinParams.EmailAtTime, signinParams.SixDigitCode, tools); err != nil {
		return siCookie, err
	}

	return siCookie, nil
}

func (s *customerService) ResendTwoFactor(dpi *DataPassIn, twofactorcookie models.TwoFactorCookie, tools *config.Tools) (models.TwoFactorCookie, error) {
	if time.Since(twofactorcookie.Set) > config.TWOFA_EXPIR_MINS*time.Minute {
		return twofactorcookie, errors.New("past expiration")
	}

	if dpi.CustomerID != twofactorcookie.CustomerID {
		return twofactorcookie, errors.New("incorrect customer between dpi and two factor cookie")
	}

	if err := s.customerRepo.SetTwoFactorNX(twofactorcookie.TwoFactorCode, dpi.Store); err != nil {
		return twofactorcookie, err
	}

	removeTFA := false
	defer func() {
		if removeTFA {
			if err := s.customerRepo.DeleteTwoFactor(twofactorcookie.TwoFactorCode, dpi.Store); err != nil {
				// Notify me unable to delete tfa redis which should be
				log.Printf("Unable to delete two factor: " + err.Error())
			}
		}
		if err := s.customerRepo.UnsetTwoFactorNX(twofactorcookie.TwoFactorCode, dpi.Store); err != nil {
			// Notify me two factor nx not working
			log.Printf("Unable to unset two factor nx: " + err.Error())
		}
	}()

	serverTwoFA, err := s.customerRepo.GetTwoFactor(twofactorcookie.TwoFactorCode, dpi.Store)
	if err != nil {
		return twofactorcookie, err
	}
	removeTFA = true

	if time.Since(serverTwoFA.Set) > config.TWOFA_EXPIR_MINS*time.Minute {
		return twofactorcookie, errors.New("past expiration server side")
	}

	if serverTwoFA.Tries >= config.MAX_TWOFA_ATTEMPTS {
		return twofactorcookie, errors.New("two many failed attempts")
	}

	if serverTwoFA.CustomerID != dpi.CustomerID {
		return twofactorcookie, errors.New("incorrect customer between dpi and two factor server side")
	}

	if serverTwoFA.Param != twofactorcookie.TwoFactorCode {
		return twofactorcookie, errors.New("incorrect two factor code between cookie and server side")
	}

	if twofactorcookie.CustomerID != serverTwoFA.CustomerID {
		return twofactorcookie, errors.New("incorrect customer between dpi and two factor server side")
	}

	cust, err := s.customerRepo.Read(serverTwoFA.CustomerID)
	if err != nil {
		return twofactorcookie, err
	}

	if cust == nil || cust.Status == "Archived" && time.Since(cust.Archived) > 7*24*time.Hour {
		return twofactorcookie, errors.New("no active customer to log in")
	}

	removeTFA = false
	serverTwoFA.SixDigitCode = uint(100000 + rand.Intn(900000))
	serverTwoFA.Set = time.Now()
	twofactorcookie.Set = serverTwoFA.Set

	if err := s.customerRepo.StoreTwoFactor(serverTwoFA, dpi.Store); err != nil {
		return twofactorcookie, err
	}

	if err := emails.TwoFactorEmail(dpi.Store, cust.Email, serverTwoFA.SixDigitCode, tools); err != nil {
		return twofactorcookie, err
	}

	return twofactorcookie, nil
}
