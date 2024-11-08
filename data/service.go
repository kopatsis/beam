package data

import (
	// "beam/data/repositories"
	"beam/data/services"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type MainService struct {
	User    services.UserService
	Product services.ProductService
}

func NewMainService(pgDBs map[string]*gorm.DB, redis *redis.Client, mongoDBs map[string]*mongo.Database) *MainService {
	return &MainService{
		// User:    services.NewUserService(repositories.NewUserRepository(db, redis)),
		// Product: services.NewProductService(repositories.NewProductRepository(db, redis)),
	}
}
