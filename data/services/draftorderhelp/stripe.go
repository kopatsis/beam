package draftorderhelp

import (
	"beam/data/models"
	"fmt"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
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
	params.SetupFutureUsage = stripe.String("off_session")
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

func DetachPaymentMethod(customerID, paymentMethodID string) error {
	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("error fetching payment method: %w", err)
	}

	if pm.Customer == nil || pm.Customer.ID != customerID {
		return fmt.Errorf("payment method %s does not belong to customer %s", paymentMethodID, customerID)
	}

	_, err = paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("error detaching payment method: %w", err)
	}

	return nil
}

func ChargePaymentIntent(intentID, methodID string, save bool, customerID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentConfirmParams{
		PaymentMethod: stripe.String(methodID),
	}
	intent, err := paymentintent.Confirm(intentID, params)
	if err != nil {
		return nil, err
	}
	if save {
		if err := AttachPaymentMethodToCustomer(methodID, customerID); err != nil {
			return nil, err
		}
	}
	return intent, nil
}

func AttachPaymentMethodToCustomer(methodID, customerID string) error {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	_, err := paymentmethod.Attach(methodID, params)
	return err
}

func CreateCustomer(email, name string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{}
	if email != "" {
		params.Email = stripe.String(email)
	}
	if name != "" {
		params.Name = stripe.String(name)
	}
	return customer.New(params)
}

func CreateAndChargePaymentIntent(methodID, customerID string, amount int) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(amount)),
		Currency:      stripe.String("usd"),
		Customer:      stripe.String(customerID),
		PaymentMethod: stripe.String(methodID),
		Confirm:       stripe.Bool(true),
	}
	return paymentintent.New(params)
}
