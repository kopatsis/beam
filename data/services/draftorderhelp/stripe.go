package draftorderhelp

import (
	"beam/data/models"
	"fmt"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/paymentmethod"
)

func CreatePaymentIntent(customerID string, amount int64, currency string) (string, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
	}
	if customerID != "" {
		params.Customer = stripe.String(customerID)
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		return "", err
	}
	return pi.ID, nil
}

func ValidatePaymentMethod(draftOrder *models.DraftOrder, stripeID, paymentMethodID string) error {
	paymentMethod, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch payment method: %w", err)
	}
	if paymentMethod.Customer == nil || paymentMethod.Customer.ID != stripeID {
		return fmt.Errorf("payment method does not belong to customer")
	}
	return nil
}

func GetAllPaymentMethods(stripeID string) ([]models.PaymentMethodStripe, error) {
	var paymentMethods []models.PaymentMethodStripe

	var cardBrandDisplayNames = map[string]string{
		"amex":       "American Express",
		"diners":     "Diners Club",
		"discover":   "Discover",
		"eftpos_au":  "Eftpos Australia",
		"jcb":        "JCB",
		"link":       "Link",
		"mastercard": "MasterCard",
		"unionpay":   "UnionPay",
		"visa":       "Visa",
		"unknown":    "Unknown",
	}

	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(stripeID),
		Type:     stripe.String(string(stripe.PaymentMethodTypeCard)),
	}

	iter := paymentmethod.List(params)

	for iter.Next() {
		paymentMethod := iter.PaymentMethod()
		if paymentMethod.Card != nil {
			cd, ok := cardBrandDisplayNames[string(paymentMethod.Card.Brand)]
			if !ok {
				cd = "Unknown"
			}
			paymentMethods = append(paymentMethods, models.PaymentMethodStripe{
				ID:       paymentMethod.ID,
				CardType: cd,
				Last4:    paymentMethod.Card.Last4,
				ExpMonth: paymentMethod.Card.ExpMonth,
				ExpYear:  paymentMethod.Card.ExpYear,
			})
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error retrieving payment methods: %v", err)
	}

	return paymentMethods, nil
}

func updateStripePaymentIntent(paymentIntentID string, total int) error {
	params := &stripe.PaymentIntentParams{
		Amount: stripe.Int64(int64(total)),
	}
	_, err := paymentintent.Update(paymentIntentID, params)
	if err != nil {
		return fmt.Errorf("failed to update payment intent: %v", err)
	}

	return nil
}
