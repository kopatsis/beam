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
	"time"

	"gorm.io/gorm"
)

type CartService interface {
	AddCart(cart models.Cart) error
	GetCartByID(id int) (*models.Cart, error)
	UpdateCart(cart models.Cart) error
	DeleteCart(id int) error
	AddToCart(id, handle, name string, quant int, prodServ *productService, custID int, guestID string, logger *eventService) (*models.Cart, error)
	GetCart(name string, custID int, guestID string, prodServ *productService) (*models.CartRender, error)
	AdjustQuantity(id, cartID, lineID string, quant int, prodServ *productService, custID int, guestID string, logger *eventService) (*models.CartRender, error)
	ClearCart(name string, custID int, guestID string, logger *eventService) (*models.CartRender, error)
	AddGiftCard(message, store string, cents int, discService *discountService, tools *config.Tools, custID int, guestID string, logger *eventService) (*models.Cart, error)
	DeleteGiftCard(name, cartID, lineID string, custID int, guestID string, prodServ *productService, logger *eventService) (*models.CartRender, error)
	UpdateRender(name string, cart *models.CartRender, ps *productService) error
	SavesListToCart(id, handle, name string, ps *productService, ls *listService, custID int, logger *eventService) (models.SavesListRender, *models.CartRender, error)

	CartMiddleware(cartID, custID int, guestID string) (int, error)
	GetCartMain(cartID, custID int, guestID string) (*models.Cart, error, bool)
	GetCartAndVerify(cartID, custID int, guestID string) (int, *models.Cart, error)
	GetCartMainWithLines(cartID, custID int, guestID string) (*models.Cart, []*models.CartLine, error, bool)
	GetCartWithLinesAndVerify(cartID, custID int, guestID string) (int, *models.Cart, []*models.CartLine, error)
	CartCountCheck(cartID, custID int, guestID string) (int, int, error)
	OrderSuccessCart(cartID, custID int, guestID string, orderLines []models.OrderLine) error

	CopyCartWithLines(cartID, newCustomer int) error
	MoveCart(cartID, newCustomer int) error
	DirectCartRetrieval(cartID, customerID int, guestID string) (int, error, bool)
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

func (s *cartService) AddToCart(id, handle, name string, quant int, prodServ *productService, custID int, guestID string, logger *eventService) (*models.Cart, error) {
	p, r, err := prodServ.productRepo.GetFullProduct(name, handle)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Error querying product", "", "", "", "", id, "", "", "", "", "", "", "", []error{err})
		return nil, err
	} else if r != "" {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Queried product, but has redirection", "", "", "", "", id, "", "", "", "", "", "", "", []error{errors.New("product has a redirection")})
		return nil, errors.New("product has a redirection")
	}

	vid, err := strconv.Atoi(id)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Variant ID is not an int", "", "", "", strconv.Itoa(p.PK), id, "", "", "", "", "", "", "", []error{err})
		return nil, err
	}

	index := -1
	for i, v := range p.Variants {
		if v.PK == vid {
			index = i
		}
	}

	if index < 0 {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "No matching variant ID as provided", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", "", "", "", "", []error{errors.New("no matching variant by id to provided handle")})
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
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "No customer id or guest ID", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", "", "", "", "", []error{errors.New("error retrieving cart unrelated to cart not existing")})
		return nil, errors.New("error retrieving cart unrelated to cart not existing")
	}
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Unable to retrieve cart and lines", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", "", "", "", "", []error{err})
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
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Unable to create cart for inexistent", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", "", "", "", "", []error{err})
		return nil, err
	}

	line.CartID = cart.ID

	if line.ID == 0 {
		line, err = s.cartRepo.AddCartLine(line)
	} else {
		line, err = s.cartRepo.SaveCartLine(line)
	}
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Unable to save or add line for cart", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", strconv.Itoa(cart.ID), "", "", "", []error{err})
		return nil, err
	}

	logger.SaveEvent(custID, guestID, "Cart", "AddToCart", "Success", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", strconv.Itoa(cart.ID), "", "", "", nil)
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

func (s *cartService) AdjustQuantity(name, cartID, lineID string, quant int, prodServ *productService, custID int, guestID string, logger *eventService) (*models.CartRender, error) {
	ret := models.CartRender{}

	cid, err := strconv.Atoi(cartID)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to convert cart ID to int", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{err})
		return nil, err
	}

	lid, err := strconv.Atoi(lineID)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to convert line ID to int", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{err})
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
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "No user ID or guest ID valid", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{errors.New("no user id of either type provided")})
		return nil, errors.New("no user id of either type provided")
	}
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to retrieve cart and lines", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{err})
		return nil, err
	}

	if !exists {
		ret.Empty = true
		ret.CartError = "Cart has been checked out already or no longer exists"
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Cart is empty", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{err})
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
			logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Line by given id was deleted, couldn't update render", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{err})
			return nil, err
		}
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to convert cart ID to int", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{errors.New("index == -1")})
		return &ret, nil
	}

	prod, redir, err := prodServ.GetProductByVariantID(name, lines[index].VariantID)
	if err != nil {

		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to retrieve product by variant ID", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", []error{err})
		return nil, err
	}

	if redir != "" {
		err := s.cartRepo.DeleteCartLine(lines[index])
		if err != nil {
			logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to delete cart line after product was redirected", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", []error{err})
			return nil, err
		}
		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		if err := s.UpdateRender(name, &ret, prodServ); err != nil {
			logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to update render after product was redirected", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", []error{err})
			return nil, err
		}
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Deleted line and updated render for redirected product", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", nil)
		return &ret, nil
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
			logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Variant not in product specified by line; couldn't delete line", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", []error{err})
			return nil, err
		}
		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		if err := s.UpdateRender(name, &ret, prodServ); err != nil {
			logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Variant not in product specified by line; couldn't update render", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", []error{err})
			return nil, err
		}
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Variant not in product specified by line", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", nil)
		return &ret, nil
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
			logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to update the render after save", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", []error{err})
			return nil, err
		}
		ret.CartError = "Unable to update cart :/ Please refresh and try again"
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to update the cart after save", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", []error{err})
		return &ret, nil
	}

	if err := s.UpdateRender(name, &ret, prodServ); err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Unable to update the render", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", []error{err})
		return nil, err
	}

	logger.SaveEvent(custID, guestID, "Cart", "AdjustQuantity", "Success", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", cartID, lineID, "", "", nil)
	return &ret, nil
}

