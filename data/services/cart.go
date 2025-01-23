package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/carthelp"
	"beam/data/services/product"
	"errors"
	"fmt"
	"strconv"
)

type CartService interface {
	AddCart(cart models.Cart) error
	GetCartByID(id int) (*models.Cart, error)
	UpdateCart(cart models.Cart) error
	DeleteCart(id int) error
	AddToCart(id, handle, name string, quant int, prodServ *productService, custID int, guestID string, logger eventService) (*models.Cart, error)
	GetCart(name string, custID int, guestID string, prodServ *productService) (*models.CartRender, error)
	AdjustQuantity(id, cartID, lineID string, quant int, prodServ *productService, custID int, guestID string, logger eventService) (*models.CartRender, error)
	ClearCart(name string, custID int, guestID string, logger eventService) (*models.CartRender, error)
	AddGiftCard(message, store string, cents int, discService *discountService, tools *config.Tools, custID int, guestID string, logger eventService) (*models.Cart, error)
	DeleteGiftCard(name, cartID, lineID string, custID int, guestID string, prodServ *productService, logger eventService) (*models.CartRender, error)
	UpdateRender(name string, cart *models.CartRender, ps *productService) error
	SavesListToCart(id, handle, name string, ps *productService, ls *listService, custID int, logger eventService) (models.SavesListRender, *models.CartRender, error)
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

func (s *cartService) AddToCart(id, handle, name string, quant int, prodServ *productService, custID int, guestID string, logger eventService) (*models.Cart, error) {
	p, r, err := prodServ.productRepo.GetFullProduct(name, handle)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Error querying product", "", "", "", strconv.Itoa(p.PK), "", "", "", "", []error{err})
		return nil, err
	} else if r != "" {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Queried product, but has redirection", "", "", "", strconv.Itoa(p.PK), "", "", "", "", []error{errors.New("product has a redirection")})
		return nil, errors.New("product has a redirection")
	}

	vid, err := strconv.Atoi(id)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Variant ID is not an int", "", "", "", strconv.Itoa(p.PK), "", "", "", "", []error{err})
		return nil, err
	}

	index := -1
	for i, v := range p.Variants {
		if v.PK == vid {
			index = i
		}
	}

	if index < 0 {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "No matching variant ID as provided", "", "", "", strconv.Itoa(p.PK), "", "", "", "", []error{errors.New("no matching variant by id to provided handle")})
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
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "No customer id or guest ID", "", "", "", strconv.Itoa(p.PK), "", "", "", "", []error{errors.New("error retrieving cart unrelated to cart not existing")})
		return nil, errors.New("error retrieving cart unrelated to cart not existing")
	}
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Unable to retrieve cart and lines", "", "", "", strconv.Itoa(p.PK), "", "", "", "", []error{err})
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
			ProductID:    p.PK,
			VariantID:    vid,
			NonDiscPrice: p.Variants[index].Price,
		}
	}

	line.Quantity += quant
	line.Price = product.VolumeDiscPrice(p.Variants[index].Price, quant, p.VolumeDisc)

	if !exists {
		cart, err = s.cartRepo.CreateCart(cart)
	}
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Unable to create cart for inexistent", "", "", "", strconv.Itoa(p.PK), "", "", "", "", []error{err})
		return nil, err
	}

	line.CartID = cart.ID

	if line.ID == 0 {
		line, err = s.cartRepo.AddCartLine(line)
	} else {
		line, err = s.cartRepo.SaveCartLine(line)
	}
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Unable to save or add line for cart", "", "", "", strconv.Itoa(p.PK), "", strconv.Itoa(cart.ID), "", "", []error{err})
		return nil, err
	}

	logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Success", "", "", "", strconv.Itoa(p.PK), "", strconv.Itoa(cart.ID), "", "", nil)
	return &cart, nil
}

