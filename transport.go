package checkoutuz

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

// httpResponse is the minimal result a transport must produce.
type httpResponse struct {
	StatusCode int
	Body       []byte
}

// transport is the low-level abstraction every client sends requests
// through. The default implementation (newDefaultTransport) uses
// net/http; tests inject a fake implementation instead of making real
// network calls.
type transport interface {
	send(ctx context.Context, url, token string, body []byte, extraHeaders map[string]string) (httpResponse, error)
}

type defaultTransport struct {
	client *http.Client
}

func newDefaultTransport(timeout time.Duration) *defaultTransport {
	return &defaultTransport{client: &http.Client{Timeout: timeout}}
}

func (t *defaultTransport) send(ctx context.Context, url, token string, body []byte, extraHeaders map[string]string) (httpResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return httpResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return httpResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return httpResponse{}, err
	}

	return httpResponse{StatusCode: resp.StatusCode, Body: respBody}, nil
}
