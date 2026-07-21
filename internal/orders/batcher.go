// Package orders drains the order queue.
package orders

import (
	"context"
	"log/slog"
	"time"
)

// Order is one unit of work off the queue.
type Order struct {
	ID string
}

// Queue is the source a Batcher reads from.
type Queue interface {
	// Poll returns at most max orders, or nothing when the queue is empty.
	Poll(max int) []Order
}

// Batcher drains a queue, taking at most batchSize orders per poll.
//
// The round trip to the broker costs the same whether the batch comes back
// full or nearly empty.
type Batcher struct {
	queue     Queue
	batchSize int
	pollCost  time.Duration
	workCost  time.Duration
}

// NewBatcher builds the batcher. pollCost is the fixed round trip per poll;
// workCost is what one order costs to handle.
func NewBatcher(queue Queue, batchSize int, pollCost, workCost time.Duration) *Batcher {
	if batchSize < 1 {
		batchSize = 1
	}
	return &Batcher{queue: queue, batchSize: batchSize, pollCost: pollCost, workCost: workCost}
}

// Drain reads until the queue runs dry or ctx is done, and reports how many
// orders it handled.
func (b *Batcher) Drain(ctx context.Context) int {
	var handled int
	for {
		if err := ctx.Err(); err != nil {
			return handled
		}

		time.Sleep(b.pollCost)
		batch := b.queue.Poll(b.batchSize)
		if len(batch) == 0 {
			return handled
		}

		for range batch {
			time.Sleep(b.workCost)
			handled++
		}
		slog.Debug("drained a batch", "size", len(batch), "handled", handled)
	}
}

// BatchSize is how many orders one poll may carry.
func (b *Batcher) BatchSize() int { return b.batchSize }
