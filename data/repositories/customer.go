package repositories

import (
	"beam/config"
	"beam/data/models"
	"beam/data/services/draftorderhelp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	Create(customer models.Customer) error
	Read(id int) (*models.Customer, error)
	Update(customer models.Customer) error
	UpdateContact(contact models.Contact) error
	Delete(id int) error
	DeleteContact(id int) error
	GetSingleContact(id int) (*models.Contact, error)
	GetContactsWithDefault(customerID int) ([]*models.Contact, error)
	AddContactToCustomer(contact *models.Contact) error
	GetCustomerAndContacts(customerID int) (*models.Customer, []*models.Contact, error)
	AddStripeToCustomer(c *models.Customer)
	GetPaymentMethodsCust(customerID int) ([]models.PaymentMethodStripe, error)
	UpdateCustomerDefault(customerID, contactID int) error
	GetServerCookie(custID int, store string) (*models.ServerCookie, error)
	SetServerCookieReset(c *models.ServerCookie, reset time.Time) (*models.ServerCookie, error)
	SetServerCookieStatus(c *models.ServerCookie, archived bool) (*models.ServerCookie, error)
	SetServerCookieCompletion(c *models.ServerCookie, verified bool) (*models.ServerCookie, error)
	CreateServerCookie(customerID int, store string, verified, archived bool) (*models.ServerCookie, error)
	SetServerCookie(c *models.ServerCookie) error
	GetCustomerIDByEmail(email string) (int, bool, error)

	ArchiveCustomerEmail(id int, email string) error
	GetActiveCustomerByEmail(email string) (*models.Customer, error)
	GetCustomerByEmail(email string) (*models.Customer, error)
	SetEmailSubbed(id int, subbed bool) error
	SetEmailVerified(id int, verif bool) error
	IsEmailVerified(id int) (bool, error)

	StoreVerificationEmail(param models.VerificationEmailParam, store string) error
	GetVerificationEmail(param, store string) (models.VerificationEmailParam, error)

	StoreSignInEmail(param models.SignInEmailParam, store string) error
	GetSignInEmail(param, store string) (models.SignInEmailParam, error)
	DeleteSignInCode(param, store string) error
	SetSignInCodeNX(param, store string) error
	UnsetSignInCodeNX(param, store string) error

	StoreTwoFactor(param models.TwoFactorEmailParam, store string) error
	DeleteTwoFactor(param, store string) error
	GetTwoFactor(param, store string) (models.TwoFactorEmailParam, error)
	SetTwoFactorNX(param, store string) error
	UnsetTwoFactorNX(param, store string) error

	UpdateCustomerCurrency(id int, usesOtherCurrency bool, otherCurrency string) error

	SetDeviceMapping(customerID int, guestID, store string) error
	GetDeviceMapping(guestID, store string) (int, error)
	DeleteIncompleteUnverifiedCustomers() error
	IncompleteScheduled()

	StoreResetEmail(param models.ResetEmailParam, store string) error
	GetResetEmail(param, store string) (models.ResetEmailParam, error)
	DeleteResetEmail(param, store string) error

	ReadByBirthday(birthMonth, birthDay int) ([]*models.Customer, error)

	CheckPasswordFailedAttempts(store, guestID string, customerID int) (bool, error)
	SetPasswordFailedAttempts(store, guestID string, customerID int) (bool, error)
	SuccessfulPasswordAttempt(store, guestID string, customerID int) error

	SaveAuthParams(param models.LoginSpecificParams, store string) error
	GetAuthParams(param, store string) (models.LoginSpecificParams, error)
	RemoveAuthParams(param, store string) error
}

type customerRepo struct {
	db         *gorm.DB
	rdb        *redis.Client
	saveTicker *time.Ticker
	store      string
}

func NewCustomerRepository(db *gorm.DB, rdb *redis.Client, store string, ct, len int) CustomerRepository {
	repo := &customerRepo{db: db, rdb: rdb, store: store}

	go func() {
		defer repo.saveTicker.Stop()

		if len > 0 && ct >= 0 {
			delayFactor := float64(ct) / float64(len)
			if delayFactor > 1 {
				delayFactor = 1
			} else if delayFactor < 0 {
				delayFactor = 0
			}

			initialDelay := time.Duration(float64(config.SCHEDULED_INCOMPLETE_CUST) * delayFactor * float64(time.Minute))
			time.Sleep(initialDelay)
		}

		for range repo.saveTicker.C {
			repo.IncompleteScheduled()
		}
	}()

	return repo

}

func (r *customerRepo) Create(customer models.Customer) error {
	return r.db.Save(&customer).Error
}

