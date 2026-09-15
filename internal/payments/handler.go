package payments

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// lookup is the body a successful lookup returns.
type lookup struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	PoolSize  int    `json:"pool_size"`
	PoolInUse int    `json:"pool_in_use"`
}

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
	started := time.Now()
	id := r.PathValue("id")

	// The wait bounds queueing only. Once a request holds a connection its
	// query runs to completion: cutting a query short would leave the database
	// doing the work anyway, and free the slot no sooner.
	queueCtx, cancel := context.WithTimeout(r.Context(), h.wait)
	defer cancel()

	if err := h.pool.Acquire(queueCtx); err != nil {
		h.log(id, http.StatusServiceUnavailable, started, time.Since(started))
		http.Error(w, "database busy", http.StatusServiceUnavailable)
		return
	}
	defer h.pool.Release()
	queued := time.Since(started)

	select {
	case <-time.After(h.query):
	case <-r.Context().Done():
		h.log(id, http.StatusRequestTimeout, started, queued)
		http.Error(w, "client went away", http.StatusRequestTimeout)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lookup{
		PaymentID: id,
		Status:    "settled",
		PoolSize:  h.pool.Size(),
		PoolInUse: h.pool.InUse(),
	})
	h.log(id, http.StatusOK, started, queued)
}

// log records one request. The queue wait is separated from the total because
// they move for different reasons: the query is what it is, and the wait is
// what the pool and the offered load make it.
func (h *Handler) log(id string, status int, started time.Time, queued time.Duration) {
	slog.Info("payment lookup",
		"payment_id", id,
		"status", status,
		"duration_ms", time.Since(started).Milliseconds(),
		"pool_wait_ms", queued.Milliseconds(),
		"pool_in_use", h.pool.InUse())
}
