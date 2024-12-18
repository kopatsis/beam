package repositories

import (
	"time"

	"beam/data/models"

	"gorm.io/gorm"
)

type DiscountRepository interface {
	Create(discount models.Discount) error
	Read(id int) (*models.Discount, error)
	Update(discount models.Discount) error
	Delete(id int) error
	CreateGiftCard(idCode string, cents int, message string) (int, error)
	IDCodeExists(idCode string) (bool, error)
	GetGiftCard(idCode string) (*models.GiftCard, error)
	GetGiftCardsByIDCodes(idCodes []string) ([]*models.GiftCard, error)
	GetDiscountsByCodes(codes []string) ([]*models.Discount, error)
	GetDiscountWithUsers(discountCode string) (*models.Discount, []*models.DiscountUser, error)
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

func (r *discountRepo) CreateGiftCard(idCode string, cents int, message string) (int, error) {
	giftCard := models.GiftCard{
		IDCode:        idCode,
		Created:       time.Now(),
		Expired:       time.Now().AddDate(6, 0, 0),
		Status:        "Draft",
		OriginalCents: cents,
		LeftoverCents: cents,
		ShortMessage:  message,
	}
	if err := r.db.Create(&giftCard).Error; err != nil {
		return 0, err
	}
	return giftCard.ID, nil
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