func (r *customerRepo) Read(id int) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.First(&customer, id).Error
	return &customer, err
}

func (r *customerRepo) GetSingleContact(id int) (*models.Contact, error) {
	var cont models.Contact
	err := r.db.First(&cont, id).Error
	return &cont, err
}

func (r *customerRepo) Update(customer models.Customer) error {
	return r.db.Save(&customer).Error
}

func (r *customerRepo) UpdateContact(contact models.Contact) error {
	return r.db.Save(&contact).Error
}

func (r *customerRepo) Delete(id int) error {
	return r.db.Delete(&models.Customer{}, id).Error
}

func (r *customerRepo) DeleteContact(id int) error {
	return r.db.Delete(&models.Contact{}, id).Error
}

func (r *customerRepo) GetContactsWithDefault(customerID int) ([]*models.Contact, error) {
	var customer models.Customer
	var contacts []*models.Contact

	var customerErr, contactsErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		customerErr = r.db.Select("default_shipping_contact_id").First(&customer, customerID).Error
	}()

	go func() {
		defer wg.Done()
		contactsErr = r.db.Where("customer_id = ?", customerID).Find(&contacts).Error
	}()

	wg.Wait()

	if contactsErr != nil {
		return nil, contactsErr
	} else if customerErr != nil {
		return nil, customerErr
	}

	for i, contact := range contacts {
		if contact.ID == customer.DefaultShippingContactID {
			contacts = append([]*models.Contact{contact}, append(contacts[:i], contacts[i+1:]...)...)
			break
		}
	}

	return contacts, nil
}

func (r *customerRepo) AddContactToCustomer(contact *models.Contact) error {
	return r.db.Create(contact).Error
}

func (r *customerRepo) GetCustomerAndContacts(customerID int) (*models.Customer, []*models.Contact, error) {
	var customer models.Customer
	var contacts []*models.Contact
	var customerErr, contactsErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		customerErr = r.db.First(&customer, customerID).Error
	}()

	go func() {
		defer wg.Done()
		contactsErr = r.db.Where("customer_id = ?", customerID).Find(&contacts).Error
	}()

	wg.Wait()

	if customerErr != nil {
		return &customer, contacts, customerErr
	} else if contactsErr != nil {
		return &customer, contacts, contactsErr
	}

	for i, contact := range contacts {
		if contact.ID == customer.DefaultShippingContactID {
			contacts = append([]*models.Contact{contact}, append(contacts[:i], contacts[i+1:]...)...)
			break
		}
	}

	return &customer, contacts, nil
}

func (r *customerRepo) AddStripeToCustomer(c *models.Customer) {
	if c == nil {
		log.Printf("Tried to attach stripe id to nil customer\n")
		return
	}

	strCust, err := draftorderhelp.CreateCustomer(c.Email, c.FirstName, c.LastName)
	if err != nil {
		log.Printf("Unable to create stripe id for customer with id: %d; error: %v\n", c.ID, err)
		return
	}

	if strCust == nil {
		log.Printf("Unable to create stripe id for customer with id: %d; error: %v\n", c.ID, errors.New("no stripe customer returned (nil)"))
		return
	}

	c.StripeID = strCust.ID

	if err := r.Update(*c); err != nil {
		log.Printf("Unable to save stripe id: %s; for customer with id: %d; error: %v\n", strCust.ID, c.ID, err)
	}
}

func (r *customerRepo) GetPaymentMethodsCust(customerID int) ([]models.PaymentMethodStripe, error) {
	c, err := r.Read(customerID)
	if err != nil {
		return nil, err
	} else if c == nil {
		return nil, errors.New("no customer exists with ID")
	}

	if c.StripeID != "" {
		go r.AddStripeToCustomer(c)
		return nil, nil
	}

	return draftorderhelp.GetAllPaymentMethods(c.StripeID)
}

func (r *customerRepo) UpdateCustomerDefault(customerID, contactID int) error {
	return r.db.Model(&models.Customer{}).
		Where("id = ?", customerID).
		Update("default_shipping_contact_id", contactID).
		Error
}

func (r *customerRepo) GetServerCookie(custID int, store string) (*models.ServerCookie, error) {
	key := store + "::SSC::" + strconv.Itoa(custID)

	val, err := r.rdb.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var cookie models.ServerCookie
	if err := json.Unmarshal([]byte(val), &cookie); err != nil {
		return nil, err
	}

	if cookie.CustomerID != custID {
		return nil, fmt.Errorf("customer id doesn't match, provided: %d; in struct: %d", custID, cookie.CustomerID)
	} else if cookie.Store != store {
		return nil, fmt.Errorf("store doesn't match, provided: %s; in struct: %s", store, cookie.Store)
	}

	return &cookie, nil
}

