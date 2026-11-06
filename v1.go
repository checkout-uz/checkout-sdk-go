package checkoutuz

import "context"

// V1BaseURL is the base URL for the original CHECKOUT.UZ merchant API.
const V1BaseURL = "https://checkout.uz/api/v1"

// V1Client talks to the original CHECKOUT.UZ merchant API (/api/v1).
// Authenticated with a shop's API key (found in the shop's settings on
// checkout.uz), sent as "Authorization: Bearer <apiKey>".
type V1Client struct {
	base baseClient
}

// NewV1Client creates a V1Client for the given shop API key.
func NewV1Client(apiKey string) (*V1Client, error) {
	base, err := newBaseClient(V1BaseURL, apiKey)
	if err != nil {
		return nil, err
	}
	return &V1Client{base: base}, nil
}

func newV1ClientWithTransport(apiKey string, t transport) (*V1Client, error) {
	base, err := newBaseClientWithTransport(V1BaseURL, apiKey, t)
	if err != nil {
		return nil, err
	}
	return &V1Client{base: base}, nil
}

// CreatePaymentParams holds the optional fields for CreatePayment.
type CreatePaymentParams struct {
	Description string
	WebhookURL  string
	ReturnURL   string
}

// CreatePayment creates a new payment invoice. amount is in UZS
// (1000-10,000,000).
func (c *V1Client) CreatePayment(ctx context.Context, amount float64, params CreatePaymentParams) (map[string]interface{}, error) {
	body := omitEmpty(map[string]interface{}{
		"description": params.Description,
		"webhook_url": params.WebhookURL,
		"return_url":  params.ReturnURL,
	})
	body["amount"] = amount
	return c.base.post(ctx, "create_payment", body, nil)
}

// StatusPaymentByID looks up an invoice by its numeric id.
func (c *V1Client) StatusPaymentByID(ctx context.Context, id int64) (map[string]interface{}, error) {
	return c.base.post(ctx, "status_payment", map[string]interface{}{"id": id}, nil)
}

// StatusPaymentByUUID looks up an invoice by its uuid (pay_token).
func (c *V1Client) StatusPaymentByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	return c.base.post(ctx, "status_payment", map[string]interface{}{"uuid": uuid}, nil)
}

// GetBalance returns the shop's current balance per currency.
func (c *V1Client) GetBalance(ctx context.Context) (map[string]interface{}, error) {
	return c.base.post(ctx, "get_balance", map[string]interface{}{}, nil)
}

// GetHistory returns the shop's most recent invoices (newest first). Pass
// 0 for limit to use the API default (10).
func (c *V1Client) GetHistory(ctx context.Context, limit int) (map[string]interface{}, error) {
	if limit == 0 {
		limit = 10
	}
	return c.base.post(ctx, "get_history", map[string]interface{}{"limit": limit}, nil)
}

// GetStats returns aggregate order count / total income for the shop.
func (c *V1Client) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return c.base.post(ctx, "get_stats", map[string]interface{}{}, nil)
}

// GetPaymentMethods lists the payment methods enabled for this shop.
func (c *V1Client) GetPaymentMethods(ctx context.Context) (map[string]interface{}, error) {
	return c.base.post(ctx, "get_payment_methods", map[string]interface{}{}, nil)
}

// PayViaCard starts a direct card payment (Uzcard/Humo) for a pending
// invoice. Sends an SMS code to the cardholder.
func (c *V1Client) PayViaCard(ctx context.Context, orderID int64, cardNumber, cardExpiry string) (map[string]interface{}, error) {
	return c.base.post(ctx, "pay_via_card", map[string]interface{}{
		"order_id":    orderID,
		"card_number": cardNumber,
		"card_expiry": cardExpiry,
	}, nil)
}

// ConfirmCardPayment confirms a card payment started with PayViaCard using
// the SMS code the cardholder received.
func (c *V1Client) ConfirmCardPayment(ctx context.Context, paymentToken, smsCode string) (map[string]interface{}, error) {
	return c.base.post(ctx, "confirm_card_payment", map[string]interface{}{
		"payment_token": paymentToken,
		"sms_code":      smsCode,
	}, nil)
}
