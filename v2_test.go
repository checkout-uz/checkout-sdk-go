package checkoutuz

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestV2CreatePaymentWithIdempotencyKeySetsHeader(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success"}`)
	client, _ := newV2ClientWithTransport("test_key_123", ft)

	_, err := client.CreatePayment(context.Background(), 50000, CreatePaymentParamsV2{
		ExpireTime:     7200,
		IdempotencyKey: "order-42",
	})
	if err != nil {
		t.Fatal(err)
	}

	call := ft.calls[0]
	if call.url != "https://checkout.uz/api/v2/create_payment" {
		t.Errorf("unexpected url: %s", call.url)
	}
	var sentBody map[string]interface{}
	json.Unmarshal(call.body, &sentBody)
	if sentBody["amount"].(float64) != 50000 {
		t.Errorf("unexpected amount: %v", sentBody["amount"])
	}
	if sentBody["expire_time"].(float64) != 7200 {
		t.Errorf("unexpected expire_time: %v", sentBody["expire_time"])
	}
	if call.extraHeaders["Idempotency-Key"] != "order-42" {
		t.Errorf("expected Idempotency-Key header, got %v", call.extraHeaders)
	}
}

func TestV2CreatePaymentWithoutIdempotencyKeyOmitsHeader(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success"}`)
	client, _ := newV2ClientWithTransport("test_key_123", ft)

	_, err := client.CreatePayment(context.Background(), 1000, CreatePaymentParamsV2{})
	if err != nil {
		t.Fatal(err)
	}

	if len(ft.calls[0].extraHeaders) != 0 {
		t.Errorf("expected no extra headers, got %v", ft.calls[0].extraHeaders)
	}
}

func TestV2GetHistorySendsFilters(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success","data":[]}`)
	client, _ := newV2ClientWithTransport("test_key_123", ft)

	_, err := client.GetHistory(context.Background(), GetHistoryParams{
		Limit: 20, Offset: 10, Status: "paid", DateFrom: "2026-01-01",
	})
	if err != nil {
		t.Fatal(err)
	}

	var sentBody map[string]interface{}
	json.Unmarshal(ft.calls[0].body, &sentBody)
	if sentBody["limit"].(float64) != 20 {
		t.Errorf("unexpected limit: %v", sentBody["limit"])
	}
	if sentBody["status"] != "paid" {
		t.Errorf("unexpected status: %v", sentBody["status"])
	}
}

func TestV2CancelPayment(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success","order_status":"canceled"}`)
	client, _ := newV2ClientWithTransport("test_key_123", ft)

	result, err := client.CancelPayment(context.Background(), 123)
	if err != nil {
		t.Fatal(err)
	}

	call := ft.calls[0]
	if call.url != "https://checkout.uz/api/v2/cancel_payment" {
		t.Errorf("unexpected url: %s", call.url)
	}
	if result["order_status"] != "canceled" {
		t.Errorf("unexpected order_status: %v", result["order_status"])
	}
}

func TestV2RefundPaymentErrorReturnsAPIError(t *testing.T) {
	ft := newFakeTransport(409, `{"error":"More than 22 hours have passed"}`)
	client, _ := newV2ClientWithTransport("test_key_123", ft)

	_, err := client.RefundPayment(context.Background(), 123)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 409 {
		t.Errorf("unexpected status code: %d", apiErr.StatusCode)
	}
}

func TestV2GetFiscal(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success","fiscal":{}}`)
	client, _ := newV2ClientWithTransport("test_key_123", ft)

	_, err := client.GetFiscal(context.Background(), 123)
	if err != nil {
		t.Fatal(err)
	}

	if ft.calls[0].url != "https://checkout.uz/api/v2/get_fiscal" {
		t.Errorf("unexpected url: %s", ft.calls[0].url)
	}
}

func TestProfileClientUsesDifferentCredentialAgainstSameBaseURL(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success","shops":[]}`)
	profile, _ := newProfileClientWithTransport("profile_key_xyz", ft)

	_, err := profile.GetShops(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	call := ft.calls[0]
	if call.url != "https://checkout.uz/api/v2/get_shops" {
		t.Errorf("unexpected url: %s", call.url)
	}
	if call.token != "profile_key_xyz" {
		t.Errorf("unexpected token: %s", call.token)
	}
}

func TestProfileClientRequiresKey(t *testing.T) {
	if _, err := NewProfileClient(""); err == nil {
		t.Fatal("expected an error for empty API key")
	}
}