func (r *customerRepo) SetServerCookieReset(c *models.ServerCookie, reset time.Time) (*models.ServerCookie, error) {
	c.LastForcedLogout = reset
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s::SSC::%d", c.Store, c.CustomerID)
	if err := r.rdb.Set(context.Background(), key, data, 0).Err(); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *customerRepo) SetServerCookieStatus(c *models.ServerCookie, archived bool) (*models.ServerCookie, error) {
	c.Archived = archived
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s::SSC::%d", c.Store, c.CustomerID)
	if err := r.rdb.Set(context.Background(), key, data, 0).Err(); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *customerRepo) SetServerCookieCompletion(c *models.ServerCookie, verified bool) (*models.ServerCookie, error) {
	c.Incomplete = !verified
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s::SSC::%d", c.Store, c.CustomerID)
	if err := r.rdb.Set(context.Background(), key, data, 0).Err(); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *customerRepo) CreateServerCookie(customerID int, store string, verified, archived bool) (*models.ServerCookie, error) {
	c := models.ServerCookie{
		CustomerID:       customerID,
		Store:            store,
		Archived:         archived,
		LastForcedLogout: time.Time{},
		Incomplete:       !verified,
	}
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s::SSC::%d", c.Store, c.CustomerID)
	if err := r.rdb.Set(context.Background(), key, data, 0).Err(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *customerRepo) SetServerCookie(c *models.ServerCookie) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s::SSC::%d", c.Store, c.CustomerID)
	return r.rdb.Set(context.Background(), key, data, 0).Err()
}

// ID if official match, whether there is a past 7 day archived account, error
func (r *customerRepo) GetCustomerIDByEmail(email string) (int, bool, error) {
	email = strings.ToLower(email)

	var customer models.Customer
	err := r.db.
		Select("id, archived, status").
		Where("email = ?", email).
		Limit(1).
		Find(&customer).Error

	if err != nil {
		return 0, false, err
	}

	if customer.ID == 0 {
		return 0, false, nil
	}

	if customer.Status == "Archived" && time.Since(customer.Archived) > 7*24*time.Hour {
		return 0, true, nil
	}

	return customer.ID, false, nil
}

func (r *customerRepo) ArchiveCustomerEmail(id int, email string) error {
	customer := &models.Customer{}
	email = "&ARCHIVED::" + email

	if err := r.db.Model(customer).Where("id = ?", id).Updates(models.Customer{
		Email:  email,
		Status: "Archived",
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *customerRepo) GetActiveCustomerByEmail(email string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("email = ? AND status = ?", email, "active").First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepo) GetCustomerByEmail(email string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("email = ?", email).First(&customer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepo) SetEmailSubbed(id int, subbed bool) error {
	return r.db.Model(&models.Customer{}).Where("id = ?", id).Update("email_subbed", subbed).Error
}

func (r *customerRepo) SetEmailVerified(id int, verif bool) error {
	status := "Incomplete"
	if verif {
		status = "Active"
	}
	return r.db.Model(&models.Customer{}).Where("id = ?", id).
		Updates(map[string]interface{}{"email_verified": verif, "status": status}).Error
}

func (r *customerRepo) IsEmailVerified(id int) (bool, error) {
	var verified bool
	err := r.db.Model(&models.Customer{}).Where("id = ?", id).Select("email_verified").Scan(&verified).Error
	return verified, err
}

func (r *customerRepo) StoreVerificationEmail(param models.VerificationEmailParam, store string) error {
	if param.Param == "" {
		return errors.New("param cannot be empty")
	}

	data, err := json.Marshal(param)
	if err != nil {
		return err
	}

	return r.rdb.Set(context.Background(), store+"::VFRE::"+param.Param, data, time.Duration(config.VERIF_EXPIR_MINS)*time.Minute).Err()
}

func (r *customerRepo) GetVerificationEmail(param, store string) (models.VerificationEmailParam, error) {
	if param == "" {
		return models.VerificationEmailParam{}, errors.New("param cannot be empty")
	}
	key := store + "::VFRE::" + param

	data, err := r.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return models.VerificationEmailParam{}, err
	}

	var result models.VerificationEmailParam
	if err := json.Unmarshal(data, &result); err != nil {
		return models.VerificationEmailParam{}, err
	}

	go func() {
		if err := r.rdb.Del(context.Background(), key).Err(); err != nil {
			log.Println("error deleting key:", err)
		}
	}()

	return result, nil
}

func (r *customerRepo) StoreSignInEmail(param models.SignInEmailParam, store string) error {
	if param.Param == "" {
		return errors.New("param cannot be empty")
	}

	if len(param.SixDigitCode) > config.MAX_SICODE_NEW+1 || param.NewCodeReqs > config.MAX_SICODE_NEW {
		return errors.New("over max new code requests")
	}

	for _, c := range param.SixDigitCode {
		if c < 100000 || c > 999999 {
			return errors.New("6 digit codes must be in range")
		}
	}

	data, err := json.Marshal(param)
	if err != nil {
		return err
	}

	return r.rdb.Set(context.Background(), store+"::SIPE::"+param.Param, data, time.Duration(config.SIGNIN_EXPIR_MINS)*time.Minute).Err()
}

func (r *customerRepo) GetSignInEmail(param, store string) (models.SignInEmailParam, error) {
	if param == "" {
		return models.SignInEmailParam{}, errors.New("param cannot be empty")
	}
	key := store + "::SIPE::" + param

	data, err := r.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return models.SignInEmailParam{}, err
	}

	var result models.SignInEmailParam
	if err := json.Unmarshal(data, &result); err != nil {
		return models.SignInEmailParam{}, err
	}

	return result, nil
}

func (r *customerRepo) SetSignInCodeNX(param, store string) error {

	key := store + "::SINX::" + param

	ok, err := r.rdb.SetNX(context.Background(), key, "1", 5*time.Second).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("key already set")
	}
	return nil
}