func (s *cartService) GetCart(name string, custID int, guestID string, prodServ *productService) (*models.CartRender, error) {
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
	if err := s.UpdateRender(name, &ret, prodServ); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *cartService) AdjustQuantity(name, cartID, lineID string, quant int, prodServ *productService, custID int, guestID string, logger eventService) (*models.CartRender, error) {
	ret := models.CartRender{}

	cid, err := strconv.Atoi(cartID)
	if err != nil {
		// 1
		return nil, err
	}

	lid, err := strconv.Atoi(lineID)
	if err != nil {
		// 2
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
		// 3
		return nil, errors.New("no user id of either type provided")
	}
	if err != nil {
		// 4
		return nil, err
	}

	if !exists {
		ret.Empty = true
		ret.CartError = "Cart has been checked out already or no longer exists"
		// 5
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
		if err := s.UpdateRender(name, &ret, prodServ); err != nil {
			// 6
			return nil, err
		}
		// 7
		return &ret, nil
	}

	prod, redir, err := prodServ.GetProductByVariantID(name, lines[index].VariantID)
	if err != nil {
		// 8
		return nil, err
	}

	if redir != "" {
		err := s.cartRepo.DeleteCartLine(lines[index])
		if err != nil {
			// 9
			return nil, err
		}
		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		if err := s.UpdateRender(name, &ret, prodServ); err != nil {
			// 10
			return nil, err
		}
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
			// 11
			return nil, err
		}
		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		if err := s.UpdateRender(name, &ret, prodServ); err != nil {
			// 12
			return nil, err
		}
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
		if err := s.UpdateRender(name, &ret, prodServ); err != nil {
			logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to update the render after save", "", "", "", strconv.Itoa(lines[index].ProductID), "", cartID, "", "", []error{err})
			return nil, err
		}
		ret.CartError = "Unable to update cart :/ Please refresh and try again"
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to update the cart after save", "", "", "", strconv.Itoa(lines[index].ProductID), "", cartID, "", "", []error{err})
		return &ret, nil
	}

	if err := s.UpdateRender(name, &ret, prodServ); err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to update the render", "", "", "", strconv.Itoa(lines[index].ProductID), "", cartID, "", "", []error{err})
		return nil, err
	}

	logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Success", "", "", "", strconv.Itoa(lines[index].ProductID), "", cartID, "", "", nil)
	return &ret, nil
}

func (s *cartService) ClearCart(name string, custID int, guestID string, logger eventService) (*models.CartRender, error) {
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

	logger.SaveEvent(custID, guestID, "Cart", "ClearCart", "Success", "", "", "", "", "", strconv.Itoa(cart.ID), "", "", nil)
	return &ret, nil
}

func (s *cartService) AddGiftCard(message, store string, cents int, discService *discountService, tools *config.Tools, custID int, guestID string, logger eventService) (*models.Cart, error) {
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

	idDB, gccode, err := discService.CreateGiftCard(cents, message, store, tools)

	if line.ID == 0 {
		line = models.CartLine{
			IsGiftCard:      true,
			ProductID:       -1,
			VariantID:       idDB,
			GiftCardCode:    gccode,
			GiftCardMessage: message,
			Price:           cents,
			Quantity:        1,
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

	logger.SaveEvent(custID, guestID, "Cart", "AddGiftCard", "Success", "", "", "", "", "", strconv.Itoa(cart.ID), "", strconv.Itoa(idDB), nil)
	return &cart, nil
}

func (s *cartService) DeleteGiftCard(name, cartID, lineID string, custID int, guestID string, prodServ *productService, logger eventService) (*models.CartRender, error) {
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
		if err := s.UpdateRender(name, &ret, prodServ); err != nil {
			return nil, err
		}
		return &ret, nil
	}

	if err := s.cartRepo.DeleteCartLine(lines[index]); err != nil {
		return nil, err
	}
	ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)

	if err := s.UpdateRender(name, &ret, prodServ); err != nil {
		return nil, err
	}

	logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Success", "", "", "", "", "", strconv.Itoa(cart.ID), "", strconv.Itoa(lines[index].VariantID), nil)
	return &ret, nil
}

func (s *cartService) SavesListToCart(id, handle, name string, ps *productService, ls *listService, custID int, logger eventService) (models.SavesListRender, *models.CartRender, error) {
	vid, err := strconv.Atoi(id)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	sl, err := ls.DeleteSavesListRender(name, custID, vid, 1, ps)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	_, err = s.AddToCart(id, handle, name, 1, ps, custID, "", logger)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	cr, err := s.GetCart(name, custID, "", ps)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	return sl, cr, nil
}

func (s *cartService) UpdateRender(name string, cart *models.CartRender, ps *productService) error {
	carthelp.UpdateCartSub(cart)
	vids := []int{}

	for _, cl := range cart.CartLines {
		if !cl.ActualLine.IsGiftCard {
			vids = append(vids, cl.ActualLine.VariantID)
		}
	}

	lvs, err := ps.GetLimitedVariants(name, vids)
	if err != nil {
		return err
	} else if len(lvs) != len(vids) {
		return errors.New("variants returned incorrect length equal to variant IDs")
	}

	for i, cl := range cart.CartLines {
		if !cl.ActualLine.IsGiftCard {
			found := false
			for _, lv := range lvs {
				if lv.VariantID == cl.ActualLine.VariantID {
					cl.Variant = *lv
					found = true
					cart.CartLines[i] = cl
					break
				}
			}
			if !found {
				return fmt.Errorf("could not find single lim var for id: %d", cl.ActualLine.VariantID)
			}
		}
	}

	return nil
}
