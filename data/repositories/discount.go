package repositories

import (
	"fmt"
	"log"
	"time"

	"beam/data/models"

	"math/rand"

	"gorm.io/gorm"
)

type DiscountRepository interface {
	Create(discount models.Discount) error
	Read(id int) (*models.Discount, error)
	Update(discount models.Discount) error
	Delete(id int) error
	CreateGiftCard(idCode string, cents int, message string) (int, string, error)
	IDCodeExists(idCode string) (bool, error)
	GetGiftCard(idCode string) (*models.GiftCard, error)
	GetGiftCardsByIDCodes(idCodes []string) ([]*models.GiftCard, error)
	GetDiscountsByCodes(codes []string) ([]*models.Discount, error)
	GetDiscountByCode(code string) (*models.Discount, error)
	GetDiscountWithUsers(discountCode string) (*models.Discount, []*models.DiscountUser, error)
	SaveDiscount(discount *models.Discount) error
	SaveDiscountWithUser(discount *models.Discount, discountUser *models.DiscountUser) error
	SaveGiftCards(giftCards []*models.GiftCard) error

	DiscountUseLine(use *models.DiscountUseLine)
	GiftCardUseLines(uses []*models.GiftCardUseLine)
}

type discountRepo struct {
	db *gorm.DB
}

func NewDiscountRepository(db *gorm.DB) DiscountRepository {
	return &discountRepo{db: db}
}

func (r *discountRepo) Create(discount models.Discount) error {
	return r.db.Create(&discount).Error
}

func (r *discountRepo) Read(id int) (*models.Discount, error) {
	var discount models.Discount
	err := r.db.First(&discount, id).Error
	return &discount, err
}

func (r *discountRepo) Update(discount models.Discount) error {
	return r.db.Save(&discount).Error
}

func (r *discountRepo) Delete(id int) error {
	return r.db.Delete(&models.Discount{}, id).Error
}

func (r *discountRepo) CreateGiftCard(idCode string, cents int, message string) (int, string, error) {

	pin := fmt.Sprintf("%03d", rand.Intn(1000))
	giftCard := models.GiftCard{
		IDCode:        idCode,
		Created:       time.Now(),
		Expired:       time.Now().AddDate(6, 0, 0),
		Status:        "Draft",
		OriginalCents: cents,
		LeftoverCents: cents,
		ShortMessage:  message,
		Pin:           pin,
	}
	if err := r.db.Create(&giftCard).Error; err != nil {
		return 0, "", err
	}
	return giftCard.ID, pin, nil
}

func (r *discountRepo) IDCodeExists(idCode string) (bool, error) {
	var exists bool
	err := r.db.Raw("SELECT EXISTS(SELECT 1 FROM gift_cards WHERE uuid_code = ?)", idCode).Scan(&exists).Error
	return exists, err
}

func (r *discountRepo) GetGiftCard(idCode string) (*models.GiftCard, error) {
	var giftCard models.GiftCard

	err := r.db.Where("id_code = ?", idCode).First(&giftCard).Error
	if err != nil {
		return nil, err
	}

	return &giftCard, nil
}

func (r *discountRepo) GetGiftCardsByIDCodes(idCodes []string) ([]*models.GiftCard, error) {
	var giftCards []*models.GiftCard
	err := r.db.Where("id_code IN ?", idCodes).Find(&giftCards).Error
	return giftCards, err
}

func (r *discountRepo) GetDiscountsByCodes(codes []string) ([]*models.Discount, error) {
	var discounts []*models.Discount
	err := r.db.Where("discount_code IN ?", codes).Find(&discounts).Error
	return discounts, err
}

func (r *discountRepo) GetDiscountByCode(code string) (*models.Discount, error) {
	var discount models.Discount
	err := r.db.Where("discount_code = ?", code).First(&discount).Error
	return &discount, err
}

func (r *discountRepo) GetDiscountWithUsers(discountCode string) (*models.Discount, []*models.DiscountUser, error) {
	var discount models.Discount
	if err := r.db.Where("discount_code = ?", discountCode).First(&discount).Error; err != nil {
		return nil, nil, err
	}

	var discountUsers []*models.DiscountUser
	if discount.HasUserList {
		if err := r.db.Where("discount_id = ?", discount.ID).Find(&discountUsers).Error; err != nil {
			return &discount, nil, err
		}
	}
	return &discount, discountUsers, nil
}

func (r *discountRepo) SaveDiscount(discount *models.Discount) error {
	return r.db.Save(discount).Error
}

func (r *discountRepo) SaveDiscountWithUser(discount *models.Discount, discountUser *models.DiscountUser) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(discount).Error; err != nil {
			return err
		}
		if err := tx.Save(discountUser).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *discountRepo) SaveGiftCards(giftCards []*models.GiftCard) error {
	return r.db.Save(giftCards).Error
}

func (r *discountRepo) DiscountUseLine(use *models.DiscountUseLine) {

	if use == nil {
		return
	}

	if err := r.db.Save(use).Error; err != nil {
		log.Printf("Unable to save discount use line, err: %v; discount ID: %d; orderID: %s\n", err, use.DiscountID, use.OrderID)
	}

}

func (r *discountRepo) GiftCardUseLines(uses []*models.GiftCardUseLine) {
	if len(uses) == 0 {
		return
	}

	if err := r.db.Save(uses).Error; err != nil {
		log.Printf("Unable to save gift card use lines, err: %v\n", err)
	}
}
