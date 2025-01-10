package draftorderhelp

import "beam/data/models"

func SetGiftMessage(order *models.DraftOrder, subject, message string) {
	if len(subject) > 128 {
		subject = subject[:125] + "..."
	}
	if len(message) > 1024 {
		message = message[:1021] + "..."
	}

	order.GiftSubject = subject
	order.GiftMessage = message
}
