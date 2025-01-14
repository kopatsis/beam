package config

import (
	"beam/data/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func PostgresConnect(mutex *AllMutexes) map[string]*gorm.DB {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	ret := map[string]*gorm.DB{}
	mutex.Store.Mu.RLock()

	for dbName := range mutex.Store.Store.ToDomain {
		dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
			dbUser, dbPassword, dbName)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}

		err = db.AutoMigrate(&models.Cart{}, &models.CartLine{}, &models.Comparable{}, &models.Contact{}, &models.Customer{}, &models.Discount{}, &models.DiscountUser{}, &models.FavesLine{}, &models.SavesList{}, &models.LastOrdersList{}, &models.Product{}, &models.Variant{})
		if err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}

		ret[dbName] = db
	}

	mutex.Store.Mu.RUnlock()
	log.Println("Database migration completed successfully")

	return ret
}
