package repositories

import (
	"beam/data/models"
	"sync"

	"gorm.io/gorm"
)

type CustomerRepository interface {
	Create(customer models.Customer) error
	Read(id int) (*models.Customer, error)
	Update(customer models.Customer) error
	Delete(id int) error
	GetContactsWithDefault(customerID int) ([]*models.Contact, error)
	AddContactToCustomer(customerID int, contact *models.Contact) error
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

func (r *customerRepo) AddContactToCustomer(customerID int, contact *models.Contact) error {
	contact.CustomerID = customerID
	return r.db.Create(contact).Error
}
