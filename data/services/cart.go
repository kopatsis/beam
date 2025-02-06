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
	AddToCart(dpi *DataPassIn, handle string, vid, quant int, prodServ *productService) (*models.Cart, error)
	GetCart(dpi *DataPassIn, prodServ *productService) (*models.CartRender, error)
	AdjustQuantity(dpi *DataPassIn, lineID, quant int, prodServ *productService) (*models.CartRender, error)
	ClearCart(dpi *DataPassIn) (*models.CartRender, error)
	AddGiftCard(dpi *DataPassIn, message string, cents int, discService *discountService, tools *config.Tools) (*models.Cart, error)
	DeleteGiftCard(dpi *DataPassIn, lineID int, prodServ *productService) (*models.CartRender, error)
	UpdateRender(name string, cart *models.CartRender, ps *productService) error
	SavesListToCart(dpi *DataPassIn, varid int, handle string, ps *productService, ls *listService) (models.SavesListRender, *models.CartRender, error)

	CartMiddleware(cartID, custID int, guestID string) (int, error)
	GetCartMain(dpi *DataPassIn) (*models.Cart, error, bool)
	GetCartAndVerify(dpi *DataPassIn) (int, *models.Cart, error)
	GetCartMainWithLines(dpi *DataPassIn) (*models.Cart, []*models.CartLine, error, bool)
	GetCartWithLinesAndVerify(dpi *DataPassIn) (int, *models.Cart, []*models.CartLine, error)
	CartCountCheck(dpi *DataPassIn) (int, int, error)
	OrderSuccessCart(dpi *DataPassIn, orderLines []models.OrderLine) error

	CopyCartWithLines(dpi *DataPassIn) error
	MoveCart(dpi *DataPassIn) error
	DirectCartRetrieval(dpi *DataPassIn) (int, error, bool)
}

type cartService struct {
	cartRepo repositories.CartRepository
}

func NewCartService(cartRepo repositories.CartRepository) CartService {
	return &cartService{cartRepo: cartRepo}
}

func (s *cartService) AddToCart(dpi *DataPassIn, handle string, vid, quant int, prodServ *productService) (*models.Cart, error) {
	p, r, err := prodServ.productRepo.GetFullProduct(dpi.Store, handle)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddToCart", "Error querying product", "", "", "", "", strconv.Itoa(vid), "", "", "", "", "", "", "", []error{err})
		return nil, err
	} else if r != "" {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddToCart", "Queried product, but has redirection", "", "", "", "", strconv.Itoa(vid), "", "", "", "", "", "", "", []error{errors.New("product has a redirection")})
		return nil, errors.New("product has a redirection")
	}

	index := -1
	for i, v := range p.Variants {
		if v.PK == vid {
			index = i
		}
	}

	if index < 0 {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddToCart", "No matching variant ID as provided", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", "", "", "", "", []error{errors.New("no matching variant by id to provided handle")})
		return nil, errors.New("no matching variant by id to provided handle")
	}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddToCart", "Unable to retrieve cart and lines", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", "", "", "", "", []error{err})
		return nil, err
	}
	dpi.CartID = id

	var line *models.CartLine
	for _, l := range lines {
		if l.ProductID == p.PK && l.VariantID == vid {
			line = l
		}
	}

	if line == nil {
		line = &models.CartLine{
			ProductID:    p.PK,
			VariantID:    vid,
			NonDiscPrice: p.Variants[index].Price,
		}
	}

	line.Quantity += quant
	line.Price = product.VolumeDiscPrice(p.Variants[index].Price, quant, p.VolumeDisc)
	line.CartID = cart.ID

	if err := s.cartRepo.SaveCartLineNew(line); err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddToCart", "Unable to save or add line for cart", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", strconv.Itoa(cart.ID), "", "", "", []error{err})
		return nil, err
	}

	dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddToCart", "Success", "", "", "", strconv.Itoa(p.PK), strconv.Itoa(vid), "", "", "", strconv.Itoa(cart.ID), "", "", "", nil)
	return cart, nil
}