func (r *customerRepo) UnsetSignInCodeNX(param, store string) error {
	key := store + "::SINX::" + param
	return r.rdb.Del(context.Background(), key).Err()
}

func (r *customerRepo) DeleteSignInCode(param, store string) error {
	key := store + "::SIPE::" + param
	return r.rdb.Del(context.Background(), key).Err()
}

func (r *customerRepo) UpdateCustomerCurrency(id int, usesOtherCurrency bool, otherCurrency string) error {
	return r.db.Model(&models.Customer{}).Where("id = ?", id).Updates(map[string]interface{}{
		"UsesOtherCurrency": usesOtherCurrency,
		"OtherCurrency":     otherCurrency,
	}).Error
}

func (r *customerRepo) SetDeviceMapping(customerID int, guestID, store string) error {
	key := store + "::DVMP::" + guestID

	if customerID == 0 {
		return r.rdb.Del(context.Background(), key).Err()
	}

	return r.rdb.Set(context.Background(), key, customerID, 0).Err()
}

func (r *customerRepo) GetDeviceMapping(guestID, store string) (int, error) {
	key := store + "::DVMP::" + guestID
	val, err := r.rdb.Get(context.Background(), key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (r *customerRepo) StoreTwoFactor(param models.TwoFactorEmailParam, store string) error {
	if param.Param == "" {
		return errors.New("param cannot be empty")
	}

	if len(param.SixDigitCode) > config.MAX_TWOFA_NEW+1 || param.NewCodeReqs > config.MAX_TWOFA_NEW {
		return errors.New("over max new code requests")
	}

	for _, c := range param.SixDigitCode {
		if c < 100000 || c > 999999 {
			return errors.New("6 digit codes must be in range")
		}
	}

	data, err := json.Marshal(param)
	if err != nil {
		return err
	}

	return r.rdb.Set(context.Background(), store+"::TWFA::"+param.Param, data, time.Duration(config.TWOFA_EXPIR_MINS)*time.Minute).Err()
}

func (r *customerRepo) GetTwoFactor(param, store string) (models.TwoFactorEmailParam, error) {
	if param == "" {
		return models.TwoFactorEmailParam{}, errors.New("param cannot be empty")
	}
	key := store + "::TWFA::" + param

	data, err := r.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return models.TwoFactorEmailParam{}, err
	}

	var result models.TwoFactorEmailParam
	if err := json.Unmarshal(data, &result); err != nil {
		return models.TwoFactorEmailParam{}, err
	}

	return result, nil
}

func (r *customerRepo) SetTwoFactorNX(param, store string) error {

	key := store + "::TFNX::" + param

	ok, err := r.rdb.SetNX(context.Background(), key, "1", 5*time.Second).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("key already set")
	}
	return nil
}

func (r *customerRepo) UnsetTwoFactorNX(param, store string) error {
	key := store + "::TFNX::" + param
	return r.rdb.Del(context.Background(), key).Err()
}

func (r *customerRepo) DeleteTwoFactor(param, store string) error {
	key := store + "::TWFA::" + param
	return r.rdb.Del(context.Background(), key).Err()
}

