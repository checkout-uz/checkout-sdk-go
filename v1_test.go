package checkoutuz

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestV1RequiresAPIKey(t *testing.T) {
	if _, err := NewV1Client(""); err == nil {
		t.Fatal("expected an error for empty API key")
	}
}

func TestV1CreatePaymentSendsCorrectRequest(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success","payment":{"_id":1}}`)
	client, err := newV1ClientWithTransport("test_key_123", ft)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.CreatePayment(context.Background(), 50000, CreatePaymentParams{Description: "test order"})
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "success" {
		t.Errorf("expected status=success, got %v", result["status"])
	}

	call := ft.calls[0]
	if call.url != "https://checkout.uz/api/v1/create_payment" {
		t.Errorf("unexpected url: %s", call.url)
	}
	if call.token != "test_key_123" {
		t.Errorf("unexpected token: %s", call.token)
	}
	var sentBody map[string]interface{}
	json.Unmarshal(call.body, &sentBody)
	if sentBody["amount"].(float64) != 50000 {
		t.Errorf("unexpected amount: %v", sentBody["amount"])
	}
	if sentBody["description"] != "test order" {
		t.Errorf("unexpected description: %v", sentBody["description"])
	}
	if _, has := sentBody["webhook_url"]; has {
		t.Error("webhook_url should have been omitted")
	}
}

func TestV1OptionalEmptyParamsAreOmitted(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success"}`)
	client, _ := newV1ClientWithTransport("test_key_123", ft)

	_, err := client.CreatePayment(context.Background(), 1000, CreatePaymentParams{})
	if err != nil {
		t.Fatal(err)
	}

	var sentBody map[string]interface{}
	json.Unmarshal(ft.calls[0].body, &sentBody)
	if len(sentBody) != 1 {
		t.Errorf("expected only 'amount' key, got %v", sentBody)
	}
}

func TestV1ErrorResponseReturnsAPIError(t *testing.T) {
	ft := newFakeTransport(401, `{"error":"Unauthorized: Invalid or missing Bearer token"}`)
	client, _ := newV1ClientWithTransport("test_key_123", ft)

	_, err := client.GetBalance(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("unexpected status code: %d", apiErr.StatusCode)
	}
}

func TestV1NetworkErrorReturnsAPIError(t *testing.T) {
	ft := throwingTransport("boom")
	client, _ := newV1ClientWithTransport("test_key_123", ft)

	_, err := client.GetBalance(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
}

func TestV1PayViaCardSendsCorrectBody(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success"}`)
	client, _ := newV1ClientWithTransport("test_key_123", ft)

	_, err := client.PayViaCard(context.Background(), 5, "8600123412341234", "1230")
	if err != nil {
		t.Fatal(err)
	}

	call := ft.calls[0]
	if call.url != "https://checkout.uz/api/v1/pay_via_card" {
		t.Errorf("unexpected url: %s", call.url)
	}
	var sentBody map[string]interface{}
	json.Unmarshal(call.body, &sentBody)
	if sentBody["order_id"].(float64) != 5 {
		t.Errorf("unexpected order_id: %v", sentBody["order_id"])
	}
}

func TestV1StatusPaymentByUUID(t *testing.T) {
	ft := newFakeTransport(200, `{"status":"success"}`)
	client, _ := newV1ClientWithTransport("test_key_123", ft)

	_, err := client.StatusPaymentByUUID(context.Background(), "abc-123")
	if err != nil {
		t.Fatal(err)
	}

	var sentBody map[string]interface{}
	json.Unmarshal(ft.calls[0].body, &sentBody)
	if sentBody["uuid"] != "abc-123" {
		t.Errorf("unexpected uuid: %v", sentBody["uuid"])
	}
}