func (s *cartService) GetCart(dpi *DataPassIn, prodServ *productService) (*models.CartRender, error) {
	ret := models.CartRender{}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		return nil, err
	}
	dpi.CartID = id

	if len(lines) == 0 {
		ret.Empty = true
		return &ret, nil
	}

	for _, l := range lines {
		ret.CartLines = append(ret.CartLines, models.CartLineRender{ActualLine: *l})
	}

	ret.Cart = *cart
	if err := s.UpdateRender(dpi.Store, &ret, prodServ); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *cartService) AdjustQuantity(dpi *DataPassIn, lineID, quant int, prodServ *productService) (*models.CartRender, error) {
	ret := models.CartRender{}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Unable to retrieve cart and lines", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
		return nil, err
	}
	dpi.CartID = id

	if len(lines) == 0 {
		ret.Empty = true
		ret.CartError = "Cart has been checked out already or no longer exists"
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Cart is empty", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
		return &ret, nil
	}

	ret.Cart = *cart

	index := -1
	for i, l := range lines {
		ret.CartLines = append(ret.CartLines, models.CartLineRender{
			ActualLine: *l,
		})
		if l.ID == lineID {
			index = i
		}
	}

	if index == -1 {
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		if err := s.UpdateRender(dpi.Store, &ret, prodServ); err != nil {
			dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Line by given id was deleted, couldn't update render", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
			return nil, err
		}
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Unable to convert cart ID to int", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{errors.New("index == -1")})
		return &ret, nil
	}

	prod, redir, err := prodServ.GetProductByVariantID(dpi.Store, lines[index].VariantID)
	if err != nil {

		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Unable to retrieve product by variant ID", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
		return nil, err
	}

	if redir != "" {
		err := s.cartRepo.DeleteCartLine(*lines[index])
		if err != nil {
			dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Unable to delete cart line after product was redirected", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
			return nil, err
		}
		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		if err := s.UpdateRender(dpi.Store, &ret, prodServ); err != nil {
			dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Unable to update render after product was redirected", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
			return nil, err
		}
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Deleted line and updated render for redirected product", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", nil)
		return &ret, nil
	}

	varIndex := -1
	for i, v := range prod.Variants {
		if v.PK == lines[index].VariantID {
			varIndex = i
		}
	}

	if varIndex == -1 {
		err := s.cartRepo.DeleteCartLine(*lines[index])
		if err != nil {
			dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Variant not in product specified by line; couldn't delete line", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
			return nil, err
		}
		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		if err := s.UpdateRender(dpi.Store, &ret, prodServ); err != nil {
			dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Variant not in product specified by line; couldn't update render", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
			return nil, err
		}
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Variant not in product specified by line", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", nil)
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

	if err := s.cartRepo.SaveCartLineNew(&ret.CartLines[index].ActualLine); err != nil {
		ret.CartLines[index].ActualLine.Quantity = oldQuant
		ret.CartLines[index].ActualLine.Price = product.VolumeDiscPrice(prod.Variants[varIndex].Price, oldQuant, prod.VolumeDisc)
		if err := s.UpdateRender(dpi.Store, &ret, prodServ); err != nil {
			dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Unable to update the render after save", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
			return nil, err
		}
		ret.CartError = "Unable to update cart :/ Please refresh and try again"
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Unable to update the cart after save", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
		return &ret, nil
	}

	if err := s.UpdateRender(dpi.Store, &ret, prodServ); err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Unable to update the render", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
		return nil, err
	}

	dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AdjustQuantity", "Success", "", "", "", strconv.Itoa(lines[index].ProductID), strconv.Itoa(lines[index].VariantID), "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", nil)
	return &ret, nil
}

func (s *cartService) ClearCart(dpi *DataPassIn) (*models.CartRender, error) {
	ret := models.CartRender{}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "ClearCart", "Unable to query cart + lines", "", "", "", "", "", "", "", "", "", "", "", "", []error{err})
		return nil, err
	}
	dpi.CartID = id

	if len(lines) == 0 {
		ret.Empty = true
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "ClearCart", "Cart does not exist to clear", "", "", "", "", "", "", "", "", "", "", "", "", nil)
		return &ret, nil
	}

	if err := s.cartRepo.DeleteCartWithLines(cart.ID); err != nil {
		for _, l := range lines {
			ret.CartLines = append(ret.CartLines, models.CartLineRender{ActualLine: *l})
		}

		ret.Cart = *cart
		carthelp.UpdateCartSub(&ret)

		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "ClearCart", "Could not clear cart", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), "", "", "", []error{errors.New("failed to clear cart")})
		return &ret, errors.New("failed to clear cart")
	}

	ret.Empty = true

	dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "ClearCart", "Success", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), "", "", "", nil)
	return &ret, nil
}

