package services

import (
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/discount"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"
)

type DiscountService interface {
	AddDiscount(dpi *DataPassIn, discount models.Discount) error
	GetDiscountByID(dpi *DataPassIn, id int) (*models.Discount, error)
	UpdateDiscount(dpi *DataPassIn, discount models.Discount) error
	DeleteDiscount(dpi *DataPassIn, id int) error

	CreateGiftCard(dpi *DataPassIn, cents int, message string, store string, tools *config.Tools) (int, string, string, error)
	RenderGiftCard(dpi *DataPassIn, code string) (*models.GiftCardRender, error)
	RetrieveGiftCard(dpi *DataPassIn, code, pin string) (*models.GiftCard, error)
	CheckMultipleGiftCards(dpi *DataPassIn, codesAndAmounts map[[2]string]int) error
	CheckDiscountCode(dpi *DataPassIn, code, store string, subtotal, cust int, noCustomer bool, email string, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) error
	CheckGiftCardsAndDiscountCodes(dpi *DataPassIn, codesAndAmounts map[[2]string]int, code, store string, subtotal int, cust int, noCustomer bool, email string, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (error, error)
	GetDiscountCodeForDraft(dpi *DataPassIn, code, store string, subtotal, cust int, noCustomer bool, email string, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (*models.Discount, []*models.DiscountUser, error)

	UseMultipleGiftCards(dpi *DataPassIn, codesAndAmounts map[[2]string]int, customderID int, guestID, orderID, sessionID string) error
	UseDiscountCode(dpi *DataPassIn, code, guestID, orderID, sessionID, store string, subtotal int, cust int, noCustomer bool, email string, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) error
}

type discountService struct {
	discountRepo repositories.DiscountRepository
}

func NewDiscountService(discountRepo repositories.DiscountRepository) DiscountService {
	return &discountService{discountRepo: discountRepo}
}

func (s *discountService) AddDiscount(dpi *DataPassIn, discount models.Discount) error {
	return s.discountRepo.Create(discount)
}

func (s *discountService) GetDiscountByID(dpi *DataPassIn, id int) (*models.Discount, error) {
	return s.discountRepo.Read(id)
}

func (s *discountService) UpdateDiscount(dpi *DataPassIn, discount models.Discount) error {
	return s.discountRepo.Update(discount)
}

func (s *discountService) DeleteDiscount(dpi *DataPassIn, id int) error {
	return s.discountRepo.Delete(id)
}

func (s *discountService) CreateGiftCard(dpi *DataPassIn, cents int, message string, store string, tools *config.Tools) (int, string, string, error) {
	if len(message) > 256 {
		message = message[:255]
	}

	if cents < 250 {
		return 0, "", "", errors.New("not a large enough amount for gift card")
	} else if cents >= 100000000 {
		return 0, "", "", errors.New("too large amount for gift card")
	}

	var err error
	idSt, exists, iter := "", false, 0
	for !exists && iter < 10 {
		idSt = discount.GenerateCartID()
		exists, err = s.discountRepo.IDCodeExists(idSt)
		if err != nil {
			return 0, "", "", err
		}
		if !exists {
			emails.AlertGiftCardID(idSt, iter, store, tools)
		}
	}

	if !exists {
		return 0, "", "", errors.New("severe issue: could not create an id for gift card in 10 attempts")
	}

	idDB, pin, err := s.discountRepo.CreateGiftCard(idSt, cents, message)
	if err != nil {
		return 0, "", "", err
	}

	return idDB, discount.SpaceDisplayGC(idSt), pin, nil
}

func (s *discountService) RetrieveGiftCard(dpi *DataPassIn, code, pin string) (*models.GiftCard, error) {
	if !discount.CheckID(code) {
		return nil, errors.New("invalid gift card code")
	} else if matched, err := regexp.MatchString(`^\d{3}$`, pin); !matched || err != nil {
		return nil, errors.New("invalid pin code")
	}

	gc, err := s.discountRepo.GetGiftCard(code)
	if err != nil {
		return nil, err
	}

	if gc.Pin != pin {
		return nil, errors.New("pins do not match")
	}

	if gc.Status == "Draft" {
		return nil, errors.New("not yet paid for")
	}

	if gc.Status == "Spent" || gc.LeftoverCents == 0 {
		return nil, errors.New("giftcard spent")
	}

	if gc.Expired.Before(time.Now()) {
		return nil, errors.New("expired")
	}

	return gc, nil
}

func (s *discountService) RenderGiftCard(dpi *DataPassIn, code string) (*models.GiftCardRender, error) {
	if !discount.CheckID(code) {
		return nil, errors.New("invalid gift card code")
	}

	gc, err := s.discountRepo.GetGiftCard(code)
	if err != nil {
		return nil, err
	}
	return &models.GiftCardRender{GiftCard: *gc, Expired: gc.Expired.Before(time.Now())}, nil
}

func (s *discountService) CheckMultipleGiftCards(dpi *DataPassIn, codesAndAmounts map[[2]string]int) error {
	allCodes := []string{}
	re := regexp.MustCompile(`^\d{3}$`)

	for idCode, amt := range codesAndAmounts {
		if amt <= 0 {
			continue
		} else if !discount.CheckID(idCode[0]) {
			continue
		} else if !re.MatchString(idCode[1]) {
			continue
		}
		allCodes = append(allCodes, idCode[0])
	}

	if len(allCodes) == 0 {
		return nil
	}

	allCards, err := s.discountRepo.GetGiftCardsByIDCodes(allCodes)
	if err != nil {
		return err
	} else if len(allCards) != len(allCodes) {
		return fmt.Errorf("issue with checking codes: queried %d, got %d", len(allCards), len(allCodes))
	}

	for _, idCode := range allCodes {

		found, pin, amount := false, "", 0
		for codes, amt := range codesAndAmounts {
			if codes[0] == idCode {
				found = true
				pin = codes[1]
				amount = amt
				break
			}
		}
		if !found {
			return fmt.Errorf("one of the provided id codes not represented: %s", idCode)
		}

		var gc *models.GiftCard
		for _, c := range allCards {
			if c.IDCode == idCode {
				gc = c
			}
		}

		if gc == nil {
			return fmt.Errorf("one of the provided id codes not represented: %s", idCode)
		}

		if gc.Pin != pin {
			return fmt.Errorf("incorrect pin: %s", idCode)
		}

		if gc.Status == "Draft" {
			return fmt.Errorf("not yet paid for: %s", idCode)
		}

		if gc.Status == "Spent" || gc.LeftoverCents == 0 {
			return fmt.Errorf("giftcard spent: %s", idCode)
		}

		if gc.Expired.Before(time.Now()) {
			return fmt.Errorf("expired: %s", idCode)
		}

		if gc.LeftoverCents < amount {
			return fmt.Errorf("cents left over: %d, cents needed: %d", gc.LeftoverCents, amount)
		}
	}

	return nil
}

func (s *discountService) CheckDiscountCode(dpi *DataPassIn, code, store string, subtotal, cust int, noCustomer bool, email string, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) error {

	_, _, err := s.GetDiscountCodeForDraft(dpi, code, store, subtotal, cust, noCustomer, email, storeSettings, tools, cs, ors)
	return err
}

func (s *discountService) CheckGiftCardsAndDiscountCodes(dpi *DataPassIn, codesAndAmounts map[[2]string]int, code, store string, subtotal int, cust int, noCustomer bool, email string, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (error, error) {
	var errGiftCards, errDiscountCodes error

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		errGiftCards = s.CheckMultipleGiftCards(dpi, codesAndAmounts)
	}()

	go func() {
		defer wg.Done()
		errDiscountCodes = s.CheckDiscountCode(dpi, code, store, subtotal, cust, noCustomer, email, storeSettings, tools, cs, ors)
	}()

	wg.Wait()
	return errGiftCards, errDiscountCodes
}

func (s *discountService) GetDiscountCodeForDraft(dpi *DataPassIn, code, store string, subtotal, cust int, noCustomer bool, email string, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) (*models.Discount, []*models.DiscountUser, error) {

	welcomeCode, welcomePct, alwaysCode, alwaysPct := discount.SpecialDiscNames(storeSettings, store)

	if code == welcomeCode {
		if email == "" {
			return nil, nil, errors.New("must supply email for welcome disc")
		}
		if allowed, err := cs.CheckIfValidForWelcome(&DataPassIn{Store: store}, cust, email, ors, tools); err != nil {
			return nil, nil, err
		} else if !allowed {
			return nil, nil, errors.New("welcome disc doesn't apply as this email has an order")
		}
		return &models.Discount{
			ID:              -1,
			DiscountCode:    welcomeCode,
			Status:          "Active",
			IsPercentageOff: true,
			PercentageOff:   welcomePct,
			ShortMessage:    "You are most certainly welcome",
		}, nil, nil
	}

	if code == alwaysCode {
		return &models.Discount{
			ID:              -2,
			DiscountCode:    alwaysCode,
			Status:          "Active",
			IsPercentageOff: true,
			PercentageOff:   alwaysPct,
			ShortMessage:    "Told you it always works",
		}, nil, nil
	}

	disc, users, err := s.discountRepo.GetDiscountWithUsers(code)
	if err != nil {
		return nil, nil, err
	}

	if !disc.AppliesToAllAny {
		if noCustomer {
			return nil, nil, fmt.Errorf("cannot apply discount for specific users without a userid")
		} else if disc.HasUserList {
			contains := false
			for _, user := range users {
				if user.CustomerID == cust {
					contains = true
				}
			}
			if !contains {
				return nil, nil, fmt.Errorf("user not in approved list for discount")
			}
		} else if disc.SingleCustomerID != cust {
			return nil, nil, fmt.Errorf("not approved user for discount")
		}
	}

	if disc.Expired.Before(time.Now()) {
		return nil, nil, fmt.Errorf("expired discount code")
	}

	if disc.Status != "Active" {
		return nil, nil, fmt.Errorf("inactive discount code")
	}

	if !disc.HasUserList && disc.HasMaxUses && disc.Uses >= disc.MaxUses {
		return nil, nil, fmt.Errorf("discount code used more than maximum allowed")
	}

	if disc.HasUserList {
		for _, u := range users {
			if u.CustomerID == cust {
				if disc.HasMaxUses && u.Uses >= disc.MaxUses {
					return nil, nil, fmt.Errorf("discount code used more than maximum allowed for this user")
				}
			}
		}
	}

	if disc.HasMinSubtotal && disc.MinSubtotal > subtotal {
		return nil, nil, fmt.Errorf("subtotal too low for discount code")
	}

	return disc, users, nil
}

func (s *discountService) UseMultipleGiftCards(dpi *DataPassIn, codesAndAmounts map[[2]string]int, customderID int, guestID, orderID, sessionID string) error {
	re := regexp.MustCompile(`^\d{3}$`)

	for idCode, amt := range codesAndAmounts {
		if amt <= 0 {
			continue
		} else if !discount.CheckID(idCode[0]) {
			return errors.New("incorrectly formatted id code: " + idCode[0])
		} else if !re.MatchString(idCode[1]) {
			return fmt.Errorf("incorrectly formatted pin: %s, for id code: %s", idCode[0], idCode[1])
		}
	}

	if len(codesAndAmounts) == 0 {
		return nil
	} else if len(codesAndAmounts) > 3 {
		return errors.New("maximum 3 allowed gift cards to pay for an order")
	}

	uses, err := s.discountRepo.UseGiftCards(codesAndAmounts, orderID, guestID, sessionID, customderID)
	if err != nil {
		return err
	}

	go func() {
		s.discountRepo.GiftCardUseLines(uses)
	}()

	return nil
}

func (s *discountService) UseDiscountCode(dpi *DataPassIn, code, guestID, orderID, sessionID, store string, subtotal int, cust int, noCustomer bool, email string, storeSettings *config.SettingsMutex, tools *config.Tools, cs CustomerService, ors OrderService) error {

	disc, users, err := s.GetDiscountCodeForDraft(dpi, code, store, subtotal, cust, noCustomer, email, storeSettings, tools, cs, ors)
	if err != nil {
		return err
	}

	disc.Uses += 1
	var saveUser *models.DiscountUser

	if disc.HasUserList {
		reachedMax := true
		for _, u := range users {
			if u.CustomerID == cust {
				u.Uses += 1
				saveUser = u
			}
			if u.Uses < disc.MaxUses || !disc.HasMaxUses {
				reachedMax = false
			}
		}
		if reachedMax {
			disc.Status = "Deactivated"
		}
	} else if disc.HasMaxUses && disc.Uses == disc.MaxUses {
		disc.Status = "Deactivated"
	}

	if disc.HasUserList {
		return s.discountRepo.SaveDiscountWithUser(disc, saveUser)
	}

	if err := s.discountRepo.SaveDiscount(disc); err != nil {
		return err
	}

	go func() {
		use := &models.DiscountUseLine{
			DiscountID:   disc.ID,
			DiscountCode: code,
			OrderID:      orderID,
			CustomerID:   cust,
			GuestID:      guestID,
			SessionID:    sessionID,
			Date:         time.Now(),
		}
		s.discountRepo.DiscountUseLine(use)
	}()

	return nil
}
