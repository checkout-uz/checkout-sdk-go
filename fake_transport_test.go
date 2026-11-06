package checkoutuz

import (
	"context"
	"errors"
)

type recordedCall struct {
	url          string
	token        string
	body         []byte
	extraHeaders map[string]string
}

// fakeTransport records every call it receives and returns a canned
// response - no real network calls.
type fakeTransport struct {
	statusCode int
	body       string
	err        error
	calls      []recordedCall
}

func newFakeTransport(statusCode int, body string) *fakeTransport {
	return &fakeTransport{statusCode: statusCode, body: body}
}

func throwingTransport(message string) *fakeTransport {
	return &fakeTransport{err: errors.New(message)}
}

func (f *fakeTransport) send(_ context.Context, url, token string, body []byte, extraHeaders map[string]string) (httpResponse, error) {
	f.calls = append(f.calls, recordedCall{url: url, token: token, body: body, extraHeaders: extraHeaders})
	if f.err != nil {
		return httpResponse{}, f.err
	}
	return httpResponse{StatusCode: f.statusCode, Body: []byte(f.body)}, nil
}
