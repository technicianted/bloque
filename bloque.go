// Copyright (c) technicianted. All rights reserved.
// Licensed under the MIT License.
package bloque

import (
	"container/list"
	"context"
	"fmt"
	"sync"
)

var (
	// ErrMaxWaiters is returned when maximum number of blocked goroutines
	// on Push() or Pop() calls is reached.
	ErrMaxWaiters = fmt.Errorf("max waiters reached")
)

// Bloque is a simple implementation of a blocking fifo queue. It allows
// various constrains to be specified such as maximum capacity, maximum waiters
// and so on.
type Bloque struct {
	itemQueue      *list.List
	capacity       int
	maxPushWaiters int
	maxPopWaiters  int
	mutex          sync.Mutex

	pushWaitersList *list.List
	popWaitersList  *list.List
}

// waiter is used to represent a waiting call.
type waiter struct {
	// waitChan is used to notify the waiter that the required condition
	// is met by closing it.
	waitChan chan interface{}
	// waiting flags that the waiter is still interested in the notification.
	// If the waiting call is canceled, it should set this flag to false.
	waiting bool
	// fired indicates that condition is met and waitChan is closed.
	// It is used by the caller when the context is canceled to determine if
	// someone else should handle the notification.
	fired bool
	mutex sync.Mutex
}

// New creates a new Bloque with opts.
func New(opts ...optionFunc) *Bloque {
	b := &Bloque{
		itemQueue:       list.New(),
		pushWaitersList: list.New(),
		popWaitersList:  list.New(),
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Push puts item at the back of the queue. If queue capacity is reached
// (as specified by WithCapacity() option), the call will block until
// either an item is removed from the queue or ctx is cancelled.
// Returns ErrMaxWaiters if maximum number of waiting goroutines is reached
// as specified by WithMaxPushWaiters() option.
func (q *Bloque) Push(ctx context.Context, item interface{}) error {
	q.mutex.Lock()

	// similar to condition variables spurious wake ups where our
	// blocked goroutine would be notified but someone else beat us
	// to the item.
	for q.capacity > 0 && q.itemQueue.Len() >= q.capacity {
		// straight up, do not exceed waiters constrain
		if q.maxPushWaiters > 0 && q.pushWaitersList.Len() >= q.maxPushWaiters {
			q.mutex.Unlock()
			return ErrMaxWaiters
		}

		waiterItem := &waiter{
			waitChan: make(chan interface{}),
			waiting:  true,
		}
		q.pushWaitersList.PushBack(waiterItem)
		q.mutex.Unlock()

		select {
		case <-waiterItem.waitChan:
			q.mutex.Lock()
			continue
		case <-ctx.Done():
			waiterItem.mutex.Lock()
			waiterItem.waiting = false
			if waiterItem.fired {
				// race detected, must pass on to the next waiter
				waiterItem.mutex.Unlock()
				q.mutex.Lock()
				q.unblockNextWaiterLocked(q.pushWaitersList)
				q.mutex.Unlock()
			} else {
				waiterItem.mutex.Unlock()
			}
			return ctx.Err()
		}
	}

	q.itemQueue.PushBack(item)
	q.unblockNextWaiterLocked(q.popWaitersList)
	q.mutex.Unlock()
	return nil
}

// Pop gets an item at the front of the queue. If queue is empty the call
// will block until either an item is available on the queue or ctx is cancelled.
// Returns ErrMaxWaiters if maximum number of waiting goroutines is reached
// as specified by WithMaxPopWaiters() option.
func (q *Bloque) Pop(ctx context.Context) (item interface{}, err error) {
	q.mutex.Lock()

	for q.itemQueue.Len() == 0 {
		if q.maxPopWaiters > 0 && q.popWaitersList.Len() >= q.maxPopWaiters {
			q.mutex.Unlock()
			return nil, ErrMaxWaiters
		}

		waiterItem := &waiter{
			waitChan: make(chan interface{}),
			waiting:  true,
		}
		q.popWaitersList.PushBack(waiterItem)
		q.mutex.Unlock()

		select {
		case <-waiterItem.waitChan:
			q.mutex.Lock()
			continue
		case <-ctx.Done():
			waiterItem.mutex.Lock()
			waiterItem.waiting = false
			if waiterItem.fired {
				// race detected, must pass on to the next waiter
				waiterItem.mutex.Unlock()
				q.mutex.Lock()
				q.unblockNextWaiterLocked(q.popWaitersList)
				q.mutex.Unlock()
			} else {
				waiterItem.mutex.Unlock()
			}
			return nil, ctx.Err()
		}
	}

	el := q.itemQueue.Front()
	val := q.itemQueue.Remove(el)
	q.unblockNextWaiterLocked(q.pushWaitersList)
	q.mutex.Unlock()

	return val, nil
}

// Len returns the current length of the queue.
func (q *Bloque) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.itemQueue.Len()
}

func (q *Bloque) unblockNextWaiterLocked(waiters *list.List) {
	for waiters.Len() > 0 {
		el := waiters.Front()
		w := waiters.Remove(el).(*waiter)
		w.mutex.Lock()
		if w.waiting {
			close(w.waitChan)
			w.fired = true
			w.mutex.Unlock()
			return
		}
	}
}
