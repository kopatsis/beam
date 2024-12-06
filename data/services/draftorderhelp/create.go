package draftorderhelp

import (
	"beam/data/models"
	"strconv"
	"time"
)

func CreateDraftOrder(customer models.Customer, cart models.Cart, cartLines []models.CartLine, products map[int]models.ProductRedis, contact *models.Contact) (*models.DraftOrder, error) {
	var shippingContact *models.OrderContact
	if contact != nil {
		shippingContact = &models.OrderContact{
			FirstName:      contact.FirstName,
			LastName:       *contact.LastName,
			CompanyName:    *contact.Company,
			PhoneNumber:    *contact.PhoneNumber,
			StreetAddress1: contact.StreetAddress1,
			StreetAddress2: *contact.StreetAddress2,
			City:           contact.City,
			ProvinceState:  contact.ProvinceState,
			ZipCode:        contact.ZipCode,
			Country:        contact.Country,
		}
	}

	orderLines := []models.OrderLine{}
	subtotal := 0
	for _, line := range cartLines {
		product := products[line.ProductID]
		orderLine := models.OrderLine{
			ImageURL:          product.ImageURL,
			ProductTitle:      product.Title,
			Handle:            product.Handle,
			PrintfulID:        product.PrintfulID,
			Variant1Key:       line.Variant1Key,
			Variant1Value:     line.Variant1Value,
			Variant2Key:       *line.Variant2Key,
			Variant2Value:     *line.Variant2Value,
			Variant3Key:       *line.Variant3Key,
			Variant3Value:     *line.Variant3Value,
			ProductID:         strconv.Itoa(line.ProductID),
			VariantID:         strconv.Itoa(line.VariantID),
			Quantity:          line.Quantity,
			UndiscountedPrice: line.Price,
			EndPrice:          line.Price,
			LineTotal:         line.Quantity * line.Price,
		}
		orderLines = append(orderLines, orderLine)
		subtotal += line.Quantity * line.Price
	}

	draftOrder := &models.DraftOrder{
		CustomerID:         customer.ID,
		Email:              customer.Email,
		Status:             cart.Status,
		FirstName:          customer.DefaultName,
		LastName:           "",
		DateCreated:        time.Now(),
		Subtotal:           subtotal,
		Shipping:           0,
		OrderLevelDiscount: 0,
		Tax:                0,
		Tip:                0,
		Total:              subtotal,
		ShippingContact:    *shippingContact,
		Lines:              orderLines,
		Guest:              false,
	}

	if cart.GuestID != "" {
		draftOrder.GuestID = &cart.GuestID
	}

	return draftOrder, nil
}
