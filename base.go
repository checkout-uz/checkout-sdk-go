package checkoutuz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const defaultTimeout = 20 * time.Second

// baseClient holds the shared request/response handling used by every
// CHECKOUT.UZ API client. Every endpoint accepts a JSON body over POST and
// replies with a JSON body - any non-2xx HTTP status is treated as an
// error.
type baseClient struct {
	baseURL   string
	token     string
	transport transport
}

func newBaseClient(baseURL, token string) (baseClient, error) {
	return newBaseClientWithTransport(baseURL, token, newDefaultTransport(defaultTimeout))
}

// newBaseClientWithTransport is used by tests to inject a fake transport.
func newBaseClientWithTransport(baseURL, token string, t transport) (baseClient, error) {
	if token == "" {
		return baseClient{}, errors.New("an API token is required")
	}
	return baseClient{
		baseURL:   strings.TrimRight(baseURL, "/"),
		token:     token,
		transport: t,
	}, nil
}

func (b *baseClient) post(ctx context.Context, method string, params map[string]interface{}, extraHeaders map[string]string) (map[string]interface{}, error) {
	url := b.baseURL + "/" + method

	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	resp, err := b.transport.send(ctx, url, b.token, jsonBody, extraHeaders)
	if err != nil {
		return nil, &APIError{Message: fmt.Sprintf("network error while calling %s: %v", method, err)}
	}

	var data map[string]interface{}
	if len(resp.Body) > 0 {
		_ = json.Unmarshal(resp.Body, &data) // malformed/non-JSON body -> data stays nil, handled below
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := "Unknown error"
		if data != nil {
			if e, ok := data["error"].(string); ok && e != "" {
				message = e
			} else if m, ok := data["message"].(string); ok && m != "" {
				message = m
			}
		}
		return nil, &APIError{Message: message, StatusCode: resp.StatusCode, Response: data}
	}

	if data == nil {
		data = map[string]interface{}{}
	}
	return data, nil
}

// omitEmpty copies non-zero-value entries from src into a fresh map -
// mirrors the "omit optional params that weren't set" behavior of the
// other language SDKs, using Go's natural zero values ("" / 0) as the
// "not set" sentinel for the params structs each method accepts.
func omitEmpty(src map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case string:
			if val != "" {
				out[k] = val
			}
		case int:
			if val != 0 {
				out[k] = val
			}
		case float64:
			if val != 0 {
				out[k] = val
			}
		default:
			if v != nil {
				out[k] = v
			}
		}
	}
	return out
}
