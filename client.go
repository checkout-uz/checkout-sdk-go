package checkoutuz

import "context"

// CheckoutClient is the top-level entry point for the CHECKOUT.UZ merchant
// API. It holds a single shop API key and exposes both API versions:
//
//	client, _ := checkoutuz.NewCheckoutClient("shop_api_key")
//	client.V1.CreatePayment(ctx, 50000, checkoutuz.CreatePaymentParams{})
//	client.V2.CreatePayment(ctx, 50000, checkoutuz.CreatePaymentParamsV2{ExpireTime: 7200})
//
// For the account-level (not shop-level) v2 endpoints (GetProfile/
// GetShops/GetProfileStats), use ProfileClient instead - they use a
// different credential.
type CheckoutClient struct {
	V1 *V1Client
	V2 *V2Client
}

// NewCheckoutClient creates a CheckoutClient for the given shop API key.
func NewCheckoutClient(apiKey string) (*CheckoutClient, error) {
	v1, err := NewV1Client(apiKey)
	if err != nil {
		return nil, err
	}
	v2, err := NewV2Client(apiKey)
	if err != nil {
		return nil, err
	}
	return &CheckoutClient{V1: v1, V2: v2}, nil
}

// ProfileClient talks to the account-level (profile) v2 endpoints.
// Authenticated with a profile API key (from account settings), sent as
// "Authorization: Bearer <profileApiKey>". This is a different credential
// from a shop's API key used by CheckoutClient - it is not tied to one
// shop and can read data across every shop the account owns.
//
//	profile, _ := checkoutuz.NewProfileClient("profile_api_key")
//	profile.GetShops(ctx)
type ProfileClient struct {
	base baseClient
}

// NewProfileClient creates a ProfileClient for the given profile API key.
func NewProfileClient(profileAPIKey string) (*ProfileClient, error) {
	base, err := newBaseClient(V2BaseURL, profileAPIKey)
	if err != nil {
		return nil, err
	}
	return &ProfileClient{base: base}, nil
}

func newProfileClientWithTransport(profileAPIKey string, t transport) (*ProfileClient, error) {
	base, err := newBaseClientWithTransport(V2BaseURL, profileAPIKey, t)
	if err != nil {
		return nil, err
	}
	return &ProfileClient{base: base}, nil
}

// GetProfile returns the authenticated account's profile information.
func (c *ProfileClient) GetProfile(ctx context.Context) (map[string]interface{}, error) {
	return c.base.post(ctx, "get_profile", map[string]interface{}{}, nil)
}

// GetShops lists every (non-deleted) shop owned by the authenticated account.
func (c *ProfileClient) GetShops(ctx context.Context) (map[string]interface{}, error) {
	return c.base.post(ctx, "get_shops", map[string]interface{}{}, nil)
}

// GetProfileStats returns aggregate stats across all of the account's
// shops, broken down by shop and by provider. Pass empty strings to omit
// the date filters.
func (c *ProfileClient) GetProfileStats(ctx context.Context, dateFrom, dateTo string) (map[string]interface{}, error) {
	body := omitEmpty(map[string]interface{}{"date_from": dateFrom, "date_to": dateTo})
	return c.base.post(ctx, "get_profile_stats", body, nil)
}
