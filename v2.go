package checkoutuz

import "context"

// V2BaseURL is the base URL for the CHECKOUT.UZ merchant API v2.
const V2BaseURL = "https://checkout.uz/api/v2"

// V2Client talks to the CHECKOUT.UZ merchant API v2 (/api/v2).
// Authenticated with the same shop API key as v1. v2 adds idempotent
// payment creation, pagination, cancellation/refunds and fiscal receipts
// on top of everything v1 offers.
//
// Note: GetProfile/GetShops/GetProfileStats (on ProfileClient) are
// account-level (not shop-level) and use a different credential.
type V2Client struct {
	base baseClient
}

// NewV2Client creates a V2Client for the given shop API key.
func NewV2Client(apiKey string) (*V2Client, error) {
	base, err := newBaseClient(V2BaseURL, apiKey)
	if err != nil {
		return nil, err
	}
	return &V2Client{base: base}, nil
}

func newV2ClientWithTransport(apiKey string, t transport) (*V2Client, error) {
	base, err := newBaseClientWithTransport(V2BaseURL, apiKey, t)
	if err != nil {
		return nil, err
	}
	return &V2Client{base: base}, nil
}

// CreatePaymentParamsV2 holds the optional fields for V2Client.CreatePayment.
type CreatePaymentParamsV2 struct {
	Description string
	// ExpireTime is in seconds (60-432000). 0 uses the API default (3600).
	ExpireTime int
	WebhookURL string
	ReturnURL  string
	// IdempotencyKey: if the same key was used for this shop within the
	// last 24 hours, the original invoice is returned instead of creating
	// a new one (idempotent_replay in the response tells you which happened).
	IdempotencyKey string
}

// CreatePayment creates a new payment invoice. amount is in UZS
// (1000-10,000,000).
func (c *V2Client) CreatePayment(ctx context.Context, amount float64, params CreatePaymentParamsV2) (map[string]interface{}, error) {
	body := omitEmpty(map[string]interface{}{
		"description": params.Description,
		"expire_time": params.ExpireTime,
		"webhook_url": params.WebhookURL,
		"return_url":  params.ReturnURL,
	})
	body["amount"] = amount

	var headers map[string]string
	if params.IdempotencyKey != "" {
		headers = map[string]string{"Idempotency-Key": params.IdempotencyKey}
	}
	return c.base.post(ctx, "create_payment", body, headers)
}

// StatusPaymentByID looks up an invoice by its numeric id.
func (c *V2Client) StatusPaymentByID(ctx context.Context, id int64) (map[string]interface{}, error) {
	return c.base.post(ctx, "status_payment", map[string]interface{}{"id": id}, nil)
}

// StatusPaymentByUUID looks up an invoice by its uuid (pay_token).
func (c *V2Client) StatusPaymentByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	return c.base.post(ctx, "status_payment", map[string]interface{}{"uuid": uuid}, nil)
}

// GetBalance returns the shop's current balance per currency (uppercase
// currency codes).
func (c *V2Client) GetBalance(ctx context.Context) (map[string]interface{}, error) {
	return c.base.post(ctx, "get_balance", map[string]interface{}{}, nil)
}

// GetHistoryParams holds the optional filters for GetHistory.
type GetHistoryParams struct {
	// Limit is 1-200. 0 uses the API default (10).
	Limit    int
	Offset   int
	Status   string // one of pending/paid/canceled/reversed
	DateFrom string
	DateTo   string
}

// GetHistory returns the shop's invoices, newest first. The response's
// "has_more" tells you whether another page exists.
func (c *V2Client) GetHistory(ctx context.Context, params GetHistoryParams) (map[string]interface{}, error) {
	limit := params.Limit
	if limit == 0 {
		limit = 10
	}
	body := omitEmpty(map[string]interface{}{
		"status":    params.Status,
		"date_from": params.DateFrom,
		"date_to":   params.DateTo,
	})
	body["limit"] = limit
	body["offset"] = params.Offset
	return c.base.post(ctx, "get_history", body, nil)
}

// GetStats returns aggregate order count / income for the shop, broken
// down by payment provider. Pass empty strings to omit the date filters.
func (c *V2Client) GetStats(ctx context.Context, dateFrom, dateTo string) (map[string]interface{}, error) {
	body := omitEmpty(map[string]interface{}{"date_from": dateFrom, "date_to": dateTo})
	return c.base.post(ctx, "get_stats", body, nil)
}

// GetPaymentMethods lists the payment methods enabled for this shop,
// including min/max amount and fiscal support.
func (c *V2Client) GetPaymentMethods(ctx context.Context) (map[string]interface{}, error) {
	return c.base.post(ctx, "get_payment_methods", map[string]interface{}{}, nil)
}

// PayViaCard starts a direct card payment (Uzcard/Humo) for a pending
// invoice. Sends an SMS code to the cardholder.
func (c *V2Client) PayViaCard(ctx context.Context, orderID int64, cardNumber, cardExpiry string) (map[string]interface{}, error) {
	return c.base.post(ctx, "pay_via_card", map[string]interface{}{
		"order_id":    orderID,
		"card_number": cardNumber,
		"card_expiry": cardExpiry,
	}, nil)
}

// ConfirmCardPayment confirms a card payment started with PayViaCard using
// the SMS code the cardholder received.
func (c *V2Client) ConfirmCardPayment(ctx context.Context, paymentToken, smsCode string) (map[string]interface{}, error) {
	return c.base.post(ctx, "confirm_card_payment", map[string]interface{}{
		"payment_token": paymentToken,
		"sms_code":      smsCode,
	}, nil)
}

// CancelPayment cancels a pending invoice. Fails if the invoice is not in
// "pending" status.
func (c *V2Client) CancelPayment(ctx context.Context, orderID int64) (map[string]interface{}, error) {
	return c.base.post(ctx, "cancel_payment", map[string]interface{}{"order_id": orderID}, nil)
}

// RefundPayment reverses a successful card/Payme/Click payment and
// deducts it back from the shop balance. Only available within a limited
// window after the original payment (currently 22 hours); the API
// returns an error past that point.
func (c *V2Client) RefundPayment(ctx context.Context, orderID int64) (map[string]interface{}, error) {
	return c.base.post(ctx, "refund_payment", map[string]interface{}{"order_id": orderID}, nil)
}

// GetFiscal fetches the fiscal (tax) receipt for a paid order. Supported
// for Payme, Click and Plum payments only.
func (c *V2Client) GetFiscal(ctx context.Context, orderID int64) (map[string]interface{}, error) {
	return c.base.post(ctx, "get_fiscal", map[string]interface{}{"order_id": orderID}, nil)
}
