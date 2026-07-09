package payments

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Handler answers one payment lookup per request, holding a pooled connection
// for as long as the query runs.
type Handler struct {
	pool  *Pool
	query time.Duration
	wait  time.Duration
}

// NewHandler builds the handler. query is how long a lookup occupies its
// connection; wait is how long a request will queue for one before giving up.
func NewHandler(pool *Pool, query, wait time.Duration) *Handler {
	return &Handler{pool: pool, query: query, wait: wait}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The wait bounds queueing only. Once a request holds a connection its
	// query runs to completion: cutting a query short would leave the database
	// doing the work anyway, and free the slot no sooner.
	queueCtx, cancel := context.WithTimeout(r.Context(), h.wait)
	defer cancel()

	if err := h.pool.Acquire(queueCtx); err != nil {
		http.Error(w, "database busy", http.StatusServiceUnavailable)
		return
	}
	defer h.pool.Release()

	select {
	case <-time.After(h.query):
	case <-r.Context().Done():
		http.Error(w, "client went away", http.StatusRequestTimeout)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"payment_id":  r.PathValue("id"),
		"status":      "settled",
		"pool_size":   h.pool.Size(),
		"pool_in_use": h.pool.InUse(),
	})
}
