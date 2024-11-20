package repositories

import (
	"beam/data/models"

	"gorm.io/gorm"
)

type CartRepository interface {
	Create(cart models.Cart) error
	Read(id int) (*models.Cart, error)
	Update(cart models.Cart) error
	Delete(id int) error
	GetCartWithLinesByCustomerID(customerID int) (models.Cart, []models.CartLine, bool, error)
	GetCartWithLinesByGuestID(guestID string) (models.Cart, []models.CartLine, bool, error)
	CreateCart(cart models.Cart) (models.Cart, error)
	SaveCart(cart models.Cart) (models.Cart, error)
	AddCartLine(cartLine models.CartLine) (models.CartLine, error)
	SaveCartLine(cartLine models.CartLine) (models.CartLine, error)
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

func (r *cartRepo) Update(cart models.Cart) error {
	return r.db.Save(&cart).Error
}

func (r *cartRepo) Delete(id int) error {
	return r.db.Delete(&models.Cart{}, id).Error
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
