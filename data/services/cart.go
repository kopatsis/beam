package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/carthelp"
	"beam/data/services/product"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type CartService interface {
	AddToCart(dpi *DataPassIn, handle string, vid, quant int, prodServ ProductService) (*models.Cart, error)
	GetCart(dpi *DataPassIn, prodServ ProductService) (*models.CartRender, error)
	AdjustQuantity(dpi *DataPassIn, lineID, quant int, prodServ ProductService) (*models.CartRender, error)
	ClearCart(dpi *DataPassIn) (*models.CartRender, error)
	AddGiftCard(dpi *DataPassIn, message string, cents int, discService DiscountService, tools *config.Tools) (*models.Cart, error)
	DeleteGiftCard(dpi *DataPassIn, lineID int, prodServ ProductService) (*models.CartRender, error)
	UpdateRender(dpi *DataPassIn, name string, cart *models.CartRender, ps ProductService) error
	SavesListToCart(dpi *DataPassIn, varid int, handle string, ps ProductService, ls ListService) (models.SavesListRender, *models.CartRender, error)

	CartMiddleware(cartID, custID int, guestID string) (int, error)
	GetCartMain(dpi *DataPassIn) (*models.Cart, error, bool)
	GetCartAndVerify(dpi *DataPassIn) (int, *models.Cart, error)
	GetCartMainWithLines(dpi *DataPassIn) (*models.Cart, []*models.CartLine, error, bool)
	GetCartWithLinesAndVerify(dpi *DataPassIn) (int, *models.Cart, []*models.CartLine, error)
	CartCountCheck(dpi *DataPassIn) (int, int, error)
	OrderSuccessCart(dpi *DataPassIn, orderLines []models.OrderLine) error

	CopyCartWithLines(dpi *DataPassIn) (int, error)
	MoveCart(dpi *DataPassIn) error
	DirectCartRetrieval(dpi *DataPassIn) (int, error, bool)
	GetCartLineWithValidation(dpi *DataPassIn, lineID int) (*models.CartLine, error)

	CopyCartFromShare(dpi *DataPassIn, sharedCartID int) error
}

type cartService struct {
	cartRepo repositories.CartRepository
}

func NewCartService(cartRepo repositories.CartRepository) CartService {
	return &cartService{cartRepo: cartRepo}
}

func (s *cartService) GetCartLineWithValidation(dpi *DataPassIn, lineID int) (*models.CartLine, error) {
	return s.cartRepo.GetCartLineWithValidation(dpi.CustomerID, dpi.CartID, lineID)
}

func (s *cartService) AddToCart(dpi *DataPassIn, handle string, vid, quant int, prodServ ProductService) (*models.Cart, error) {
	p, r, err := prodServ.GetFullProduct(dpi, dpi.Store, handle)
	if err != nil {
		dpi.AddLog("Cart", "AddToCart", "Error querying product", "", err, models.EventPassInFinal{VariantID: vid, CartID: dpi.CartID})
		return nil, err
	} else if r != "" {
		dpi.AddLog("Cart", "AddToCart", "Queried product, but has redirection", "", errors.New("product has a redirection"), models.EventPassInFinal{VariantID: vid, CartID: dpi.CartID})
		return nil, errors.New("product has a redirection")
	}

	index := -1
	for i, v := range p.Variants {
		if v.PK == vid {
			index = i
		}
	}

	if index < 0 {
		dpi.AddLog("Cart", "AddToCart", "No matching variant ID as provided", "", errors.New("no matching variant by id to provided handle"), models.EventPassInFinal{ProductID: p.PK, VariantID: vid, CartID: dpi.CartID})
		return nil, errors.New("no matching variant by id to provided handle")
	}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		dpi.AddLog("Cart", "AddToCart", "Unable to retrieve cart and lines", "", err, models.EventPassInFinal{ProductID: p.PK, VariantID: vid, CartID: dpi.CartID})
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
		dpi.AddLog("Cart", "AddToCart", "Unable to save or add line for cart", "", err, models.EventPassInFinal{ProductID: p.PK, VariantID: vid, CartID: dpi.CartID})
		return nil, err
	}

	dpi.AddLog("Cart", "AddToCart", "", "", nil, models.EventPassInFinal{ProductID: p.PK, VariantID: vid, CartID: dpi.CartID})
	return cart, nil
}

