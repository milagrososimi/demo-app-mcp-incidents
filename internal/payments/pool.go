// Package payments serves the payments API and owns its database pool.
package payments

import (
	"context"
	"errors"
)

// ErrPoolExhausted reports that no connection came free before the caller gave
// up waiting.
var ErrPoolExhausted = errors.New("payments: connection pool exhausted")

// Pool bounds how many queries run against the database at once.
//
// It is a semaphore rather than a driver pool: nothing here talks to a real
// database, and the bound is the part that matters to callers.
type Pool struct {
	slots chan struct{}
}

// NewPool returns a pool of the given size. A size below one is raised to one:
// a pool that hands out nothing is a stopped service, not a small one.
func NewPool(size int) *Pool {
	if size < 1 {
		size = 1
	}
	return &Pool{slots: make(chan struct{}, size)}
}

// Acquire takes a connection, waiting until one frees up or ctx is done.
func (p *Pool) Acquire(ctx context.Context) error {
	select {
	case p.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ErrPoolExhausted
	}
}

// Release returns a connection. It must follow a successful Acquire.
func (p *Pool) Release() {
	select {
	case <-p.slots:
	default:
	}
}

// Size is how many connections the pool can hand out at once.
func (p *Pool) Size() int { return cap(p.slots) }

// InUse is how many are out right now.
func (p *Pool) InUse() int { return len(p.slots) }