func (s *cartService) AddGiftCard(dpi *DataPassIn, message string, cents int, discService *discountService, tools *config.Tools) (*models.Cart, error) {
	id, cart, err := s.GetCartAndVerify(dpi)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddGiftCard", "Unable to query cart + lines", "", "", "", "", "", "", "", "", "", "", "", "", []error{err})
		return nil, err
	}
	dpi.CartID = id

	idDB, gccode, err := discService.CreateGiftCard(cents, message, dpi.Store, tools)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddGiftCard", "Unable to create gift card to add to cart", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), "", "", "", []error{err})
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
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddGiftCard", "Unable to save line for created gift card", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), "", "", strconv.Itoa(idDB), []error{err})
		return nil, err
	}

	dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "AddGiftCard", "Success", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), strconv.Itoa(line.ID), "", strconv.Itoa(idDB), nil)
	return cart, nil
}

func (s *cartService) DeleteGiftCard(dpi *DataPassIn, lineID int, prodServ *productService) (*models.CartRender, error) {
	ret := models.CartRender{}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "DeleteGiftCard", "Unable to query cart + lines", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
		return nil, err
	}
	dpi.CartID = id

	if len(lines) == 0 {
		ret.Empty = true
		ret.CartError = "Cart has been checked out already or no longer exists"
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "DeleteGiftCard", "Cart has been checked out already or no longer exists", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", nil)
		return &ret, nil
	}

	ret.Cart = *cart

	index := -1
	for i, l := range lines {
		ret.CartLines = append(ret.CartLines, models.CartLineRender{
			ActualLine: *l,
		})
		if l.ID == lineID {
			index = i
		}
	}

	if index == -1 {
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		if err := s.UpdateRender(dpi.Store, &ret, prodServ); err != nil {
			dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "DeleteGiftCard", "Line for gift card was deleted -> Unable to update render", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", []error{err})
			return nil, err
		}
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "DeleteGiftCard", "Line for gift card was deleted", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", "", nil)
		return &ret, nil
	}

	if err := s.cartRepo.DeleteCartLine(*lines[index]); err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "DeleteGiftCard", "Unable to delete cart line for gift card", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", strconv.Itoa(lines[index].VariantID), []error{err})
		return nil, err
	}
	ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)

	if err := s.UpdateRender(dpi.Store, &ret, prodServ); err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "DeleteGiftCard", "Unable to update render after deleting line", "", "", "", "", "", "", "", "", strconv.Itoa(dpi.CartID), strconv.Itoa(lineID), "", strconv.Itoa(lines[index].VariantID), []error{err})
		return nil, err
	}

	dpi.Logger.SaveEvent(dpi.CustomerID, dpi.GuestID, "Cart", "DeleteGiftCard", "Success", "", "", "", "", "", "", "", "", strconv.Itoa(cart.ID), strconv.Itoa(lines[index].ID), "", strconv.Itoa(lines[index].VariantID), nil)
	return &ret, nil
}

func (s *cartService) SavesListToCart(dpi *DataPassIn, varid int, handle string, ps *productService, ls *listService) (models.SavesListRender, *models.CartRender, error) {

	sl, err := ls.DeleteSavesListRender(dpi, varid, 1, ps)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, "", "Cart", "SavesListToCart", "Unable to delete off of saves list", "", "", "", "", strconv.Itoa(varid), "", "", "", "", "", "", "", []error{err})
		return models.SavesListRender{}, nil, err
	}

	_, err = s.AddToCart(dpi, handle, varid, 1, ps)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, "", "Cart", "SavesListToCart", "Unable to add to cart after delete off of saves list", "", "", "", "", strconv.Itoa(varid), "", "", "", "", "", "", "", []error{err})
		return models.SavesListRender{}, nil, err
	}

	cr, err := s.GetCart(dpi, ps)
	if err != nil {
		dpi.Logger.SaveEvent(dpi.CustomerID, "", "Cart", "SavesListToCart", "Unable to add to retrieve cart after adding from saves list", "", "", "", "", strconv.Itoa(varid), "", "", "", "", "", "", "", []error{err})
		return models.SavesListRender{}, nil, err
	}

	dpi.Logger.SaveEvent(dpi.CustomerID, "", "Cart", "SavesListToCart", "Success", "", "", "", "", strconv.Itoa(varid), "", "", "", strconv.Itoa(cr.Cart.ID), "", "", "", nil)
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

