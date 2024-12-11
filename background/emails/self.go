package emails

import (
	"beam/config"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func AlertEmailRateDanger(store string, wait time.Duration, tools *config.Tools, completed bool) {
	fromEmail := os.Getenv("ADMIN_EMAIL")
	if fromEmail == "" {
		log.Fatal("ADMIN_EMAIL is not set")
	}

	toEmail := fromEmail
	subject := "Alert: High Wait Time for Ship Rate API"

	if !completed {
		subject += " + DID NOT COMPLETE"
	}

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

func AlertIPRateDanger(ip string, wait time.Duration, tools *config.Tools, completed bool) {
	fromEmail := os.Getenv("ADMIN_EMAIL")
	if fromEmail == "" {
		log.Fatal("ADMIN_EMAIL is not set")
	}

	toEmail := fromEmail
	subject := "Alert: High Wait Time for Ship Rate API from IP"

	if !completed {
		subject += " + DID NOT COMPLETE"
	}

	message := fmt.Sprintf("The wait time for the Ship Rate API from IP is very high.\n\nIP: %s\nWait Time: %v\n\nPlease investigate.", ip, wait)

	from := mail.NewEmail("Admin", fromEmail)
	to := mail.NewEmail("Admin", toEmail)
	content := mail.NewContent("text/plain", message)
	mailMessage := mail.NewV3MailInit(from, subject, to, content)

	_, err := tools.SendGrid.Send(mailMessage)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	}
}

func AlertGiftCardID(id string, iter int, store string, tools *config.Tools) {
	fromEmail := os.Getenv("ADMIN_EMAIL")
	if fromEmail == "" {
		log.Fatal("ADMIN_EMAIL is not set")
	}

	toEmail := fromEmail
	subject := "Alert: Managed to Have Duplicate Gift Card ID on attempt " + strconv.Itoa(iter)

	message := fmt.Sprintf("Managed to Have Duplicate Gift Card ID\n\nID: %s\nIteration: %d\n\nStore: %s.", id, iter, store)

	from := mail.NewEmail("Admin", fromEmail)
	to := mail.NewEmail("Admin", toEmail)
	content := mail.NewContent("text/plain", message)
	mailMessage := mail.NewV3MailInit(from, subject, to, content)

	_, err := tools.SendGrid.Send(mailMessage)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	}
}

func HandleWebhook(tools *config.Tools, payload map[string]any) {

	statusToTitle := map[string]string{
		"order_remove_hold":       "Order Remove Hold",
		"order_put_hold_approval": "Order Put Hold Approval",
		"order_put_hold":          "Order Put on Hold",
		"package_returned":        "Package Returned",
		"order_failed":            "Order FAILED",
		"order_canceled":          "Order Cancelled",
		"order_refunded":          "Order Refunded",
	}

	eventType, ok := payload["type"].(string)
	if !ok {
		eventType = "UNKNOWN"
	}

	subject := statusToTitle[eventType]
	if subject == "" {
		subject = eventType
	}

	payloadJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Printf("Error marshalling payload: %v", err)
		return
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		log.Println("ADMIN_EMAIL is not set")
		return
	}

	email := mail.NewV3MailInit(
		mail.NewEmail("Webhook Service", adminEmail),
		subject,
		mail.NewEmail("Admin", adminEmail),
		mail.NewContent("text/plain", string(payloadJSON)),
	)

	response, err := tools.SendGrid.Send(email)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return
	}

	if response.StatusCode >= 400 {
		log.Printf("SendGrid responded with status code %d: %s", response.StatusCode, response.Body)
	}
}
