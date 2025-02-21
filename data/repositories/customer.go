package repositories

import (
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
	CheckFirebaseUID(firebaseUID string) (int, string, error)
	GetCustomerByFirebase(firebaseUID string) (*models.Customer, error)
	GetServerCookie(custID int, store string) (*models.ServerCookie, error)
	SetServerCookieReset(c *models.ServerCookie, reset time.Time) (*models.ServerCookie, error)
	SetServerCookieStatus(c *models.ServerCookie, archived bool) (*models.ServerCookie, error)
	CreateServerCookie(customerID int, store string) (*models.ServerCookie, error)
	GetCustomerIDByEmail(email string) (int, bool, error)

	ArchiveCustomerEmail(id int, email string) error
}

type customerRepo struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewCustomerRepository(db *gorm.DB, rdb *redis.Client) CustomerRepository {
	return &customerRepo{db: db, rdb: rdb}
}

func (r *customerRepo) Create(customer models.Customer) error {
	return r.db.Create(&customer).Error
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

	strCust, err := draftorderhelp.CreateCustomer(c.Email, c.DefaultName)
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

func (r *customerRepo) CheckFirebaseUID(firebaseUID string) (int, string, error) {
	var result struct {
		ID     int
		Status string
	}

	err := r.db.Model(&models.Customer{}).
		Select("id, status").
		Where("firebase_uid = ?", firebaseUID).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, "", nil
		}
		return 0, "", err
	}

	return result.ID, result.Status, nil
}

func (r *customerRepo) GetCustomerByFirebase(firebaseUID string) (*models.Customer, error) {
	var cust models.Customer
	err := r.db.Where("firebase_uid = ?", firebaseUID).Find(&cust).Error
	return &cust, err
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

func (r *customerRepo) CreateServerCookie(customerID int, store string) (*models.ServerCookie, error) {
	c := models.ServerCookie{
		CustomerID:       customerID,
		Store:            store,
		Archived:         false,
		LastForcedLogout: time.Time{},
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
