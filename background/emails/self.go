package emails

import (
	"beam/background/apidata"
	"beam/config"
	"beam/data/models"
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
		log.Println("ADMIN_EMAIL is not set")
		return
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
		log.Println("ADMIN_EMAIL is not set")
		return
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
		log.Println("ADMIN_EMAIL is not set")
		return
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

func AlertEmailEstRateDanger(store string, wait time.Duration, tools *config.Tools, completed bool) {
	fromEmail := os.Getenv("ADMIN_EMAIL")
	if fromEmail == "" {
		log.Println("ADMIN_EMAIL is not set")
		return
	}

	toEmail := fromEmail
	subject := "Alert: High Wait Time for Order Estimate Rate API"

	if !completed {
		subject += " + DID NOT COMPLETE"
	}

	message := fmt.Sprintf("The wait time for the Order Estimate Rate API is very high.\n\nStore: %s\nWait Time: %v\n\nPlease investigate.", store, wait)

	from := mail.NewEmail("Admin", fromEmail)
	to := mail.NewEmail("Admin", toEmail)
	content := mail.NewContent("text/plain", message)
	mailMessage := mail.NewV3MailInit(from, subject, to, content)

	_, err := tools.SendGrid.Send(mailMessage)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	}
}

func AlertIPEstRateDanger(ip string, wait time.Duration, tools *config.Tools, completed bool) {
	fromEmail := os.Getenv("ADMIN_EMAIL")
	if fromEmail == "" {
		log.Println("ADMIN_EMAIL is not set")
		return
	}

	toEmail := fromEmail
	subject := "Alert: High Wait Time for Order Estimate Rate API from IP"

	if !completed {
		subject += " + DID NOT COMPLETE"
	}

	message := fmt.Sprintf("The wait time for the rder Estimate API from IP is very high.\n\nIP: %s\nWait Time: %v\n\nPlease investigate.", ip, wait)

	from := mail.NewEmail("Admin", fromEmail)
	to := mail.NewEmail("Admin", toEmail)
	content := mail.NewContent("text/plain", message)
	mailMessage := mail.NewV3MailInit(from, subject, to, content)

	_, err := tools.SendGrid.Send(mailMessage)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	}
}

func AlertEstimateTooHigh(store string, draftID string, tools *config.Tools, noProfit bool, cost, price int) {
	fromEmail := os.Getenv("ADMIN_EMAIL")
	if fromEmail == "" {
		log.Println("ADMIN_EMAIL is not set")
	}

	toEmail := fromEmail
	subject := "SERIOUS Alert: Order Estimate Too High"

	if noProfit {
		subject += " + ZERO OR NEGATIVE PROFIT"
	}

	message := fmt.Sprintf("The order estimate is too high for this current cost.\n\nStore: %s\nDraft Order ID: %s\nOrder Estimate in cents: %d\nPre Gift Card Total in cents: %d\n\nCHECK NOW.", store, draftID, cost, price)

	from := mail.NewEmail("Admin", fromEmail)
	to := mail.NewEmail("Admin", toEmail)
	content := mail.NewContent("text/plain", message)
	mailMessage := mail.NewV3MailInit(from, subject, to, content)

	_, err := tools.SendGrid.Send(mailMessage)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return
	}
}

func OrderSuccessWithProfit(store string, orderID, printfulID string, tools *config.Tools, cost, price int) {
	fromEmail := os.Getenv("ADMIN_EMAIL")
	if fromEmail == "" {
		log.Println("ADMIN_EMAIL is not set")
	}

	toEmail := fromEmail
	subject := store + ": Order Went Through"

	if price <= cost {
		subject = "NO OR NEGATIVE PROFIT ALERT -- "
	}

	message := fmt.Sprintf("An order successfully went through with this information.\n\nStore: %s\nOrder ID: %s\nPrintful ID: %s\nOrder Cost in cents: %d\nPre Gift Card Total in cents: %d.", store, orderID, printfulID, cost, price)

	from := mail.NewEmail("Admin", fromEmail)
	to := mail.NewEmail("Admin", toEmail)
	content := mail.NewContent("text/plain", message)
	mailMessage := mail.NewV3MailInit(from, subject, to, content)

	_, err := tools.SendGrid.Send(mailMessage)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	}
}

func AlertRecoverableOrderSubmitError(store string, draftID, orderID, explain string, tools *config.Tools, order *models.Order, draft *models.DraftOrder, pfresponse *apidata.OrderResponse, isExtreme bool, providedErr error) {
	fromEmail := os.Getenv("ADMIN_EMAIL")
	if isExtreme {
		fromEmail = os.Getenv("ADMIN_EMAIL_CATAS")
	}

	if fromEmail == "" {
		log.Println("ADMIN_EMAIL is not set")
		return
	}

	toEmail := fromEmail
	subject := "ALERT -- " + store + ": Error Submitting Order = " + explain

	if isExtreme {
		subject = "EXTREME " + subject
	}

	var message string
	if isExtreme {
		message = fmt.Sprintf("UNRECOVERABLE error when submitting an order.\n\nGiven Error: %v\nDetails: %s\nStore: %s\nDraft Order ID: %s\nOrder ID: %s", providedErr, explain, store, draftID, orderID)
	} else {
		message = fmt.Sprintf("There was a recoverable error when submitting an order.\n\nGiven Error: %v\nDetails: %s\nStore: %s\nDraft Order ID: %s\nOrder ID: %s", providedErr, explain, store, draftID, orderID)
	}

	if draft != nil {

		message += "\n\nJSON of present Draft Order struct:\n\n"

		payloadJSON, err := json.MarshalIndent(draft, "", "  ")
		if err == nil {
			message += string(payloadJSON)
		} else {
			message += "Unable to be marshaled: " + err.Error()
		}
	}

	if order != nil {

		message += "\n\nJSON of present Order struct:\n\n"

		payloadJSON, err := json.MarshalIndent(order, "", "  ")
		if err == nil {
			message += string(payloadJSON)
		} else {
			message += "Unable to be marshaled: " + err.Error()
		}
	}

	if pfresponse != nil {

		message += "\n\nJSON of present Printful Reponse struct:\n\n"

		payloadJSON, err := json.MarshalIndent(pfresponse, "", "  ")
		if err == nil {
			message += string(payloadJSON)
		} else {
			message += "Unable to be marshaled: " + err.Error()
		}
	}

	from := mail.NewEmail("Admin", fromEmail)
	to := mail.NewEmail("Admin", toEmail)
	content := mail.NewContent("text/plain", message)
	mailMessage := mail.NewV3MailInit(from, subject, to, content)

	_, err := tools.SendGrid.Send(mailMessage)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	}
}
