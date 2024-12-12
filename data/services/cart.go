package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/carthelp"
	"beam/data/services/product"
	"errors"
	"os"
	"strconv"
)

type CartService interface {
	AddCart(cart models.Cart) error
	GetCartByID(id int) (*models.Cart, error)
	UpdateCart(cart models.Cart) error
	DeleteCart(id int) error
	AddToCart(id, handle, name string, quant int, prodServ *productService, custID int, guestID string) (*models.Cart, error)
	GetCart(name string, custID int, guestID string) (*models.CartRender, error)
	AdjustQuantity(id, handle, name string, quant int, prodServ *productService, custID int, guestID string) (*models.CartRender, error)
	ClearCart(name string, custID int, guestID string) (*models.CartRender, error)
	AddGiftCard(message, store string, cents int, discService *discountService, tools *config.Tools, custID int, guestID string) (*models.Cart, error)
	DeleteGiftCard(cartID, lineID string, custID int, guestID string) (*models.CartRender, error)
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

	var cart models.Cart
	var lines []models.CartLine
	var exists bool

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
			ProductHandle: p.Handle,
			ProductID:     p.PK,
			VariantID:     vid,
			ImageURL:      p.ImageURL,
			ProductTitle:  p.Title,
			Variant1Key:   p.Var1Key,
			Variant1Value: p.Variants[index].Var1Value,
			NonDiscPrice:  p.Variants[index].Price,
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

	line.Quantity += quant
	line.Price = product.VolumeDiscPrice(p.Variants[index].Price, quant, p.VolumeDisc)

	if !exists {
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

func (s *cartService) GetCart(name string, custID int, guestID string) (*models.CartRender, error) {
	ret := models.CartRender{}

	var err error
	var cart models.Cart
	var lines []models.CartLine
	var exists bool

	if custID > 0 {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByCustomerID(custID)
	} else if guestID != "" {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByGuestID(guestID)
	} else {
		return nil, errors.New("no user id of either type provided")
	}

	if err != nil {
		return nil, err
	}

	if !exists {
		ret.Empty = true
		return &ret, nil
	}

	for _, l := range lines {
		ret.CartLines = append(ret.CartLines, models.CartLineRender{ActualLine: l})
	}

	ret.Cart = cart
	carthelp.UpdateCartSub(&ret)

	return &ret, nil
}

func (s *cartService) AdjustQuantity(name, cartID, lineID string, quant int, prodServ *productService, custID int, guestID string) (*models.CartRender, error) {
	ret := models.CartRender{}

	cid, err := strconv.Atoi(cartID)
	if err != nil {
		return nil, err
	}

	lid, err := strconv.Atoi(lineID)
	if err != nil {
		return nil, err
	}

	var cart models.Cart
	var lines []models.CartLine
	var exists bool

	if custID > 0 {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByIDAndCustomerID(cid, custID)
	} else if guestID != "" {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByIDAndGuestID(cid, guestID)
	} else {
		return nil, errors.New("no user id of either type provided")
	}
	if err != nil {
		return nil, err
	}

	if !exists {
		ret.Empty = true
		ret.CartError = "Cart has been checked out already or no longer exists"
		return &ret, nil
	}

	ret.Cart = cart

	index := -1
	for i, l := range lines {
		ret.CartLines = append(ret.CartLines, models.CartLineRender{
			ActualLine: l,
		})
		if l.ID == lid {
			index = i
		}
	}

	if index == -1 {
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		carthelp.UpdateCartSub(&ret)
		return &ret, nil
	}

	prod, redir, err := prodServ.productRepo.GetFullProduct(name, lines[index].ProductHandle)
	if err != nil {
		return nil, err
	}

	if redir != "" {
		err := s.cartRepo.DeleteCartLine(lines[index])
		if err != nil {
			return nil, err
		}
		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		carthelp.UpdateCartSub(&ret)
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
	}

	varIndex := -1
	for i, v := range prod.Variants {
		if v.PK == lines[index].VariantID {
			varIndex = i
		}
	}

	if varIndex == -1 {
		err := s.cartRepo.DeleteCartLine(lines[index])
		if err != nil {
			return nil, err
		}
		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		carthelp.UpdateCartSub(&ret)
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
	}

	newQuant := quant
	tooHigh := false
	if prod.Variants[varIndex].Quantity < newQuant {
		tooHigh = true
		newQuant = prod.Variants[varIndex].Quantity
	}

	oldQuant := ret.CartLines[index].ActualLine.Quantity
	ret.CartLines[index].ActualLine.Quantity = newQuant
	ret.CartLines[index].ActualLine.Price = product.VolumeDiscPrice(prod.Variants[varIndex].Price, newQuant, prod.VolumeDisc)
	ret.CartLines[index].QuantityMaxed = tooHigh

	_, err = s.cartRepo.SaveCartLine(ret.CartLines[index].ActualLine)
	if err != nil {
		ret.CartLines[index].ActualLine.Quantity = oldQuant
		ret.CartLines[index].ActualLine.Price = product.VolumeDiscPrice(prod.Variants[varIndex].Price, oldQuant, prod.VolumeDisc)
		carthelp.UpdateCartSub(&ret)
		ret.CartError = "Unable to update cart :/ Please refresh and try again"
		return &ret, nil
	}

	carthelp.UpdateCartSub(&ret)
	return &ret, nil
}

func (s *cartService) ClearCart(name string, custID int, guestID string) (*models.CartRender, error) {
	ret := models.CartRender{}

	var err error
	var cart models.Cart
	var lines []models.CartLine
	var exists bool

	if custID > 0 {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByCustomerID(custID)
	} else if guestID != "" {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByGuestID(guestID)
	} else {
		return nil, errors.New("no user id of either type provided")
	}

	if err != nil {
		return nil, err
	}

	if !exists {
		ret.Empty = true
		return &ret, nil
	}

	if err := s.cartRepo.DeleteCartWithLines(cart.ID); err != nil {
		for _, l := range lines {
			ret.CartLines = append(ret.CartLines, models.CartLineRender{ActualLine: l})
		}

		ret.Cart = cart
		carthelp.UpdateCartSub(&ret)

		return &ret, errors.New("failed to clear cart")
	}

	ret.Empty = true
	return &ret, nil
}

func (s *cartService) AddGiftCard(message, store string, cents int, discService *discountService, tools *config.Tools, custID int, guestID string) (*models.Cart, error) {
	var err error
	var cart models.Cart
	var exists bool

	if custID > 0 {
		cart, _, exists, err = s.cartRepo.GetCartWithLinesByCustomerID(custID)
	} else if guestID != "" {
		cart, _, exists, err = s.cartRepo.GetCartWithLinesByGuestID(guestID)
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

	idDB, _, err := discService.CreateGiftCard(cents, message, store, tools)

	gcHandle, gcImg := os.Getenv("GC_HANDLE"), os.Getenv("GC_IMG")
	if line.ID == 0 {
		line = models.CartLine{
			IsGiftCard:    true,
			ProductHandle: gcHandle,
			ProductID:     -1,
			VariantID:     idDB,
			ImageURL:      gcImg,
			ProductTitle:  "GIFT CARD",
			Variant1Key:   "Message",
			Variant1Value: message,
			Price:         cents,
			Quantity:      1,
		}
	}

	if !exists {
		cart, err = s.cartRepo.CreateCart(cart)
	}
	if err != nil {
		return nil, err
	}

	line.CartID = cart.ID

	line, err = s.cartRepo.AddCartLine(line)
	if err != nil {
		return nil, err
	}

	return &cart, nil
}

func (s *cartService) DeleteGiftCard(cartID, lineID string, custID int, guestID string) (*models.CartRender, error) {
	ret := models.CartRender{}

	cid, err := strconv.Atoi(cartID)
	if err != nil {
		return nil, err
	}

	lid, err := strconv.Atoi(lineID)
	if err != nil {
		return nil, err
	}

	var cart models.Cart
	var lines []models.CartLine
	var exists bool

	if custID > 0 {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByIDAndCustomerID(cid, custID)
	} else if guestID != "" {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByIDAndGuestID(cid, guestID)
	} else {
		return nil, errors.New("no user id of either type provided")
	}
	if err != nil {
		return nil, err
	}

	if !exists {
		ret.Empty = true
		ret.CartError = "Cart has been checked out already or no longer exists"
		return &ret, nil
	}

	ret.Cart = cart

	index := -1
	for i, l := range lines {
		ret.CartLines = append(ret.CartLines, models.CartLineRender{
			ActualLine: l,
		})
		if l.ID == lid {
			index = i
		}
	}

	if index == -1 {
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		carthelp.UpdateCartSub(&ret)
		return &ret, nil
	}

	if err := s.cartRepo.DeleteCartLine(lines[index]); err != nil {
		return nil, err
	}
	ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)

	carthelp.UpdateCartSub(&ret)
	return &ret, nil
}
