// Package gateway calls the upstream payment processor.
package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client authorizes payments against the upstream processor.
//
// The timeout is per attempt, not per authorization: a retry starts its own,
// so the longest an authorization can take is the timeout times the number of
// attempts it is allowed.
type Client struct {
	http    *http.Client
	baseURL string
	retries int
}

// New builds a client. timeout bounds one attempt; retries is how many extra
// attempts an authorization gets after the first.
func New(baseURL string, timeout time.Duration, retries int) *Client {
	return &Client{
		http:    &http.Client{Timeout: timeout},
		baseURL: baseURL,
		retries: retries,
	}
}

// Timeout is how long one attempt may take.
func (c *Client) Timeout() time.Duration { return c.http.Timeout }

// Authorize asks the processor to authorize an order, retrying on failure.
func (c *Client) Authorize(ctx context.Context, orderID string) error {
	var last error
	for attempt := 0; attempt <= c.retries; attempt++ {
		err := c.attempt(ctx, orderID)
		if err == nil {
			return nil
		}
		last = err
		if ctx.Err() != nil {
			break
		}
	}
	return fmt.Errorf("gateway: authorize %s: %w", orderID, last)
}

func (c *Client) attempt(ctx context.Context, orderID string) error {
	url := fmt.Sprintf("%s/authorize/%s", c.baseURL, orderID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
