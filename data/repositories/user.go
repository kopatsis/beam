package repositories

import (
	"beam/data/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user models.User) error
	FindByID(uid string) (*models.User, error)
	Update(user models.User) error
	Delete(uid string) error
}

type userRepo struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewUserRepository(db *gorm.DB, redis *redis.Client) UserRepository {
	return &userRepo{db: db, redis: redis}
}

func (r *userRepo) Create(user models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepo) FindByID(uid string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "uid = ?", uid).Error
	return &user, err
}

func (r *userRepo) Update(user models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepo) Delete(uid string) error {
	return r.db.Delete(&models.User{}, "uid = ?", uid).Error
}
