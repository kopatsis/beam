package draftorderhelp

import (
	"beam/data/models"
	"beam/data/services/product"
	"errors"
	"os"
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
		var orderLine models.OrderLine

		if line.IsGiftCard {
			orderLine = models.OrderLine{
				ImageURL:          os.Getenv("GC_IMG"),
				ProductTitle:      line.ProductTitle,
				Handle:            os.Getenv("GC_HANDLE"),
				Variant1Key:       line.Variant1Key,
				Variant1Value:     line.Variant1Value,
				ProductID:         strconv.Itoa(line.ProductID),
				VariantID:         strconv.Itoa(line.VariantID),
				Quantity:          1,
				UndiscountedPrice: line.Price,
				EndPrice:          line.Price,
				LineTotal:         line.Price,
			}
			subtotal += line.Quantity * line.Price
		} else {
			prod, ok := products[line.ProductID]
			if !ok {
				return nil, errors.New("no matching redis product by id")
			}

			var variant models.VariantRedis
			found := false
			for _, v := range prod.Variants {
				if v.PK == line.VariantID {
					variant = v
				}
			}
			if !found {
				return nil, errors.New("no matching redis variant by id")
			}

			vp := product.VolumeDiscPrice(variant.Price, line.Quantity, prod.VolumeDisc)

			orderLine = models.OrderLine{
				ImageURL:          prod.ImageURL,
				ProductTitle:      prod.Title,
				Handle:            prod.Handle,
				PrintfulID:        prod.PrintfulID,
				Variant1Key:       line.Variant1Key,
				Variant1Value:     line.Variant1Value,
				Variant2Key:       *line.Variant2Key,
				Variant2Value:     *line.Variant2Value,
				Variant3Key:       *line.Variant3Key,
				Variant3Value:     *line.Variant3Value,
				ProductID:         strconv.Itoa(line.ProductID),
				VariantID:         strconv.Itoa(line.VariantID),
				Quantity:          line.Quantity,
				UndiscountedPrice: vp,
				EndPrice:          vp,
				LineTotal:         line.Quantity * vp,
			}
			subtotal += line.Quantity * vp
		}
		orderLines = append(orderLines, orderLine)
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
		PreGiftCardTotal:   subtotal,
		PostTaxTotal:       subtotal,
		GiftCardSum:        0,
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