func (s *cartService) ClearCart(name string, custID int, guestID string, logger *eventService) (*models.CartRender, error) {
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
		logger.SaveEvent(custID, guestID, "Cart", "ClearCart", "No user ID or guest ID valid", "", "", "", "", "", "", "", "", "", "", "", "", []error{errors.New("no user id of either type provided")})
		return nil, errors.New("no user id of either type provided")
	}

	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "ClearCart", "Unable to query cart + lines", "", "", "", "", "", "", "", "", "", "", "", "", []error{err})
		return nil, err
	}

	if !exists {
		ret.Empty = true
		logger.SaveEvent(custID, guestID, "Cart", "ClearCart", "Cart does not exist to clear", "", "", "", "", "", "", "", "", "", "", "", "", nil)
		return &ret, nil
	}

	if err := s.cartRepo.DeleteCartWithLines(cart.ID); err != nil {
		for _, l := range lines {
			ret.CartLines = append(ret.CartLines, models.CartLineRender{ActualLine: l})
		}

		ret.Cart = cart
		carthelp.UpdateCartSub(&ret)

		logger.SaveEvent(custID, guestID, "Cart", "ClearCart", "Could not clear cart", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), "", "", "", []error{errors.New("failed to clear cart")})
		return &ret, errors.New("failed to clear cart")
	}

	ret.Empty = true

	logger.SaveEvent(custID, guestID, "Cart", "ClearCart", "Success", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), "", "", "", nil)
	return &ret, nil
}

