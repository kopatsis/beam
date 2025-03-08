package emails

import (
	"beam/config"
	"beam/data/models"
	"beam/data/services/discount"
	"fmt"
)

func VerificationEmail(store, email, param, ipStr string, tools *config.Tools) error {
	city, country := config.GetLocation(ipStr, tools)
	if city == "" && country == "" {
		fmt.Println("unknown location")
	}
	panic("not implemented yet")
}

func SignInPin(store, email, ipStr string, sixDigits uint, tools *config.Tools) error {
	city, country := config.GetLocation(ipStr, tools)
	if city == "" && country == "" {
		fmt.Println("unknown location")
	}
	panic("not implemented yet")
}

func TwoFactorEmail(store, email, ipStr string, sixDigits uint, tools *config.Tools) error {
	city, country := config.GetLocation(ipStr, tools)
	if city == "" && country == "" {
		fmt.Println("unknown location")
	}
	panic("not implemented yet")
}

func ResetEmail(store, email, param, ipStr string, tools *config.Tools) error {
	city, country := config.GetLocation(ipStr, tools)
	if city == "" && country == "" {
		fmt.Println("unknown location")
	}
	panic("not implemented yet")
}

func OrderConfirmAndRate(store, email string, order *models.Order, tools *config.Tools) error {
	panic("not implemented yet")
}

func CustBirthdayEmail(store, email, discCode string, cust *models.Customer, isLeap bool, tools *config.Tools) error {
	panic("not implemented yet")
}

func WelcomeDiscountEmail(store, email string, cust *models.Customer, isWelcome, isCreate bool, storeSettings *config.SettingsMutex, tools *config.Tools) error {
	var discCode string

	welcome, _, always, _ := discount.SpecialDiscNames(storeSettings, store)
	if !isWelcome {
		discCode = welcome
	} else {
		discCode = always
	}

	fmt.Println(discCode)

	return nil
}
