package emails

import (
	"beam/config"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func AlertEmailRateDanger(store string, wait time.Duration, tools *config.Tools) {
	fromEmail := os.Getenv("ADMIN_EMAIL")
	if fromEmail == "" {
		log.Fatal("ADMIN_EMAIL is not set")
	}

	toEmail := fromEmail
	subject := "Alert: High Wait Time for Ship Rate API"
	message := fmt.Sprintf("The wait time for the Ship Rate API is very high.\n\nStore: %s\nWait Time: %v\n\nPlease investigate.", store, wait)

	from := mail.NewEmail("Admin", fromEmail)
	to := mail.NewEmail("Admin", toEmail)
	content := mail.NewContent("text/plain", message)
	mailMessage := mail.NewV3MailInit(from, subject, to, content)

	_, err := tools.SendGrid.Send(mailMessage)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	}
}
