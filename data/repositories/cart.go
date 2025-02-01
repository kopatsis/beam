package repositories

import (
	"beam/data/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type CartRepository interface {
	Create(cart models.Cart) error
	Read(id int) (*models.Cart, error)
	Update(cart models.Cart) error
	Delete(id int) error
	DeleteCartLine(line models.CartLine) error
	GetCartWithLinesByIDAndCustomerID(cartID, customerID int) (models.Cart, []models.CartLine, bool, error)
	GetCartWithLinesByIDAndGuestID(cartID int, guestID string) (models.Cart, []models.CartLine, bool, error)
	GetCartWithLinesByCustomerID(customerID int) (models.Cart, []models.CartLine, bool, error)
	GetCartWithLinesByGuestID(guestID string) (models.Cart, []models.CartLine, bool, error)
	CreateCart(cart models.Cart) (models.Cart, error)
	SaveCart(cart models.Cart) (models.Cart, error)
	AddCartLine(cartLine models.CartLine) (models.CartLine, error)
	SaveCartLine(cartLine models.CartLine) (models.CartLine, error)
	DeleteCartWithLines(id int) error

	GetCartLineWithValidation(customerID, cartID, lineID int) (*models.CartLine, error)
	MostRecentAllowedCart(customerID int) (*models.Cart, error)
	TotalQuantity(cartID int) (int, error)

	ReadWithPreload(id int) (*models.Cart, error)
	CartLinesRetrieval(cartID int) ([]*models.CartLine, error)
	ArchiveCart(id int) error
	ActiveCart(id int) error
	ReactivateCartWithLines(cartID int, newLines []models.CartLine) error

	CopyCartWithLines(cartID, newCustomer int) error
	MoveCart(cartID, newCustomer int) error
	DirectCartRetrieval(cartID, customerID int, guestID string) (int, error, bool)
}

type cartRepo struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepo{db: db}
}

func (r *cartRepo) Create(cart models.Cart) error {
	return r.db.Create(&cart).Error
}

func (r *cartRepo) Read(id int) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.First(&cart, id).Error
	return &cart, err
}

func (r *cartRepo) ReadWithPreload(id int) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.Preload("CartLines").First(&cart, id).Error
	return &cart, err
}

func (r *cartRepo) CartLinesRetrieval(cartID int) ([]*models.CartLine, error) {
	var cartLines []*models.CartLine
	err := r.db.Where("cart_id = ?", cartID).Find(&cartLines).Error
	return cartLines, err
}

func (r *cartRepo) Update(cart models.Cart) error {
	return r.db.Save(&cart).Error
}

func (r *cartRepo) Delete(id int) error {
	return r.db.Delete(&models.Cart{}, id).Error
}

func (r *cartRepo) DeleteCartLine(line models.CartLine) error {
	return r.db.Delete(&line).Error
}

func (r *cartRepo) GetCartWithLinesByCustomerID(customerID int) (models.Cart, []models.CartLine, bool, error) {
	var cart models.Cart
	var cartLines []models.CartLine
	err := r.db.Preload("CartLines").Where("customer_id = ? AND status = ?", customerID, "Active").First(&cart).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return cart, cartLines, true, nil
		}
		return models.Cart{}, nil, false, err
	}
	err = r.db.Where("cart_id = ?", cart.ID).Find(&cartLines).Error
	if err != nil {
		return models.Cart{}, nil, false, err
	}
	return cart, cartLines, false, nil
}

func (r *cartRepo) GetCartWithLinesByGuestID(guestID string) (models.Cart, []models.CartLine, bool, error) {
	var cart models.Cart
	var cartLines []models.CartLine
	err := r.db.Preload("CartLines").Where("guest_id = ? AND status = ?", guestID, "Active").First(&cart).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return cart, cartLines, true, nil
		}
		return models.Cart{}, nil, false, err
	}
	err = r.db.Where("cart_id = ?", cart.ID).Find(&cartLines).Error
	if err != nil {
		return models.Cart{}, nil, false, err
	}
	return cart, cartLines, false, nil
}

func (r *cartRepo) GetCartWithLinesByIDAndCustomerID(cartID, customerID int) (models.Cart, []models.CartLine, bool, error) {
	var cart models.Cart
	var cartLines []models.CartLine

	err := r.db.Preload("CartLines").Where("id = ? AND customer_id = ? AND status = ?", cartID, customerID, "Active").First(&cart).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return cart, cartLines, true, nil
		}
		return models.Cart{}, nil, false, err
	}

	err = r.db.Where("cart_id = ?", cart.ID).Find(&cartLines).Error
	if err != nil {
		return models.Cart{}, nil, false, err
	}

	return cart, cartLines, false, nil
}

func (r *cartRepo) GetCartWithLinesByIDAndGuestID(cartID int, guestID string) (models.Cart, []models.CartLine, bool, error) {
	var cart models.Cart
	var cartLines []models.CartLine

	err := r.db.Preload("CartLines").Where("id = ? AND guest_id = ? AND status = ?", cartID, guestID, "Active").First(&cart).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return cart, cartLines, true, nil
		}
		return models.Cart{}, nil, false, err
	}

	err = r.db.Where("cart_id = ?", cart.ID).Find(&cartLines).Error
	if err != nil {
		return models.Cart{}, nil, false, err
	}

	return cart, cartLines, false, nil
}

