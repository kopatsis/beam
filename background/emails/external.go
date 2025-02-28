package emails

import (
	"beam/config"
	"beam/data/models"
)

func VerificationEmail(store, email, param string, tools *config.Tools) error {
	panic("not implemented yet")
}

func OrderConfirmAndRate(store, email string, order *models.Order, tools *config.Tools) error {
	panic("not implemented yet")
}
