package draftorderhelp

import (
	"beam/data/models"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/paymentmethod"
)

func CreatePaymentIntent(customerID string, amount int64, currency string) (string, error) {
	params := &stripe.PaymentIntentParams{
		Amount:           stripe.Int64(amount),
		Currency:         stripe.String(currency),
		Customer:         stripe.String(customerID),
		SetupFutureUsage: stripe.String("on_session"),
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		return "", err
	}
	return pi.ID, nil
}

func CheckPaymentIntent(paymentIntentID, customerID string, amt int64) error {
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve payment intent: %v", err)
	}

	if pi.Status != stripe.PaymentIntentStatusRequiresConfirmation && pi.Status != stripe.PaymentIntentStatusProcessing && pi.Status != stripe.PaymentIntentStatusSucceeded {
		return fmt.Errorf("payment intent status is not chargeable: %s", pi.Status)
	}

	if pi.Customer == nil || pi.Customer.ID != customerID {
		return fmt.Errorf("customer ID does not match payment intent's customer")
	}

	if pi.Amount != amt {
		return updateStripePaymentIntent(paymentIntentID, int(amt))
	}

	return nil
}

func ConfirmPaymentIntentDraft(draftOrder *models.DraftOrder, customer *models.Customer, guestID string) (custChange bool, draftChange bool, err error) {
	custChange, draftChange = false, false
	if customer == nil {
		if guestID == "" {
			return custChange, draftChange, errors.New("no guest and no customer")
		}
		if draftOrder.GuestStripeID == "" {
			c, err := CreateCustomer("", "", "")
			if err != nil {
				return custChange, draftChange, err
			}
			draftOrder.GuestStripeID = c.ID
			draftChange = true
		}
	} else if customer.StripeID == "" {
		c, err := CreateCustomer(customer.Email, customer.FirstName, customer.LastName)
		if err != nil {
			return custChange, draftChange, err
		}
		customer.StripeID = c.ID
		custChange = true
		draftOrder.CustStripeID = c.ID
		draftChange = true
	} else if draftOrder.CustStripeID == "" {
		draftOrder.CustStripeID = customer.StripeID
		draftChange = true
	}

	useID := ""
	if customer == nil {
		useID = draftOrder.GuestStripeID
	} else {
		useID = customer.StripeID
	}

	needsNew := false
	if draftOrder.StripePaymentIntentID != "" {
		if err := CheckPaymentIntent(draftOrder.StripePaymentIntentID, useID, int64(draftOrder.Total)); err != nil {
			needsNew = true
		}
	} else {
		needsNew = true
	}

	if needsNew {
		id, err := CreatePaymentIntent(useID, int64(draftOrder.Total), "usd")
		if err != nil {
			return custChange, draftChange, err
		}
		draftOrder.StripePaymentIntentID = id
		draftChange = true
	}

	return custChange, draftChange, nil
}

func ValidatePaymentMethod(stripeID, paymentMethodID string) error {
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
		"unknown":    "Other",
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

func DraftPaymentMethodUpdate(draft *models.DraftOrder, stripeID string) error {
	pms, err := GetAllPaymentMethods(stripeID)
	if err != nil {
		return err
	}

	draft.AllPaymentMethods = pms

	if draft.ExistingPaymentMethod.ID != "" {
		found := false
		for _, pm := range pms {
			if pm.ID == draft.ExistingPaymentMethod.ID {
				found = true
				break
			}
		}
		if !found {
			draft.ExistingPaymentMethod = models.PaymentMethodStripe{}
		}
	}

	return nil
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

	if intent.Status != stripe.PaymentIntentStatusSucceeded {
		return nil, fmt.Errorf("payment not successful, status: %s", intent.Status)
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

func CreateCustomer(email, firstname, lastname string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{}
	if email != "" {
		params.Email = stripe.String(email)
	}
	if firstname != "" {
		if lastname != "" {
			params.Name = stripe.String(firstname + " " + lastname)
		}
		params.Name = stripe.String(firstname)
	}
	return customer.New(params)
}

func CreateAndChargePaymentIntent(methodID, customerID string, amount int) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:           stripe.Int64(int64(amount)),
		Currency:         stripe.String("usd"),
		Customer:         stripe.String(customerID),
		PaymentMethod:    stripe.String(methodID),
		Confirm:          stripe.Bool(true),
		SetupFutureUsage: stripe.String("on_session"),
	}
	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	if intent.Status != stripe.PaymentIntentStatusSucceeded {
		return nil, fmt.Errorf("payment not successful, status: %s", intent.Status)
	}

	return intent, nil
}
