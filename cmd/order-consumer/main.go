// Command order-consumer drains the order queue in batches.
//
// Settings come from the environment; deploy/order-consumer.yaml is what sets
// them in production.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/milagrososimi/demo-app-mcp-incidents/internal/config"
	"github.com/milagrososimi/demo-app-mcp-incidents/internal/orders"
)

// backlog stands in for the broker, so the consumer can be run on its own.
type backlog struct{ remaining int }

func (b *backlog) Poll(max int) []orders.Order {
	if b.remaining <= 0 {
		return nil
	}
	if max > b.remaining {
		max = b.remaining
	}
	b.remaining -= max

	batch := make([]orders.Order, 0, max)
	for i := range max {
		batch = append(batch, orders.Order{ID: string(rune('a' + i%26))})
	}
	return batch
}

// version is stamped at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	var (
		batchSize = config.Int("CONSUMER_BATCH_SIZE", 500)
		pollCost  = config.Duration("CONSUMER_POLL_COST", 40*time.Millisecond)
		workCost  = config.Duration("CONSUMER_WORK_COST", 100*time.Microsecond)
		queued    = config.Int("CONSUMER_BACKLOG", 5000)
	)

	batcher := orders.NewBatcher(&backlog{remaining: queued}, batchSize, pollCost, workCost)
	slog.Info("starting order-consumer", "version", version)

	// A drain in progress should finish its batch, not die mid-flight.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	started := time.Now()
	handled := batcher.Drain(ctx)
	elapsed := time.Since(started)

	slog.Info("drained the backlog",
		"handled", handled,
		"elapsed", elapsed.Round(time.Millisecond),
		"orders_per_second", int(float64(handled)/elapsed.Seconds()))
}
