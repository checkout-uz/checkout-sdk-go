package checkoutuz

import "fmt"

// APIError is returned whenever the CHECKOUT.UZ API replies with a non-2xx
// HTTP status, or a network error occurs. StatusCode is 0 if the request
// never reached the server.
type APIError struct {
	Message    string
	StatusCode int
	Response   map[string]interface{}
}

func (e *APIError) Error() string {
	if e.StatusCode == 0 {
		return e.Message
	}
	return fmt.Sprintf("checkout.uz API error (status %d): %s", e.StatusCode, e.Message)
}
