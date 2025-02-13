package draftorderhelp

import (
	"beam/config"
	"beam/data/models"
	"beam/data/services/product"
	"errors"
	"time"
)

func CreateDraftOrder(customer *models.Customer, guestID string, cart models.Cart, cartLines []models.CartLine, products map[int]*models.ProductRedis, contacts []*models.Contact) (*models.DraftOrder, error) {

	orderLines := []models.OrderLine{}
	subtotal := 0
	for _, line := range cartLines {
		var orderLine models.OrderLine

		if line.IsGiftCard {
			orderLine = models.OrderLine{
				ImageURL:          config.GC_IMG,
				ProductTitle:      config.GC_NAME,
				Handle:            config.GC_HANDLE,
				Variant1Key:       "Message",
				Variant1Value:     line.GiftCardMessage,
				ProductID:         line.ProductID,
				VariantID:         line.VariantID,
				Quantity:          1,
				UndiscountedPrice: line.Price,
				Price:             line.Price,
				EndPrice:          line.Price,
				LineTotal:         line.Price,
				IsGiftCard:        true,
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
				PrintfulID:        variant.Printful,
				Variant1Key:       prod.Var1Key,
				Variant1Value:     variant.Var1Value,
				Variant2Key:       prod.Var2Key,
				Variant2Value:     variant.Var2Value,
				Variant3Key:       prod.Var3Key,
				Variant3Value:     variant.Var3Value,
				ProductID:         line.ProductID,
				VariantID:         line.VariantID,
				Quantity:          line.Quantity,
				UndiscountedPrice: variant.Price,
				Price:             vp,
				EndPrice:          vp,
				LineTotal:         line.Quantity * vp,
			}
			subtotal += line.Quantity * vp
		}
		orderLines = append(orderLines, orderLine)
	}

	draftOrder := &models.DraftOrder{
		Status:             "Created",
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
		Lines:              orderLines,
		Guest:              false,
	}

	if len(contacts) > 0 {
		draftOrder.ShippingContact = contacts[0]
		draftOrder.ListedContacts = contacts
	}

	if customer != nil {
		draftOrder.CustomerID = customer.ID
		draftOrder.Email = customer.Email
		draftOrder.Name = customer.DefaultName
		pmid, err := CreatePaymentIntent(customer.StripeID, int64(draftOrder.Subtotal), "usd")
		if err != nil {
			return nil, err
		}
		draftOrder.StripePaymentIntentID = pmid
	} else if guestID != "" {
		draftOrder.GuestID = guestID
		pmid, err := CreatePaymentIntent("", int64(draftOrder.Subtotal), "usd")
		if err != nil {
			return nil, err
		}
		draftOrder.StripePaymentIntentID = pmid
	} else {
		draftOrder.GuestID = cart.GuestID
		pmid, err := CreatePaymentIntent("", int64(draftOrder.Subtotal), "usd")
		if err != nil {
			return nil, err
		}
		draftOrder.StripePaymentIntentID = pmid
	}

	EvaluateFreeShip(draftOrder, customer, products)
	return draftOrder, nil
}