func (s *cartService) AddGiftCard(message, store string, cents int, discService *discountService, tools *config.Tools, custID int, guestID string, logger *eventService) (*models.Cart, error) {
	var err error
	var cart models.Cart
	var exists bool

	if custID > 0 {
		cart, _, exists, err = s.cartRepo.GetCartWithLinesByCustomerID(custID)
	} else if guestID != "" {
		cart, _, exists, err = s.cartRepo.GetCartWithLinesByGuestID(guestID)
	} else {
		logger.SaveEvent(custID, guestID, "Cart", "AddGiftCard", "No user ID or guest ID valid", "", "", "", "", "", "", "", "", "", "", "", "", []error{errors.New("error retrieving cart unrelated to cart not existing")})
		return nil, errors.New("error retrieving cart unrelated to cart not existing")
	}
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddGiftCard", "Unable to query cart + lines", "", "", "", "", "", "", "", "", "", "", "", "", []error{err})
		return nil, err
	}

	if !exists {
		cart.Status = "Active"
		if custID > 0 {
			cart.CustomerID = custID
		} else {
			cart.GuestID = guestID
		}
		cart, err = s.cartRepo.CreateCart(cart)
		if err != nil {
			logger.SaveEvent(custID, guestID, "Cart", "AddGiftCard", "Unable to create cart previously empty", "", "", "", "", "", "", "", "", "", "", "", "", []error{err})
			return nil, err
		}
	}

	idDB, gccode, err := discService.CreateGiftCard(cents, message, store, tools)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddGiftCard", "Unable to create gift card to add to cart", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), "", "", "", []error{err})
		return nil, err
	}

	line := models.CartLine{
		CartID:          cart.ID,
		IsGiftCard:      true,
		ProductID:       -1,
		VariantID:       idDB,
		GiftCardCode:    gccode,
		GiftCardMessage: message,
		Price:           cents,
		Quantity:        1,
	}

	line, err = s.cartRepo.AddCartLine(line)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "AddGiftCard", "Unable to save line for created gift card", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), "", "", strconv.Itoa(idDB), []error{err})
		return nil, err
	}

	logger.SaveEvent(custID, guestID, "Cart", "AddGiftCard", "Success", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), strconv.Itoa(line.ID), "", strconv.Itoa(idDB), nil)
	return &cart, nil
}

func (s *cartService) DeleteGiftCard(name, cartID, lineID string, custID int, guestID string, prodServ *productService, logger *eventService) (*models.CartRender, error) {
	ret := models.CartRender{}

	cid, err := strconv.Atoi(cartID)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Cart ID given not int format", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{err})
		return nil, err
	}

	lid, err := strconv.Atoi(lineID)
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Cart Line ID given not int format", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{err})
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
		logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "No user ID or guest ID valid", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{errors.New("no user id of either type provided")})
		return nil, errors.New("no user id of either type provided")
	}
	if err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Unable to query cart + lines", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{err})
		return nil, err
	}

	if !exists {
		ret.Empty = true
		ret.CartError = "Cart has been checked out already or no longer exists"
		logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Cart has been checked out already or no longer exists", "", "", "", "", "", "", "", "", cartID, lineID, "", "", nil)
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
			logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Line for gift card was deleted -> Unable to update render", "", "", "", "", "", "", "", "", cartID, lineID, "", "", []error{err})
			return nil, err
		}
		logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Line for gift card was deleted", "", "", "", "", "", "", "", "", cartID, lineID, "", "", nil)
		return &ret, nil
	}

	if err := s.cartRepo.DeleteCartLine(lines[index]); err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Unable to delete cart line for gift card", "", "", "", "", "", "", "", "", cartID, lineID, "", strconv.Itoa(lines[index].VariantID), []error{err})
		return nil, err
	}
	ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)

	if err := s.UpdateRender(name, &ret, prodServ); err != nil {
		logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Unable to update render after deleting line", "", "", "", "", "", "", "", "", cartID, lineID, "", strconv.Itoa(lines[index].VariantID), []error{err})
		return nil, err
	}

	logger.SaveEvent(custID, guestID, "Cart", "DeleteGiftCard", "Success", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), strconv.Itoa(lines[index].ID), "", strconv.Itoa(lines[index].VariantID), nil)
	return &ret, nil
}

