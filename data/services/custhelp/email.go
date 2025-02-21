package custhelp

import (
	"beam/config"
	"net/mail"
	"strings"
)

func VerifyEmail(email string, tools *config.Tools) bool {
	email = strings.ToLower(email)

	_, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	result, err := tools.EmailVerifier.Verify(email)
	if err != nil || result == nil {
		return true
	}

	return result.Syntax.Valid && result.SMTP.Deliverable

}
