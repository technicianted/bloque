// Copyright (c) technicianted. All rights reserved.
// Licensed under the MIT License.
package bloque

type optionFunc func(*Bloque)

// WithCapacity sets maximum queue capacity to capacity. A value of 0
// means unlimited capacity.
// When the queue reaches capacity, Push() is going to block until
// items are removed from the queue.
// You can use WithMaxPushWaiters() to set maximum number of blocked
// Push() goroutines.
// Default is unlimited capacity.
func WithCapacity(capacity int) optionFunc {
	return func(b *Bloque) {
		b.capacity = capacity
	}
}

// WithMaxPushWaiters sets maximum maxWaiters of goroutine calls blocked on
// Push() calls. Once the limit is reached, ErrMaxWaiters is returned.
// Default is unlimited waiters.
func WithMaxPushWaiters(maxWaiters int) optionFunc {
	return func(b *Bloque) {
		b.maxPushWaiters = maxWaiters
	}
}

// WithMaxPopWaiters sets maximum maxWaiters of goroutine calls blocked on
// Pop() calls. Once the limit is reached, ErrMaxWaiters is returned.
// Default is unlimited waiters.
func WithMaxPopWaiters(maxWaiters int) optionFunc {
	return func(b *Bloque) {
		b.maxPopWaiters = maxWaiters
	}
}
