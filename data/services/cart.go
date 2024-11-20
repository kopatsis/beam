package services

import (
	"beam/data/models"
	"beam/data/repositories"
	"errors"
	"strconv"
)

type CartService interface {
	AddCart(cart models.Cart) error
	GetCartByID(id int) (*models.Cart, error)
	UpdateCart(cart models.Cart) error
	DeleteCart(id int) error
	AddToCart(id, handle, name string, quant int, prodServ *productService, custID int, guestID string) (*models.Cart, error)
}

type cartService struct {
	cartRepo repositories.CartRepository
}

func NewCartService(cartRepo repositories.CartRepository) CartService {
	return &cartService{cartRepo: cartRepo}
}

func (s *cartService) AddCart(cart models.Cart) error {
	return s.cartRepo.Create(cart)
}

func (s *cartService) GetCartByID(id int) (*models.Cart, error) {
	return s.cartRepo.Read(id)
}

func (s *cartService) UpdateCart(cart models.Cart) error {
	return s.cartRepo.Update(cart)
}

func (s *cartService) DeleteCart(id int) error {
	return s.cartRepo.Delete(id)
}

func (s *cartService) AddToCart(id, handle, name string, quant int, prodServ *productService, custID int, guestID string) (*models.Cart, error) {
	p, r, err := prodServ.productRepo.GetFullProduct(name, handle)
	if err != nil {
		return nil, err
	} else if r != "" {
		return nil, errors.New("product has a redirection")
	}

	vid, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	index := -1
	for i, v := range p.Variants {
		if v.PK == vid {
			index = i
		}
	}

	if index < 0 {
		return nil, errors.New("no matching variant by id to provided handle")
	}

	cart, lines, exists, err := models.Cart{}, []models.CartLine{}, false, nil
	if custID > 0 {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByCustomerID(custID)
	} else if guestID != "" {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByGuestID(guestID)
	} else {
		return nil, errors.New("error retrieving cart unrelated to cart not existing")
	}
	if err != nil {
		return nil, err
	}

	if !exists {
		cart.Status = "Active"
		if custID > 0 {
			cart.CustomerID = custID
		} else {
			cart.GuestID = guestID
		}
	}

	var line models.CartLine
	for _, l := range lines {
		if l.ProductID == p.PK && l.VariantID == vid {
			line = l
		}
	}

	if line.ID == 0 {
		line = models.CartLine{
			ProductID:     p.PK,
			VariantID:     vid,
			ImageURL:      p.ImageURL,
			ProductTitle:  p.Title,
			Variant1Key:   p.Var1Key,
			Variant1Value: p.Variants[index].Var1Value,
			Price:         p.Variants[index].Price,
		}
		if p.Var2Key != "" {
			line.Variant2Key = &p.Var2Key
			line.Variant2Value = &p.Variants[index].Var2Value
		}
		if p.Var3Key != "" {
			line.Variant3Key = &p.Var3Key
			line.Variant3Value = &p.Variants[index].Var3Value
		}
	}

	cart.ItemCount += quant
	line.Quantity += quant

	if exists {
		cart, err = s.cartRepo.SaveCart(cart)
	} else {
		cart, err = s.cartRepo.CreateCart(cart)
	}
	if err != nil {
		return nil, err
	}

	line.CartID = cart.ID

	if line.ID == 0 {
		line, err = s.cartRepo.AddCartLine(line)
	} else {
		line, err = s.cartRepo.SaveCartLine(line)
	}
	if err != nil {
		return nil, err
	}

	return &cart, nil
}
