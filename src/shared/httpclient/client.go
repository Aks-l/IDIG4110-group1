// Package httpclient is a JSON-over-HTTP client for service-to-service
// calls. Non-2xx responses decode the shared {"code","message"} error
// contract (see shared/httperror), so a caller can tell a permanent 4xx
// rejection from a transient 5xx.
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

// DefaultTimeout bounds one request when New is called without a timeout.
const DefaultTimeout = 10 * time.Second

// Client is a JSON-over-HTTP client. It is safe for concurrent use.
type Client struct {
	http *http.Client
}

// New returns a Client with the given per-request timeout; timeout <= 0
// falls back to DefaultTimeout.
func New(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Client{http: &http.Client{Timeout: timeout}}
}

// ResponseError is returned for every non-2xx response. Code and Message
// come from the shared {"code","message"} error contract when the server
// speaks it; otherwise Message falls back to the HTTP status text.
type ResponseError struct {
	StatusCode int    // HTTP status code of the response
	Code       int    // code echoed from the error-contract body, 0 when absent
	Message    string // message from the error-contract body, or the status text
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Message)
}

// Retryable reports whether the request may succeed if retried later:
// server errors and rate limiting are transient, client errors are not.
func (e *ResponseError) Retryable() bool {
	return e.StatusCode >= 500 || e.StatusCode == http.StatusTooManyRequests
}

// PostJSON sends payload as JSON with POST and expects a 2xx response. Use
// Do when the response body matters.
func (c *Client) PostJSON(ctx context.Context, url string, payload any) error {
	return c.Do(ctx, http.MethodPost, url, payload, nil)
}

// GetJSON sends GET and decodes a 2xx response body into out.
func (c *Client) GetJSON(ctx context.Context, url string, out any) error {
	return c.Do(ctx, http.MethodGet, url, nil, out)
}

// Do sends one request: a non-nil payload becomes the JSON body and a 2xx
// body is decoded into a non-nil out.
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

// responseError turns a non-2xx response into a ResponseError, decoding
// the shared error contract when the server speaks it.
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

// IsRetryable reports whether err is worth retrying: a 4xx rejection is
// permanent, server errors, rate limiting, and transport failures may
// succeed later, and a canceled context never retries.
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
