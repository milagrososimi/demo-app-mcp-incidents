package payments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func serve(t *testing.T, h *Handler) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.Handle("GET /payments/{id}", h)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestHandler_ReturnsTheLookup(t *testing.T) {
	t.Parallel()

	server := serve(t, NewHandler(NewPool(4), time.Millisecond, time.Second))

	resp, err := server.Client().Get(server.URL + "/payments/pay_7")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
		PoolSize  int    `json:"pool_size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.PaymentID != "pay_7" || body.Status != "settled" || body.PoolSize != 4 {
		t.Fatalf("body = %+v", body)
	}
}

func TestHandler_ReturnsServiceUnavailableWhenNoConnectionIsFree(t *testing.T) {
	t.Parallel()

	pool := NewPool(1)
	if err := pool.Acquire(context.Background()); err != nil {
		t.Fatalf("hold the only connection: %v", err)
	}
	defer pool.Release()

	server := serve(t, NewHandler(pool, time.Millisecond, 20*time.Millisecond))

	resp, err := server.Client().Get(server.URL + "/payments/pay_8")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
}
