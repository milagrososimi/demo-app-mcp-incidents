package payments

import (
	"context"
	"testing"
	"time"
)

func TestPool_HandsOutUpToItsSize(t *testing.T) {
	t.Parallel()

	pool := NewPool(3)
	for i := range 3 {
		if err := pool.Acquire(context.Background()); err != nil {
			t.Fatalf("acquire %d: %v", i, err)
		}
	}
	if pool.InUse() != 3 {
		t.Fatalf("in use = %d, want 3", pool.InUse())
	}
}

func TestPool_WaitingPastTheDeadlineIsExhaustion(t *testing.T) {
	t.Parallel()

	pool := NewPool(1)
	if err := pool.Acquire(context.Background()); err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := pool.Acquire(ctx); err != ErrPoolExhausted {
		t.Fatalf("second acquire = %v, want %v", err, ErrPoolExhausted)
	}
}

func TestPool_ReleaseFreesASlot(t *testing.T) {
	t.Parallel()

	pool := NewPool(1)
	if err := pool.Acquire(context.Background()); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	pool.Release()

	if err := pool.Acquire(context.Background()); err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
}

func TestNewPool_SizeBelowOneIsRaised(t *testing.T) {
	t.Parallel()

	if got := NewPool(0).Size(); got != 1 {
		t.Fatalf("size = %d, want 1", got)
	}
}

func TestPool_SizeReportsWhatItWasBuiltWith(t *testing.T) {
	t.Parallel()

	if got := NewPool(12).Size(); got != 12 {
		t.Fatalf("size = %d, want 12", got)
	}
}