func (s *cartService) GetCart(dpi *DataPassIn, prodServ ProductService) (*models.CartRender, error) {
	ret := models.CartRender{}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		dpi.AddLog("Cart", "GetCart", "Unable to query and verify cart", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return nil, err
	}
	dpi.CartID = id

	if len(lines) == 0 {
		ret.Empty = true
		dpi.AddLog("Cart", "GetCart", "", "Empty cart queried", nil, models.EventPassInFinal{CartID: dpi.CartID})
		return &ret, nil
	}

	for _, l := range lines {
		ret.CartLines = append(ret.CartLines, models.CartLineRender{ActualLine: *l})
	}

	ret.Cart = *cart
	if err := s.UpdateRender(dpi, dpi.Store, &ret, prodServ); err != nil {
		dpi.AddLog("Cart", "GetCart", "Unable to update cart render", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return nil, err
	}

	return &ret, nil
}

func (s *cartService) AdjustQuantity(dpi *DataPassIn, lineID, quant int, prodServ ProductService) (*models.CartRender, error) {
	ret := models.CartRender{}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		dpi.AddLog("Cart", "AdjustQuantity", "Unable to retrieve cart and lines", "", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
		return nil, err
	}
	dpi.CartID = id

	if len(lines) == 0 {
		ret.Empty = true
		ret.CartError = "Cart has been checked out already or no longer exists"
		dpi.AddLog("Cart", "AdjustQuantity", "", "Cart is empty", nil, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
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
		if err := s.UpdateRender(dpi, dpi.Store, &ret, prodServ); err != nil {
			dpi.AddLog("Cart", "AdjustQuantity", "Couldn't update render", "Line by given id was deleted", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
			return nil, err
		}
		dpi.AddLog("Cart", "AdjustQuantity", "", "Line by given id was deleted", nil, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
		return &ret, nil
	}

	prod, redir, err := prodServ.GetProductByVariantID(dpi, dpi.Store, lines[index].VariantID)
	if err != nil {
		dpi.AddLog("Cart", "AdjustQuantity", "Unable to retrieve product by variant ID", "", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
		return nil, err
	}

	if redir != "" {
		err := s.cartRepo.DeleteCartLine(*lines[index])
		if err != nil {
			dpi.AddLog("Cart", "AdjustQuantity", "Unable to delete cart line", "Variant for line has redirect, to be deleted", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
			return nil, err
		}
		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		if err := s.UpdateRender(dpi, dpi.Store, &ret, prodServ); err != nil {
			dpi.AddLog("Cart", "AdjustQuantity", "Unable to update render", "Variant for line has redirect, to be deleted", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
			return nil, err
		}
		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		dpi.AddLog("Cart", "AdjustQuantity", "", "Variant for line has redirect, to be deleted", nil, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
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
			dpi.AddLog("Cart", "AdjustQuantity", "Unable to delete cart line", "Variant not in product specified by line", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
			return nil, err
		}

		ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)
		if err := s.UpdateRender(dpi, dpi.Store, &ret, prodServ); err != nil {
			dpi.AddLog("Cart", "AdjustQuantity", "Unable to update render", "Variant not in product specified by line", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
			return nil, err
		}

		ret.LineError = "That line was deleted. Cart refreshed to latest data."
		dpi.AddLog("Cart", "AdjustQuantity", "", "Variant not in product specified by line", nil, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
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

		dpi.AddLog("Cart", "AdjustQuantity", "Unable to save cart line", "", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})

		if err := s.UpdateRender(dpi, dpi.Store, &ret, prodServ); err != nil {
			dpi.AddLog("Cart", "AdjustQuantity", "Unable to update render after save cart line failure", "", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
			return nil, err
		}

		ret.CartError = "Unable to update cart :/ Please refresh and try again"
		return &ret, nil
	}

	if err := s.UpdateRender(dpi, dpi.Store, &ret, prodServ); err != nil {
		dpi.AddLog("Cart", "AdjustQuantity", "Unable to update render", "", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
		return nil, err
	}

	dpi.AddLog("Cart", "AdjustQuantity", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID, ProductID: lines[index].ProductID, VariantID: lines[index].VariantID})
	return &ret, nil
}

func (s *cartService) ClearCart(dpi *DataPassIn) (*models.CartRender, error) {
	ret := models.CartRender{}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		dpi.AddLog("Cart", "ClearCart", "Unable to query cart + lines", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return nil, err
	}
	dpi.CartID = id

	if len(lines) == 0 {
		ret.Empty = true
		dpi.AddLog("Cart", "ClearCart", "", "Cart does not exist to clear", nil, models.EventPassInFinal{CartID: dpi.CartID})
		return &ret, nil
	}

	if err := s.cartRepo.DeleteCartWithLines(cart.ID); err != nil {
		for _, l := range lines {
			ret.CartLines = append(ret.CartLines, models.CartLineRender{ActualLine: *l})
		}

		ret.Cart = *cart
		carthelp.UpdateCartSub(&ret)

		dpi.AddLog("Cart", "ClearCart", "Could not clear cart", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return &ret, errors.New("failed to clear cart")
	}

	ret.Empty = true

	dpi.AddLog("Cart", "ClearCart", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
	return &ret, nil
}

func (s *cartService) AddGiftCard(dpi *DataPassIn, message string, cents int, discService DiscountService, tools *config.Tools) (*models.Cart, error) {
	id, cart, err := s.GetCartAndVerify(dpi)
	if err != nil {
		dpi.AddLog("Cart", "AddGiftCard", "Unable to query cart + lines", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return nil, err
	}
	dpi.CartID = id

	idDB, _, gccode, err := discService.CreateGiftCard(dpi, cents, message, dpi.Store, tools)
	if err != nil {
		dpi.AddLog("Cart", "AddGiftCard", "Unable to create gift card to add to cart", "", err, models.EventPassInFinal{CartID: dpi.CartID})
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

	// Was previously add line
	if err := s.cartRepo.SaveCartLineNew(&line); err != nil {
		dpi.AddLog("Cart", "AddGiftCard", "Unable to save line for created gift card", "", err, models.EventPassInFinal{CartID: dpi.CartID, GiftCardID: idDB, GiftCardCode: gccode})
		return nil, err
	}

	dpi.AddLog("Cart", "AddGiftCard", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID, GiftCardID: idDB, GiftCardCode: gccode})
	return cart, nil
}

func (s *cartService) DeleteGiftCard(dpi *DataPassIn, lineID int, prodServ ProductService) (*models.CartRender, error) {
	ret := models.CartRender{}

	id, cart, lines, err := s.GetCartWithLinesAndVerify(dpi)
	if err != nil {
		dpi.AddLog("Cart", "DeleteGiftCard", "Unable to query cart + lines", "", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
		return nil, err
	}
	dpi.CartID = id

	if len(lines) == 0 {
		ret.Empty = true
		ret.CartError = "Cart has been checked out already or no longer exists"
		dpi.AddLog("Cart", "DeleteGiftCard", "", "Cart has been checked out already or no longer exists", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
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
		dpi.AddLog("Cart", "DeleteGiftCard", "", "Line for gift card was deleted", nil, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})

		if err := s.UpdateRender(dpi, dpi.Store, &ret, prodServ); err != nil {
			dpi.AddLog("Cart", "DeleteGiftCard", "Unable to update render", "Line for gift card was deleted", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
			return nil, err
		}
		return &ret, nil
	}

	if err := s.cartRepo.DeleteCartLine(*lines[index]); err != nil {
		dpi.AddLog("Cart", "DeleteGiftCard", "Unable to delete cart line for gift card", "", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
		return nil, err
	}
	ret.CartLines = append(ret.CartLines[:index], ret.CartLines[index+1:]...)

	if err := s.UpdateRender(dpi, dpi.Store, &ret, prodServ); err != nil {
		dpi.AddLog("Cart", "DeleteGiftCard", "Unable to update render after deleting line", "", err, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
		return nil, err
	}

	dpi.AddLog("Cart", "DeleteGiftCard", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID, CartLineID: lineID})
	return &ret, nil
}

func (s *cartService) SavesListToCart(dpi *DataPassIn, varid int, handle string, ps ProductService, ls ListService) (models.SavesListRender, *models.CartRender, error) {

	sl, err := ls.DeleteSavesListRender(dpi, varid, 1, ps)
	if err != nil {
		dpi.AddLog("Cart", "SavesListToCart", "Unable to delete off of saves list", "", err, models.EventPassInFinal{CartID: dpi.CartID, VariantID: varid})
		return models.SavesListRender{}, nil, err
	}

	_, err = s.AddToCart(dpi, handle, varid, 1, ps)
	if err != nil {
		dpi.AddLog("Cart", "SavesListToCart", "Unable to add to cart after delete off of saves list", "", err, models.EventPassInFinal{CartID: dpi.CartID, VariantID: varid})
		return models.SavesListRender{}, nil, err
	}

	cr, err := s.GetCart(dpi, ps)
	if err != nil {
		dpi.AddLog("Cart", "SavesListToCart", "Unable to add to retrieve cart after adding from saves list", "", err, models.EventPassInFinal{CartID: dpi.CartID, VariantID: varid})
		return models.SavesListRender{}, nil, err
	}

	dpi.AddLog("Cart", "SavesListToCart", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID, VariantID: varid})
	return sl, cr, nil
}

func (s *cartService) UpdateRender(dpi *DataPassIn, name string, cart *models.CartRender, ps ProductService) error {
	carthelp.UpdateCartSub(cart)
	vids := []int{}

	for _, cl := range cart.CartLines {
		if !cl.ActualLine.IsGiftCard {
			vids = append(vids, cl.ActualLine.VariantID)
		}
	}

	lvs, err := ps.GetLimitedVariants(dpi, name, vids)
	if err != nil {
		dpi.AddLog("Cart", "UpdateRender", "Unable to get limited variants", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return err
	} else if len(lvs) != len(vids) {
		dpi.AddLog("Cart", "UpdateRender", "Variants returned incorrect length equal to variant IDs", "", errors.New("variants returned incorrect length equal to variant IDs"), models.EventPassInFinal{CartID: dpi.CartID})
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
				dpi.AddLog("Cart", "UpdateRender", "Unable to find specific limited variant", "", fmt.Errorf("could not find single lim var for id: %d", cl.ActualLine.VariantID), models.EventPassInFinal{CartID: dpi.CartID})
				return fmt.Errorf("could not find single lim var for id: %d", cl.ActualLine.VariantID)
			}
		}
	}

	dpi.AddLog("Cart", "UpdateRender", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
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
			dpi.AddLog("Cart", "GetCartMain", "Unable to get card by id (no record found)", "", err, models.EventPassInFinal{CartID: dpi.CartID})
			return nil, err, true
		}
		dpi.AddLog("Cart", "GetCartMain", "Unable to get card by id", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return nil, err, false
	}

	if cart.Status != "Active" {
		dpi.AddLog("Cart", "GetCartMain", "Inactive cart by cart id", "", errors.New("inactive cart"), models.EventPassInFinal{CartID: dpi.CartID})
		return nil, errors.New("inactive cart"), true
	} else if cart.ID != dpi.CartID {
		dpi.AddLog("Cart", "GetCartMain", "Incorrect cart queried by id", "", errors.New("queried the incorrect cart by id"), models.EventPassInFinal{CartID: dpi.CartID})
		return nil, errors.New("queried the incorrect cart by id"), true
	}

	if dpi.CartID > 0 {
		if cart.CustomerID != dpi.CustomerID {
			dpi.AddLog("Cart", "GetCartMain", "Customer cart doesn't belong to customer", "", errors.New("customer cart doesn't belong to customer"), models.EventPassInFinal{CartID: dpi.CartID})
			return nil, errors.New("customer cart doesn't belong to customer"), true
		}
	} else if dpi.GuestID != "" {
		if cart.GuestID != dpi.GuestID {
			dpi.AddLog("Cart", "GetCartMain", "Guest cart doesn't belong to guest", "", errors.New("guest cart doesn't belong to guest"), models.EventPassInFinal{CartID: dpi.CartID})
			return nil, errors.New("guest cart doesn't belong to guest"), true
		}
	} else {
		dpi.AddLog("Cart", "GetCartMain", "No customer id or guest id provided", "", errors.New("no one logged in"), models.EventPassInFinal{CartID: dpi.CartID})
		return nil, errors.New("no one logged in"), true
	}

	dpi.AddLog("Cart", "GetCartMain", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
	return cart, nil, false
}

func (s *cartService) GetCartAndVerify(dpi *DataPassIn) (int, *models.Cart, error) {
	cart, err, retry := s.GetCartMain(dpi)
	if err != nil {
		if retry {
			newID, err := s.CartMiddleware(dpi.CartID, dpi.CustomerID, dpi.GuestID)
			if err != nil {
				dpi.AddLog("Cart", "GetCartAndVerify", "Unable to create or get cart", "", err, models.EventPassInFinal{CartID: dpi.CartID})
				return dpi.CartID, nil, err
			}
			cart, err, _ = s.GetCartMain(dpi)
			if err != nil {
				dpi.AddLog("Cart", "GetCartAndVerify", "Unable to get cart by dpi", "", err, models.EventPassInFinal{CartID: dpi.CartID})
				return newID, cart, err
			}
			dpi.AddLog("Cart", "GetCartAndVerify", "", "Had to attempt retry to get cart", nil, models.EventPassInFinal{CartID: dpi.CartID})
			return newID, cart, nil
		}
		dpi.AddLog("Cart", "GetCartAndVerify", "Unable to get cart by GetCartMain", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return dpi.CartID, nil, err
	}
	dpi.AddLog("Cart", "GetCartAndVerify", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
	return cart.ID, cart, nil
}

func (s *cartService) GetCartMainWithLines(dpi *DataPassIn) (*models.Cart, []*models.CartLine, error, bool) {

	cart, err := s.cartRepo.ReadWithPreload(dpi.CartID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			dpi.AddLog("Cart", "GetCartMainWithLines", "Unable to read cart and preload lines (no record found)", "", err, models.EventPassInFinal{CartID: dpi.CartID})
			return nil, nil, err, true
		}
		dpi.AddLog("Cart", "GetCartMainWithLines", "Unable to read cart and preload lines", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return nil, nil, err, false
	} else if cart.Status != "Active" {
		dpi.AddLog("Cart", "GetCartMainWithLines", "Inactive cart", "", errors.New("inactive cart"), models.EventPassInFinal{CartID: dpi.CartID})
		return nil, nil, errors.New("inactive cart"), true
	} else if cart.ID != dpi.CartID {
		dpi.AddLog("Cart", "GetCartMainWithLines", "Queried incorrect cart by id", "", errors.New("queried the incorrect cart by id"), models.EventPassInFinal{CartID: dpi.CartID})
		return nil, nil, errors.New("queried the incorrect cart by id"), true
	}

	if dpi.CartID > 0 {
		if cart.CustomerID != dpi.CartID {
			dpi.AddLog("Cart", "GetCartMainWithLines", "Customer cart doesn't belong to customer", "", errors.New("customer cart doesn't belong to customer"), models.EventPassInFinal{CartID: dpi.CartID})
			return nil, nil, errors.New("customer cart doesn't belong to customer"), true
		}
	} else if dpi.GuestID != "" {
		if cart.GuestID != dpi.GuestID {
			dpi.AddLog("Cart", "GetCartMainWithLines", "Guest cart doesn't belong to guest", "", errors.New("guest cart doesn't belong to guest"), models.EventPassInFinal{CartID: dpi.CartID})
			return nil, nil, errors.New("guest cart doesn't belong to guest"), true
		}
	} else {
		dpi.AddLog("Cart", "GetCartMainWithLines", "No customer id or guest id provided", "", errors.New("no one logged in"), models.EventPassInFinal{CartID: dpi.CartID})
		return nil, nil, errors.New("no one logged in"), true
	}

	cartLines, err := s.cartRepo.CartLinesRetrieval(cart.ID)
	if err != nil {
		dpi.AddLog("Cart", "GetCartMainWithLines", "Unable to read cart lines", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return nil, nil, err, false
	}

	dpi.AddLog("Cart", "GetCartMainWithLines", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
	return cart, cartLines, nil, false
}

func (s *cartService) GetCartWithLinesAndVerify(dpi *DataPassIn) (int, *models.Cart, []*models.CartLine, error) {
	cart, lines, err, retry := s.GetCartMainWithLines(dpi)
	if err != nil {
		if retry {
			newID, err := s.CartMiddleware(dpi.CartID, dpi.CustomerID, dpi.GuestID)
			if err != nil {
				dpi.AddLog("Cart", "GetCartWithLinesAndVerify", "Unable to create or get cart", "", err, models.EventPassInFinal{CartID: dpi.CartID})
				return dpi.CartID, nil, nil, err
			}

			cart, lines, err, _ = s.GetCartMainWithLines(dpi)
			if err != nil {
				dpi.AddLog("Cart", "GetCartWithLinesAndVerify", "Unable to get cart by dpi", "", err, models.EventPassInFinal{CartID: dpi.CartID})
				return newID, cart, lines, err
			}

			dpi.AddLog("Cart", "GetCartWithLinesAndVerify", "", "Had to attempt retry to get cart", nil, models.EventPassInFinal{CartID: dpi.CartID})
			return newID, cart, lines, nil
		}

		dpi.AddLog("Cart", "GetCartWithLinesAndVerify", "Unable to get cart by GetCartMainWithLines", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return dpi.CartID, nil, nil, err
	}

	dpi.AddLog("Cart", "GetCartWithLinesAndVerify", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
	return cart.ID, cart, lines, nil
}

// Cart ID, count, err
func (s *cartService) CartCountCheck(dpi *DataPassIn) (int, int, error) {
	id, cart, err := s.GetCartAndVerify(dpi)
	if err != nil {
		dpi.AddLog("Cart", "CartCountCheck", "Unable to get cart and verify", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return dpi.CustomerID, 0, err
	}

	count, err := s.cartRepo.TotalQuantity(cart.ID)
	if err != nil {
		dpi.AddLog("Cart", "CartCountCheck", "Unable to obtain count from cart", "", err, models.EventPassInFinal{CartID: dpi.CartID})
	}
	dpi.AddLog("Cart", "CartCountCheck", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
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
		dpi.AddLog("Cart", "OrderSuccessCart", "No user id", "", errors.New("no user id of either type provided"), models.EventPassInFinal{CartID: dpi.CartID})
		return errors.New("no user id of either type provided")
	}

	if !exists {
		dpi.AddLog("Cart", "OrderSuccessCart", "", "Cart does not exist", nil, models.EventPassInFinal{CartID: dpi.CartID})
		return nil
	} else if err != nil {
		dpi.AddLog("Cart", "OrderSuccessCart", "Unable to query cart + lines", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return err
	}

	varIDs := map[int]struct{}{}
	for _, l := range orderLines {
		varIDs[l.VariantID] = struct{}{}
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
		if err := s.cartRepo.ArchiveCart(cart.ID); err != nil {
			dpi.AddLog("Cart", "OrderSuccessCart", "Unable to archive full cart", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		}
	}

	if err := s.cartRepo.ReactivateCartWithLines(cart.ID, newLines); err != nil {
		dpi.AddLog("Cart", "OrderSuccessCart", "Unable to reactivate cart with filtered lines", "", err, models.EventPassInFinal{CartID: dpi.CartID})
	}

	dpi.AddLog("Cart", "OrderSuccessCart", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
	return nil
}

func (s *cartService) CopyCartWithLines(dpi *DataPassIn) (int, error) {
	if id, err := s.cartRepo.CopyCartWithLines(dpi.CartID, dpi.CustomerID, dpi.GuestID); err != nil {
		dpi.AddLog("Cart", "CopyCartWithLines", "Unable to copy cart with lines with cart repo", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return id, err
	} else {
		dpi.AddLog("Cart", "CopyCartWithLines", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
		return id, err
	}
}

func (s *cartService) MoveCart(dpi *DataPassIn) error {
	if err := s.cartRepo.MoveCart(dpi.CartID, dpi.CustomerID); err != nil {
		dpi.AddLog("Cart", "MoveCart", "Unable to move cart with cart repo", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return err
	} else {
		dpi.AddLog("Cart", "MoveCart", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
		return nil
	}
}

func (s *cartService) DirectCartRetrieval(dpi *DataPassIn) (int, error, bool) {
	if id, err, rdr := s.cartRepo.DirectCartRetrieval(dpi.CartID, dpi.CustomerID, dpi.GuestID); err != nil {
		dpi.AddLog("Cart", "DirectCartRetrieval", "Unable to directly retrieve cart", "", err, models.EventPassInFinal{CartID: dpi.CartID})
		return id, err, rdr
	} else {
		dpi.AddLog("Cart", "DirectCartRetrieval", "", "", nil, models.EventPassInFinal{CartID: dpi.CartID})
		return id, err, rdr
	}
}

func (s *cartService) CopyCartFromShare(dpi *DataPassIn, sharedCartID int) error {
	newID, err := s.cartRepo.CopyCartWithLines(sharedCartID, dpi.CustomerID, dpi.GuestID)
	if err != nil {
		dpi.AddLog("Cart", "CopyCartFromShare", "Unable to copy cart with lines with cart repo", "", err, models.EventPassInFinal{CartID: sharedCartID})
		return err
	}

	dpi.CartID = newID
	dpi.AddLog("Cart", "CopyCartFromShare", "", "", nil, models.EventPassInFinal{CartID: newID})
	return nil
}
