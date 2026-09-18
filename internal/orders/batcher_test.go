package orders

import (
	"context"
	"testing"
)

type fixedQueue struct {
	remaining int
	polls     []int
}

func (q *fixedQueue) Poll(max int) []Order {
	if q.remaining <= 0 {
		return nil
	}
	if max > q.remaining {
		max = q.remaining
	}
	q.remaining -= max
	q.polls = append(q.polls, max)
	return make([]Order, max)
}

func TestBatcher_DrainsEverythingQueued(t *testing.T) {
	t.Parallel()

	if got := NewBatcher(&fixedQueue{remaining: 250}, 100, 0, 0).Drain(context.Background()); got != 250 {
		t.Fatalf("handled = %d, want 250", got)
	}
}

func TestBatcher_NeverPollsForMoreThanTheBatchSize(t *testing.T) {
	t.Parallel()

	queue := &fixedQueue{remaining: 250}
	NewBatcher(queue, 100, 0, 0).Drain(context.Background())

	for i, size := range queue.polls {
		if size > 100 {
			t.Fatalf("poll %d asked for %d, want at most 100", i, size)
		}
	}
	if len(queue.polls) != 3 {
		t.Fatalf("polls = %v, want three", queue.polls)
	}
}

func TestBatcher_StopsWhenTheContextIsDone(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if got := NewBatcher(&fixedQueue{remaining: 100}, 10, 0, 0).Drain(ctx); got != 0 {
		t.Fatalf("handled = %d on a cancelled context, want 0", got)
	}
}

func TestNewBatcher_SizeBelowOneIsRaised(t *testing.T) {
	t.Parallel()

	if got := NewBatcher(&fixedQueue{}, 0, 0, 0).BatchSize(); got != 1 {
		t.Fatalf("batch size = %d, want 1", got)
	}
}

func TestBatcher_HandlesAPartialFinalBatch(t *testing.T) {
	t.Parallel()

	queue := &fixedQueue{remaining: 120}
	if got := NewBatcher(queue, 50, 0, 0).Drain(context.Background()); got != 120 {
		t.Fatalf("handled = %d, want 120", got)
	}
	if last := queue.polls[len(queue.polls)-1]; last != 20 {
		t.Fatalf("final poll = %d, want the 20 left over", last)
	}
}
