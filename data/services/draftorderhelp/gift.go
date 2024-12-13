package draftorderhelp

import "beam/data/models"

func SetGiftMessage(order *models.DraftOrder, subject, message string) {
	order.GiftSubject = subject
	order.GiftMessage = message
}
