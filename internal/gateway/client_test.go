package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// processor stands up an upstream that answers after latency with status.
func processor(t *testing.T, latency time.Duration, status int) (url string, calls *int) {
	t.Helper()

	var seen int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		seen++
		time.Sleep(latency)
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)
	return server.URL, &seen
}

func TestClient_Authorizes(t *testing.T) {
	t.Parallel()

	url, calls := processor(t, 0, http.StatusOK)

	if err := New(url, time.Second, 2).Authorize(context.Background(), "ord_1"); err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if *calls != 1 {
		t.Fatalf("upstream saw %d calls, want 1", *calls)
	}
}

func TestClient_RetriesUntilTheAttemptsRunOut(t *testing.T) {
	t.Parallel()

	url, calls := processor(t, 0, http.StatusInternalServerError)

	if err := New(url, time.Second, 2).Authorize(context.Background(), "ord_2"); err == nil {
		t.Fatal("authorize succeeded against a failing upstream")
	}
	if *calls != 3 {
		t.Fatalf("upstream saw %d calls, want 3 (the first plus two retries)", *calls)
	}
}

func TestClient_TimeoutAppliesToEachAttempt(t *testing.T) {
	t.Parallel()

	url, calls := processor(t, 100*time.Millisecond, http.StatusOK)

	if err := New(url, 10*time.Millisecond, 1).Authorize(context.Background(), "ord_3"); err == nil {
		t.Fatal("authorize succeeded past its timeout")
	}
	if *calls != 2 {
		t.Fatalf("upstream saw %d calls, want 2", *calls)
	}
}

func TestClient_TimeoutReportsWhatItWasBuiltWith(t *testing.T) {
	t.Parallel()

	if got := New("http://example.invalid", 3*time.Second, 0).Timeout(); got != 3*time.Second {
		t.Fatalf("timeout = %v, want 3s", got)
	}
}
