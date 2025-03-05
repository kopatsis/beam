package emails

import (
	"beam/config"
	"beam/data/models"
	"strconv"
)

func VerificationEmail(store, email, param string, tools *config.Tools) error {
	panic("not implemented yet")
}

func SignInPin(store, email string, sixDigits uint, tools *config.Tools) error {
	panic("not implemented yet")
}

func TwoFactorEmail(store, email string, sixDigits uint, tools *config.Tools) error {
	panic("not implemented yet")
}

func OrderConfirmAndRate(store, email string, order *models.Order, tools *config.Tools) error {
	panic("not implemented yet")
}

func CustBirthdayEmail(store, email, discCode string, cust *models.Customer, isLeap bool, tools *config.Tools) error {
	panic("not implemented yet")
}

func WelcomeDiscountEmail(store, email string, cust *models.Customer, isWelcome, isCreate bool, storeSettings *config.SettingsMutex, tools *config.Tools) error {
	discCode := config.BASE_WELCOME_CODE
	if !isWelcome {
		discCode = config.BASE_ALWAYS_CODE
	}

	storeSettings.Mu.RLock()
	if isWelcome {
		pct, ok := storeSettings.Settings.WelcomePct[store]
		if ok {
			discCode += strconv.Itoa(pct)
		} else {
			discCode += config.DEFAULT_WELCOME_PCT
		}
	} else {
		pct, ok := storeSettings.Settings.AlwaysWorksPct[store]
		if ok {
			discCode += strconv.Itoa(pct)
		} else {
			discCode += config.DEFAULT_ALWAYS_PCT
		}
	}
	storeSettings.Mu.RUnlock()

	return nil
}
