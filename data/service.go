package data

import (
	// "beam/data/repositories"
	"beam/config"
	"beam/data/repositories"
	"beam/data/services"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type MainService struct {
	Cart     services.CartService
	List     services.ListService
	Customer services.CustomerService
	Product  services.ProductService
}

type AllServices struct {
	Map map[string]*MainService
}

func NewMainService(pgDBs map[string]*gorm.DB, redis *redis.Client, mongoDBs map[string]*mongo.Database, mutex *config.AllMutexes) *AllServices {

	ret := AllServices{Map: map[string]*MainService{}}

	mutex.Store.Mu.RLock()

	for name := range mutex.Store.Store.ToDomain {
		ret.Map[name] = &MainService{
			Cart:     services.NewCartService(repositories.NewCartRepository(pgDBs[name])),
			List:     services.NewListService(repositories.NewListRepository(pgDBs[name])),
			Customer: services.NewCustomerService(repositories.NewCustomerRepository(pgDBs[name])),
			Product:  services.NewProductService(repositories.NewProductRepository(pgDBs[name], redis)),
		}
	}

	mutex.Store.Mu.RUnlock()
	return &ret
}
