package services

import (
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/discount"
	"errors"
	"fmt"
	"sync"
	"time"
)

type DiscountService interface {
	AddDiscount(discount models.Discount) error
	GetDiscountByID(id int) (*models.Discount, error)
	UpdateDiscount(discount models.Discount) error
	DeleteDiscount(id int) error
	CreateGiftCard(cents int, message string, store string, tools *config.Tools) (int, string, error)
	RenderGiftCard(code string) (*models.GiftCardRender, error)
	CheckMultipleGiftCards(codesAndAmounts map[string]int) error
	CheckDiscountCode(code string, subtotal int, cust int, noCustomer bool) error
	CheckGiftCardsAndDiscountCodes(codesAndAmounts map[string]int, code string, subtotal int, cust int, noCustomer bool) (error, error)
	GetDiscountCodeForDraft(code string, subtotal, cust int, noCustomer bool) (*models.Discount, []*models.DiscountUser, error)
	UseMultipleGiftCards(codesAndAmounts map[string]int) error
	UseDiscountCode(code string, subtotal int, cust int, noCustomer bool) error
}

type discountService struct {
	discountRepo repositories.DiscountRepository
}

func NewDiscountService(discountRepo repositories.DiscountRepository) DiscountService {
	return &discountService{discountRepo: discountRepo}
}

func (s *discountService) AddDiscount(discount models.Discount) error {
	return s.discountRepo.Create(discount)
}

func (s *discountService) GetDiscountByID(id int) (*models.Discount, error) {
	return s.discountRepo.Read(id)
}

func (s *discountService) UpdateDiscount(discount models.Discount) error {
	return s.discountRepo.Update(discount)
}

func (s *discountService) DeleteDiscount(id int) error {
	return s.discountRepo.Delete(id)
}

func (s *discountService) CreateGiftCard(cents int, message string, store string, tools *config.Tools) (int, string, error) {
	if len(message) > 256 {
		message = message[:255]
	}

	if cents < 100 {
		return 0, "", errors.New("not a large enough amount for gift card")
	} else if cents >= 100000000 {
		return 0, "", errors.New("too large amount for gift card")
	}

	var err error
	idSt, exists, iter := "", false, 0
	for !exists && iter < 10 {
		idSt = discount.GenerateCartID()
		exists, err = s.discountRepo.IDCodeExists(idSt)
		if err != nil {
			return 0, "", err
		}
		if !exists {
			emails.AlertGiftCardID(idSt, iter, store, tools)
		}
	}

	if !exists {
		return 0, "", errors.New("severe issue: could not create an id for gift card in 10 attempts")
	}

	idDB, err := s.discountRepo.CreateGiftCard(idSt, cents, message)
	if err != nil {
		return 0, "", err
	}

	return idDB, idSt, nil
}

func (s *discountService) RetrieveGiftCard(code string, store string) (*models.GiftCard, error) {
	if !discount.CheckID(code) {
		return nil, errors.New("invalid gift card code")
	}

	gc, err := s.discountRepo.GetGiftCard(code)
	if err != nil {
		return nil, err
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

func (s *discountService) RenderGiftCard(code string) (*models.GiftCardRender, error) {
	gc, err := s.discountRepo.GetGiftCard(code)
	if err != nil {
		return nil, err
	}
	return &models.GiftCardRender{GiftCard: *gc, Expired: gc.Expired.Before(time.Now())}, nil
}

func (s *discountService) CheckMultipleGiftCards(codesAndAmounts map[string]int) error {
	allCodes := []string{}
	for idCode, amt := range codesAndAmounts {
		if amt <= 0 {
			continue
		}
		allCodes = append(allCodes, idCode)
	}

	allCards, err := s.discountRepo.GetGiftCardsByIDCodes(allCodes)
	if err != nil {
		return err
	} else if len(allCards) != len(allCodes) {
		return fmt.Errorf("issue with checking codes: queried %d, got %d", len(allCards), len(allCodes))
	}

	for _, idCode := range allCodes {

		amount := codesAndAmounts[idCode]

		var gc *models.GiftCard
		for _, c := range allCards {
			if c.IDCode == idCode {
				gc = c
			}
		}

		if gc == nil {
			return fmt.Errorf("one of the provided id codes not represented: %s", idCode)
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

func (s *discountService) CheckDiscountCode(code string, subtotal int, cust int, noCustomer bool) error {

	_, _, err := s.GetDiscountCodeForDraft(code, subtotal, cust, noCustomer)
	return err
}

func (s *discountService) CheckGiftCardsAndDiscountCodes(codesAndAmounts map[string]int, code string, subtotal int, cust int, noCustomer bool) (error, error) {
	var errGiftCards, errDiscountCodes error

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		errGiftCards = s.CheckMultipleGiftCards(codesAndAmounts)
	}()

	go func() {
		defer wg.Done()
		errDiscountCodes = s.CheckDiscountCode(code, subtotal, cust, noCustomer)
	}()

	wg.Wait()
	return errGiftCards, errDiscountCodes
}

func (s *discountService) GetDiscountCodeForDraft(code string, subtotal, cust int, noCustomer bool) (*models.Discount, []*models.DiscountUser, error) {
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

func (s *discountService) UseMultipleGiftCards(codesAndAmounts map[string]int) error {
	allCodes := []string{}
	for idCode, amt := range codesAndAmounts {
		if amt <= 0 {
			continue
		}
		allCodes = append(allCodes, idCode)
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

		amount := codesAndAmounts[idCode]

		var gc *models.GiftCard
		for _, c := range allCards {
			if c.IDCode == idCode {
				gc = c
			}
		}

		if gc == nil {
			return fmt.Errorf("one of the provided id codes not represented: %s", idCode)
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

		gc.LeftoverCents -= amount

		if gc.LeftoverCents == 0 {
			gc.Status = "Spent"
			gc.Spent = time.Now()
		}
	}

	return s.discountRepo.SaveGiftCards(allCards)
}

func (s *discountService) UseDiscountCode(code string, subtotal int, cust int, noCustomer bool) error {

	disc, users, err := s.GetDiscountCodeForDraft(code, subtotal, cust, noCustomer)
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
	return s.discountRepo.SaveDiscount(disc)
}