func (s *cartService) SavesListToCart(id, handle, name string, ps *productService, ls *listService, custID int, logger *eventService) (models.SavesListRender, *models.CartRender, error) {
	vid, err := strconv.Atoi(id)
	if err != nil {
		logger.SaveEvent(custID, "", "Cart", "SavesListToCart", "Unable to convert saves list var id to int", "", "", "", "", id, "", "", "", "", "", "", "", []error{err})
		return models.SavesListRender{}, nil, err
	}

	sl, err := ls.DeleteSavesListRender(name, custID, vid, 1, ps)
	if err != nil {
		logger.SaveEvent(custID, "", "Cart", "SavesListToCart", "Unable to delete off of saves list", "", "", "", "", id, "", "", "", "", "", "", "", []error{err})
		return models.SavesListRender{}, nil, err
	}

	_, err = s.AddToCart(id, handle, name, 1, ps, custID, "", logger)
	if err != nil {
		logger.SaveEvent(custID, "", "Cart", "SavesListToCart", "Unable to add to cart after delete off of saves list", "", "", "", "", id, "", "", "", "", "", "", "", []error{err})
		return models.SavesListRender{}, nil, err
	}

	cr, err := s.GetCart(name, custID, "", ps)
	if err != nil {
		logger.SaveEvent(custID, "", "Cart", "SavesListToCart", "Unable to add to retrieve cart after adding from saves list", "", "", "", "", id, "", "", "", "", "", "", "", []error{err})
		return models.SavesListRender{}, nil, err
	}

	logger.SaveEvent(custID, "", "Cart", "SavesListToCart", "Success", "", "", "", "", id, "", "", "", strconv.Itoa(cr.Cart.ID), "", "", "", nil)
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

func (s *cartService) CartMiddleware(cartID, custID int, guestID string) (int, error) {
	if custID > 0 {
		if cartID <= 0 {
			exCart, err := s.cartRepo.MostRecentAllowedCart(custID)
			if exCart != nil && err == nil {
				return exCart.ID, nil
			}
			cart := models.Cart{
				CustomerID:    custID,
				DateCreated:   time.Now(),
				DateModified:  time.Now(),
				LastRetrieved: time.Now(),
				Status:        "Active",
			}
			cart, err = s.cartRepo.SaveCart(cart)
			if err != nil {
				return 0, err
			}
			return cart.ID, nil
		}
		return cartID, nil
	} else if guestID != "" {
		if cartID < 0 {
			cart := models.Cart{
				GuestID:       guestID,
				DateCreated:   time.Now(),
				DateModified:  time.Now(),
				LastRetrieved: time.Now(),
				Status:        "Active",
			}
			cart, err := s.cartRepo.SaveCart(cart)
			if err != nil {
				return 0, err
			}
			return cart.ID, nil
		}
		return cartID, nil
	}
	return 0, errors.New("no one logged in")
}

func (s *cartService) GetCartMain(cartID, custID int, guestID string) (*models.Cart, error, bool) {
	cart, err := s.cartRepo.Read(cartID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err, true
		}
		return nil, err, false
	}

	if cart.Status != "Active" {
		return nil, errors.New("inactive cart"), true
	} else if cart.ID != cartID {
		return nil, errors.New("queried the incorrect cart by id"), true
	}

	if custID > 0 {
		if cart.CustomerID != custID {
			return nil, errors.New("customer cart doesn't belong to customer"), true
		}
	} else if guestID != "" {
		if cart.CustomerID != custID {
			return nil, errors.New("guest cart doesn't belong to guest"), true
		}
	} else {
		return nil, errors.New("no one logged in"), true
	}

	return cart, nil, false
}