func (s *cartService) GetCartMain(dpi *DataPassIn) (*models.Cart, error, bool) {
	cart, err := s.cartRepo.Read(dpi.CartID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err, true
		}
		return nil, err, false
	}

	if cart.Status != "Active" {
		return nil, errors.New("inactive cart"), true
	} else if cart.ID != dpi.CartID {
		return nil, errors.New("queried the incorrect cart by id"), true
	}

	if dpi.CartID > 0 {
		if cart.CustomerID != dpi.CustomerID {
			return nil, errors.New("customer cart doesn't belong to customer"), true
		}
	} else if dpi.GuestID != "" {
		if cart.GuestID != dpi.GuestID {
			return nil, errors.New("guest cart doesn't belong to guest"), true
		}
	} else {
		return nil, errors.New("no one logged in"), true
	}

	return cart, nil, false
}

func (s *cartService) GetCartAndVerify(dpi *DataPassIn) (int, *models.Cart, error) {
	cart, err, retry := s.GetCartMain(dpi)
	if err != nil {
		if retry {
			newID, err := s.CartMiddleware(dpi.CartID, dpi.CustomerID, dpi.GuestID)
			if err != nil {
				return dpi.CartID, nil, err
			}
			cart, err, _ = s.GetCartMain(dpi)
			return newID, cart, err
		}
		return dpi.CartID, nil, err
	}
	return cart.ID, cart, nil
}

func (s *cartService) GetCartMainWithLines(dpi *DataPassIn) (*models.Cart, []*models.CartLine, error, bool) {

	cart, err := s.cartRepo.ReadWithPreload(dpi.CartID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, err, true
		}
		return nil, nil, err, false
	} else if cart.Status != "Active" {
		return nil, nil, errors.New("inactive cart"), true
	} else if cart.ID != dpi.CartID {
		return nil, nil, errors.New("queried the incorrect cart by id"), true
	}

	if dpi.CartID > 0 {
		if cart.CustomerID != dpi.CartID {
			return nil, nil, errors.New("customer cart doesn't belong to customer"), true
		}
	} else if dpi.GuestID != "" {
		if cart.GuestID != dpi.GuestID {
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

func (s *cartService) GetCartWithLinesAndVerify(dpi *DataPassIn) (int, *models.Cart, []*models.CartLine, error) {
	cart, lines, err, retry := s.GetCartMainWithLines(dpi)
	if err != nil {
		if retry {
			newID, err := s.CartMiddleware(dpi.CartID, dpi.CustomerID, dpi.GuestID)
			if err != nil {
				return dpi.CartID, nil, nil, err
			}
			cart, lines, err, _ = s.GetCartMainWithLines(dpi)
			return newID, cart, lines, err
		}
		return dpi.CartID, nil, nil, err
	}
	return cart.ID, cart, lines, nil
}

// Cart ID, count, err
func (s *cartService) CartCountCheck(dpi *DataPassIn) (int, int, error) {
	id, cart, err := s.GetCartAndVerify(dpi)
	if err != nil {
		return dpi.CustomerID, 0, err
	}

	count, err := s.cartRepo.TotalQuantity(cart.ID)
	dpi.Logger.SaveEventNew("Cart", "CartCountCheck", "Success", "", models.EventIDPassIn{CartID: dpi.CartID, CustomerID: dpi.CustomerID, GuestID: dpi.GuestID}, nil)
	return id, count, err
}

func (s *cartService) OrderSuccessCart(dpi *DataPassIn, orderLines []models.OrderLine) error {
	var err error
	var cart models.Cart
	var lines []models.CartLine
	var exists bool

	if dpi.CustomerID > 0 {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByIDAndCustomerID(dpi.CartID, dpi.CustomerID)
	} else if dpi.GuestID != "" {
		cart, lines, exists, err = s.cartRepo.GetCartWithLinesByIDAndGuestID(dpi.CartID, dpi.GuestID)
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

func (s *cartService) CopyCartWithLines(dpi *DataPassIn) error {
	return s.cartRepo.CopyCartWithLines(dpi.CartID, dpi.CustomerID)
}

func (s *cartService) MoveCart(dpi *DataPassIn) error {
	return s.cartRepo.MoveCart(dpi.CartID, dpi.CustomerID)
}

func (s *cartService) DirectCartRetrieval(dpi *DataPassIn) (int, error, bool) {
	return s.cartRepo.DirectCartRetrieval(dpi.CartID, dpi.CustomerID, dpi.GuestID)
}
