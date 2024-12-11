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