func (s *cartService) GetCartAndVerify(cartID, custID int, guestID string) (int, *models.Cart, error) {
	cart, err, retry := s.GetCartMain(cartID, custID, guestID)
	if err != nil {
		if retry {
			newID, err := s.CartMiddleware(cartID, custID, guestID)
			if err != nil {
				return cartID, nil, err
			}
			cart, err, _ = s.GetCartMain(cartID, custID, guestID)
			return newID, cart, err
		}
		return cartID, nil, err
	}
	return cart.ID, cart, nil
}

func (s *cartService) GetCartMainWithLines(cartID, custID int, guestID string) (*models.Cart, []*models.CartLine, error, bool) {

	cart, err := s.cartRepo.ReadWithPreload(cartID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, err, true
		}
		return nil, nil, err, false
	} else if cart.Status != "Active" {
		return nil, nil, errors.New("inactive cart"), true
	} else if cart.ID != cartID {
		return nil, nil, errors.New("queried the incorrect cart by id"), true
	}

	if custID > 0 {
		if cart.CustomerID != custID {
			return nil, nil, errors.New("customer cart doesn't belong to customer"), true
		}
	} else if guestID != "" {
		if cart.CustomerID != custID {
			return nil, nil, errors.New("guest cart doesn't belong to guest"), true
		}
	} else {
		return nil, nil, errors.New("no one logged in"), true
	}

	cartLines, err := s.cartRepo.CartLinesRetrieval(cart.ID)
	if err != nil {
		return nil, nil, err, false
	}

	return cart, cartLines, nil, false
}

func (s *cartService) GetCartWithLinesAndVerify(cartID, custID int, guestID string) (int, *models.Cart, []*models.CartLine, error) {
	cart, lines, err, retry := s.GetCartMainWithLines(cartID, custID, guestID)
	if err != nil {
		if retry {
			newID, err := s.CartMiddleware(cartID, custID, guestID)
			if err != nil {
				return cartID, nil, nil, err
			}
			cart, lines, err, _ = s.GetCartMainWithLines(cartID, custID, guestID)
			return newID, cart, lines, err
		}
		return cartID, nil, nil, err
	}
	return cart.ID, cart, lines, nil
}

// Cart ID, count, err
func (s *cartService) CartCountCheck(cartID, custID int, guestID string) (int, int, error) {
	id, cart, err := s.GetCartAndVerify(cartID, custID, guestID)
	if err != nil {
		return cartID, 0, err
	}

	count, err := s.cartRepo.TotalQuantity(cart.ID)
	return id, count, err
}

func (s *cartService) OrderSuccessCart(cartID, custID int, guestID string, orderLines []models.OrderLine) error {
	var err error
	var cart models.Cart
	var lines []models.CartLine
	var exists bool

	if custID > 0 {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByIDAndCustomerID(cartID, custID)
	} else if guestID != "" {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByIDAndGuestID(cartID, guestID)
	} else {
		return errors.New("no user id of either type provided")
	}

	if !exists {
		return nil
	} else if err != nil {
		return err
	}

	varIDs := map[int]struct{}{}
	for _, l := range orderLines {
		if l.IsGiftCard {
			continue
		}
		varID, err := strconv.Atoi(l.VariantID)
		if err != nil {
			return err
		}
		varIDs[varID] = struct{}{}
	}

	newLines := []models.CartLine{}
	for _, line := range lines {
		if line.IsGiftCard {
			continue
		} else if _, ok := varIDs[line.VariantID]; ok {
			continue
		}
		newLines = append(newLines, line)
	}

	if len(newLines) == 0 {
		return s.cartRepo.ArchiveCart(cart.ID)
	}
	return s.cartRepo.ReactivateCartWithLines(cart.ID, newLines)
}

func (s *cartService) CopyCartWithLines(cartID, newCustomer int) error {
	return s.cartRepo.CopyCartWithLines(cartID, newCustomer)
}
func (s *cartService) MoveCart(cartID, newCustomer int) error {
	return s.cartRepo.MoveCart(cartID, newCustomer)
}

func (s *cartService) DirectCartRetrieval(cartID, customerID int, guestID string) (int, error, bool) {
	return s.cartRepo.DirectCartRetrieval(cartID, customerID, guestID)
}
