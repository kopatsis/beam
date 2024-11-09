package data

import (
	// "beam/data/repositories"
	"beam/config"
	"beam/data/services"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type MainService struct {
	User    services.UserService
	Product services.ProductService
}

type AllServices struct {
	Map map[string]*MainService
}

func NewMainService(pgDBs map[string]*gorm.DB, redis *redis.Client, mongoDBs map[string]*mongo.Database, mutex *config.AllMutexes) *AllServices {

	ret := AllServices{Map: map[string]*MainService{}}

	mutex.Store.Mu.RLock()

	for name := range mutex.Store.Store.ToDomain {
		ret.Map[name] = &MainService{}
	}

	mutex.Store.Mu.RUnlock()
	return &ret

	// return &MainService{
	// 	User:    services.NewUserService(repositories.NewUserRepository(db, redis)),
	// 	Product: services.NewProductService(repositories.NewProductRepository(db, redis)),
	// }
}
