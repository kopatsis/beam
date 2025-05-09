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
	Cart         services.CartService
	List         services.ListService
	Customer     services.CustomerService
	Product      services.ProductService
	Discount     services.DiscountService
	DraftOrder   services.DraftOrderService
	Order        services.OrderService
	Event        services.EventService
	Notification services.NotificationService
	Session      services.SessionService
	Mutex        *config.AllMutexes
}

type AllServices struct {
	Map   map[string]*MainService
	Mutex *config.AllMutexes
}

func NewMainService(pgDBs map[string]*gorm.DB, redis *redis.Client, mongoDBs map[string]*mongo.Database, mutex *config.AllMutexes) *AllServices {

	ret := AllServices{Map: map[string]*MainService{}, Mutex: mutex}

	mutex.Store.Mu.RLock()

	storeLen := len(mutex.Store.Store.ToDomain)
	ct := 0

	for name := range mutex.Store.Store.ToDomain {
		ret.Map[name] = &MainService{
			Cart:         services.NewCartService(repositories.NewCartRepository(pgDBs[name])),
			List:         services.NewListService(repositories.NewListRepository(pgDBs[name])),
			Customer:     services.NewCustomerService(repositories.NewCustomerRepository(pgDBs[name], redis, name, ct, storeLen)),
			Product:      services.NewProductService(repositories.NewProductRepository(pgDBs[name], redis)),
			Discount:     services.NewDiscountService(repositories.NewDiscountRepository(pgDBs[name])),
			DraftOrder:   services.NewDraftOrderService(repositories.NewDraftOrderRepository(mongoDBs[name])),
			Order:        services.NewOrderService(repositories.NewOrderRepository(mongoDBs[name], redis)),
			Event:        services.NewEventService(repositories.NewEventRepository(mongoDBs[name], redis, name, ct, storeLen)),
			Notification: services.NewNotificationService(repositories.NewNotificationRepository(mongoDBs[name])),
			Session:      services.NewSessionService(repositories.NewSessionRepository(pgDBs[name], redis, name, ct, storeLen)),
		}

		ct++

	}

	mutex.Store.Mu.RUnlock()
	return &ret
}
