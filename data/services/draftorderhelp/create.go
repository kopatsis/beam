package draftorderhelp

import (
	"beam/config"
	"beam/data/models"
	"beam/data/services/product"
	"errors"
	"time"
)

func CreateDraftOrder(customer *models.Customer, guestID string, cart *models.Cart, cartLines []*models.CartLine, products map[int]*models.ProductRedis, contacts []*models.Contact) (*models.DraftOrder, error) {

	orderLines, gcLines := []models.OrderLine{}, []models.GiftCardBuyLine{}
	subtotal, gcTotal := 0, 0
	for _, line := range cartLines {
		if line.IsGiftCard {
			orderLine := models.GiftCardBuyLine{
				ImageURL:     config.GC_IMG,
				ProductTitle: config.GC_NAME,
				Handle:       config.GC_HANDLE,
				Message:      line.GiftCardMessage,
				CardID:       line.VariantID,
				CardCode:     line.GiftCardCode,
				Price:        line.Price,
			}
			gcTotal += line.Price
			gcLines = append(gcLines, orderLine)
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

			orderLine := models.OrderLine{
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

			orderLines = append(orderLines, orderLine)
		}

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
		GiftCardBuyTotal:   gcTotal,
		GiftCardSum:        0,
		Total:              subtotal + gcTotal,
		Lines:              orderLines,
		GiftCardBuyLines:   gcLines,
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
		pmid, err := CreatePaymentIntent(customer.StripeID, int64(draftOrder.Total), "usd")
		if err != nil {
			return nil, err
		}
		draftOrder.StripePaymentIntentID = pmid
	} else if guestID != "" {
		draftOrder.GuestID = guestID
		pmid, err := CreatePaymentIntent("", int64(draftOrder.Total), "usd")
		if err != nil {
			return nil, err
		}
		draftOrder.StripePaymentIntentID = pmid
	} else {
		draftOrder.GuestID = cart.GuestID
		pmid, err := CreatePaymentIntent("", int64(draftOrder.Total), "usd")
		if err != nil {
			return nil, err
		}
		draftOrder.StripePaymentIntentID = pmid
	}

	EvaluateFreeShip(draftOrder, customer, products)
	return draftOrder, nil
}