func (r *customerRepo) DeleteIncompleteUnverifiedCustomers() error {
	cutoff := time.Now().Add(-24 * time.Hour)

	return r.db.Where("status = ? AND email_verified = ? AND created < ? AND email_changed < ?",
		"Incomplete", false, cutoff, cutoff).Delete(&models.Customer{}).Error
}

func (r *customerRepo) IncompleteScheduled() {
	key := r.store + "::FLNC"

	val, err := r.rdb.Get(context.Background(), key).Result()
	if err != nil && err != redis.Nil {
		log.Printf("Error checking Redis key %s: %v", key, err)
		return
	}

	if val != "" {
		return
	}

	err = r.rdb.SetEX(context.Background(), key, "1", config.SCHEDULED_INCOMPLETE_CUST*time.Minute).Err()
	if err != nil {
		log.Printf("Error setting Redis key %s: %v", key, err)
		return
	}

	err = r.DeleteIncompleteUnverifiedCustomers()
	if err != nil {
		log.Printf("Error deleting incomplete unverified customers: %v", err)
	}
}

func (r *customerRepo) StoreResetEmail(param models.ResetEmailParam, store string) error {
	if param.Param == "" {
		return errors.New("param cannot be empty")
	}

	data, err := json.Marshal(param)
	if err != nil {
		return err
	}

	return r.rdb.Set(context.Background(), store+"::RSCE::"+param.Param, data, time.Duration(config.RESET_EXPIR_MINS)*time.Minute).Err()
}

func (r *customerRepo) GetResetEmail(param, store string) (models.ResetEmailParam, error) {
	if param == "" {
		return models.ResetEmailParam{}, errors.New("param cannot be empty")
	}
	key := store + "::RSCE::" + param

	data, err := r.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return models.ResetEmailParam{}, err
	}

	var result models.ResetEmailParam
	if err := json.Unmarshal(data, &result); err != nil {
		return models.ResetEmailParam{}, err
	}

	return result, nil
}

func (r *customerRepo) DeleteResetEmail(param, store string) error {
	key := store + "::RSCE::" + param
	return r.rdb.Del(context.Background(), key).Err()
}

func (r *customerRepo) ReadByBirthday(birthMonth, birthDay int) ([]*models.Customer, error) {
	var customers []*models.Customer
	err := r.db.Where("email_subbed = ? AND status = ? AND birthday_set = ? AND birth_month = ? AND birth_day = ?",
		true, "Active", true, birthMonth, birthDay).Find(&customers).Error
	return customers, err
}

func (r *customerRepo) CheckPasswordFailedAttempts(store, guestID string, customerID int) (bool, error) {
	key := store + "::PWFA::" + guestID + "::" + strconv.Itoa(customerID)

	val, err := r.rdb.Get(context.Background(), key).Int()
	return val >= config.PASSWORD_MAX_ATTEMPTS, err
}

func (r *customerRepo) SetPasswordFailedAttempts(store, guestID string, customerID int) (bool, error) {
	key := store + "::PWFA::" + guestID + "::" + strconv.Itoa(customerID)

	val, err := r.rdb.Incr(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	err = r.rdb.Expire(context.Background(), key, config.PASSWORD_FAIL_COOLWDOWN*time.Hour).Err()
	return val >= config.PASSWORD_MAX_ATTEMPTS, err
}

func (r *customerRepo) SuccessfulPasswordAttempt(store, guestID string, customerID int) error {
	key := store + "::PWFA::" + guestID + "::" + strconv.Itoa(customerID)

	return r.rdb.Del(context.Background(), key).Err()
}

func (r *customerRepo) SaveAuthParams(param models.LoginSpecificParams, store string) error {
	if param.Param == "" {
		return errors.New("param cannot be empty")
	}

	data, err := json.Marshal(param)
	if err != nil {
		return err
	}

	return r.rdb.Set(context.Background(), store+"::AUSP::"+param.Param, data, config.AUTH_PARAMS_EXPIR*time.Hour).Err()
}

func (r *customerRepo) GetAuthParams(param, store string) (models.LoginSpecificParams, error) {
	if param == "" {
		return models.LoginSpecificParams{}, errors.New("param cannot be empty")
	}
	key := store + "::AUSP::" + param

	data, err := r.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return models.LoginSpecificParams{}, err
	}

	var result models.LoginSpecificParams
	if err := json.Unmarshal(data, &result); err != nil {
		return models.LoginSpecificParams{}, err
	}

	return result, nil
}

func (r *customerRepo) RemoveAuthParams(param, store string) error {
	return r.rdb.Del(context.Background(), store+"::AUSP::"+param).Err()
}
