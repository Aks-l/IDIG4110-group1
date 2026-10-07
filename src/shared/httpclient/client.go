// Package httpclient: JSON over HTTP client for service to service calls
// non-2xx responses decode the shared error contract
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"IDIG4110/shared/httperror"
)

// Request timeout when New gets timeout <= 0
const DefaultTimeout = 10 * time.Second

// JSON over HTTP client, safe for concurrent use
type Client struct {
	http *http.Client
}

// Creates client with per request timeout
//
// # Inputs:
//
//   - timeout [time.Duration] per request timeout
//
// # Returns:
//
//   - New client, DefaultTimeout when timeout <= 0
func New(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Client{http: &http.Client{Timeout: timeout}}
}

// Error for every non-2xx response
type ResponseError struct {
	StatusCode int    // http status
	Code       int    // code from error contract, 0 when absent
	Message    string // message from error contract or status text
}

// Formats response error as "http <status>: <message>"
//
// # Returns:
//
//   - Formatted error string
func (e *ResponseError) Error() string {
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Message)
}

// Reports whether request is transient
//
// # Returns:
//
//   - True on 5xx or rate limiting, false on 4xx
func (e *ResponseError) Retryable() bool {
	return e.StatusCode >= 500 || e.StatusCode == http.StatusTooManyRequests
}

// Sends payload as JSON with POST, discards response body
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - url [string] target url
//   - payload [any] body, marshaled as JSON
//
// # Returns:
//
//   - Error on transport failure or non-2xx response
func (c *Client) PostJSON(ctx context.Context, url string, payload any) error {
	return c.Do(ctx, http.MethodPost, url, payload, nil)
}

// Sends GET and decodes 2xx response body into out
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - url [string] target url
//   - out [any] decoded response body
//
// # Returns:
//
//   - Error on transport failure, non-2xx response or decode failure
func (c *Client) GetJSON(ctx context.Context, url string, out any) error {
	return c.Do(ctx, http.MethodGet, url, nil, out)
}

// Sends one request
// non-nil payload becomes JSON body, 2xx body decoded into non-nil out
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - method [string] http method
//   - url [string] target url
//   - payload [any] JSON body, nil for none
//   - out [any] decoded response body, nil to discard
//
// # Returns:
//
//   - Error on failure or non-2xx response
func (c *Client) Do(ctx context.Context, method, url string, payload, out any) error {
	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		body = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err // transport error; classify with IsRetryable
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError(resp)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}

// Builds ResponseError from non-2xx response
//
// # Inputs:
//
//   - resp [*http.Response] non-2xx response
//
// # Returns:
//
//   - ResponseError with decoded error contract when present
func responseError(resp *http.Response) error {
	rErr := &ResponseError{
		StatusCode: resp.StatusCode,
		Message:    http.StatusText(resp.StatusCode),
	}
	var body httperror.ErrorMessage
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&body); err == nil && body.Message != "" {
		rErr.Code = body.Code
		rErr.Message = body.Message
	}
	return rErr
}

// Reports whether err is worth retrying
//
// # Inputs:
//
//   - err [error] error to inspect
//
// # Returns:
//
//   - False on 4xx rejection or canceled context, true otherwise
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var rErr *ResponseError
	if errors.As(err, &rErr) {
		return rErr.Retryable()
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	return true
}