func (r *cartRepo) CreateCart(cart models.Cart) (models.Cart, error) {
	err := r.db.Create(&cart).Error
	if err != nil {
		return models.Cart{}, err
	}
	return cart, nil
}

func (r *cartRepo) SaveCart(cart models.Cart) (models.Cart, error) {
	err := r.db.Save(&cart).Error
	if err != nil {
		return models.Cart{}, err
	}
	return cart, nil
}

func (r *cartRepo) AddCartLine(cartLine models.CartLine) (models.CartLine, error) {
	err := r.db.Create(&cartLine).Error
	if err != nil {
		return models.CartLine{}, err
	}
	return cartLine, nil
}

func (r *cartRepo) SaveCartLine(cartLine models.CartLine) (models.CartLine, error) {
	err := r.db.Save(&cartLine).Error
	if err != nil {
		return models.CartLine{}, err
	}
	return cartLine, nil
}

func (r *cartRepo) DeleteCartWithLines(id int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cart_id = ?", id).Delete(&models.CartLine{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&models.Cart{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *cartRepo) GetCartLineWithValidation(customerID, cartID, lineID int) (*models.CartLine, error) {
	var cart models.Cart
	var cartLine models.CartLine

	if err := r.db.Where("id = ?", cartID).First(&cart).Error; err != nil {
		return nil, err
	}
	if cart.CustomerID != customerID || cart.Status != "Active" {
		return nil, errors.New("invalid cart for customer or inactive status")
	}

	if err := r.db.Where("id = ? AND cart_id = ?", lineID, cartID).First(&cartLine).Error; err != nil {
		return nil, err
	}

	return &cartLine, nil
}

func (r *cartRepo) MostRecentAllowedCart(customerID int) (*models.Cart, error) {
	var cart models.Cart
	cutoffTime := time.Now().Add(-120 * time.Hour)

	err := r.db.Where("customer_id = ? AND status = ? AND ever_checked_out = ? AND date_modified >= ?",
		customerID, "Active", false, cutoffTime).
		Order("last_retrieved DESC").
		First(&cart).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepo) TotalQuantity(cartID int) (int, error) {
	var total int
	err := r.db.Model(&models.CartLine{}).
		Where("cart_id = ?", cartID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error

	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *cartRepo) ArchiveCart(id int) error {
	return r.db.Model(&models.Cart{}).Where("id = ?", id).Update("status", "Archived").Error
}

func (r *cartRepo) ActiveCart(id int) error {
	return r.db.Model(&models.Cart{}).Where("id = ?", id).Update("status", "Active").Error
}

func (r *cartRepo) ReactivateCartWithLines(cartID int, newLines []models.CartLine) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Cart{}).Where("id = ?", cartID).
			Updates(map[string]interface{}{
				"status":           "Active",
				"ever_checked_out": false,
				"date_modified":    time.Now(),
			}).Error; err != nil {
			return err
		}
		if err := tx.Where("cart_id = ?", cartID).Delete(&models.CartLine{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&newLines).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *cartRepo) CopyCartWithLines(cartID int, newCustomer int) error {
	cart, err := r.ReadWithPreload(cartID)
	if err != nil {
		return err
	}

	if cart.CustomerID == newCustomer && cart.Status == "Active" {
		return nil
	} else if cart.CustomerID == newCustomer {
		return r.ActiveCart(cartID)
	}

	newCart := models.Cart{
		CustomerID:    newCustomer,
		DateCreated:   time.Now(),
		DateModified:  time.Now(),
		LastRetrieved: time.Now(),
		Status:        "Active",
	}

	cartLines, err := r.CartLinesRetrieval(cartID)
	if err != nil {
		return err
	}

	newCartLines := make([]models.CartLine, len(cartLines))
	for i, line := range cartLines {
		newCartLines[i] = models.CartLine{
			VariantID:       line.VariantID,
			ProductID:       line.ProductID,
			Quantity:        line.Quantity,
			NonDiscPrice:    line.NonDiscPrice,
			Price:           line.Price,
			IsGiftCard:      line.IsGiftCard,
			GiftCardCode:    line.GiftCardCode,
			GiftCardMessage: line.GiftCardMessage,
		}
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newCart).Error; err != nil {
			return err
		}
		for i := range newCartLines {
			newCartLines[i].CartID = newCart.ID
		}
		return tx.Create(&newCartLines).Error
	})
}

func (r *cartRepo) MoveCart(cartID int, newCustomer int) error {
	cart, err := r.Read(cartID)
	if err != nil {
		return err
	}

	if cart.CustomerID == newCustomer && cart.Status == "Active" {
		return nil
	} else if cart.CustomerID == newCustomer {
		return r.ActiveCart(cartID)
	}

	cart.GuestID = ""
	cart.CustomerID = newCustomer
	cart.Status = "Active"
	cart.EverCheckedOut = false
	cart.DateCreated = time.Now()
	cart.LastRetrieved = time.Now()
	cart.DateModified = time.Now()

	return r.Update(*cart)
}

// Cart ID confirm, actual error, redirect to signin
func (r *cartRepo) DirectCartRetrieval(cartID, customerID int, guestID string) (int, error, bool) {
	cart, err := r.Read(cartID)
	if err != nil {
		return cartID, err, false
	} else if cart.Status != "Active" {
		return cartID, errors.New("inactive cart"), false
	} else if customerID == 0 || customerID != cart.CustomerID {
		return cartID, nil, true
	}
	return cartID, nil, false
}
