// Copyright (c) technicianted. All rights reserved.
// Licensed under the MIT License.
package bloque

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSingleOps(t *testing.T) {
	q := New()

	err := q.Push(context.Background(), 1)
	require.NoError(t, err)
	err = q.Push(context.Background(), 2)
	require.NoError(t, err)
	err = q.Push(context.Background(), 3)
	require.NoError(t, err)

	require.Equal(t, 3, q.Len())

	i, err := q.Pop(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, i)
	i, err = q.Pop(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, i)
	i, err = q.Pop(context.Background())
	require.NoError(t, err)
	require.Equal(t, 3, i)
}

func TestBlockingPop(t *testing.T) {
	q := New()

	poppers := 4
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	doneChan := make(chan error)
	for i := 0; i < poppers; i++ {
		go func(i int) {
			_, err := q.Pop(ctx)
			if err != nil {
				doneChan <- err
			}
			doneChan <- nil
		}(i)
	}

	for i := 0; i < poppers; i++ {
		time.Sleep(10 * time.Millisecond)
		err := q.Push(context.Background(), i)
		require.NoError(t, err)
		err = <-doneChan
		require.NoError(t, err)
	}

	require.Equal(t, 0, q.Len())
}

func TestBlockingPopTimeout(t *testing.T) {
	q := New()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := q.Pop(ctx)
	require.Error(t, err)
	require.Equal(t, context.DeadlineExceeded, err)
}

func TestBlockingPushTimeout(t *testing.T) {
	q := New(WithCapacity(1))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := q.Push(ctx, 1)
	require.NoError(t, err)
	err = q.Push(ctx, 2)
	require.Error(t, err)
	require.Equal(t, context.DeadlineExceeded, err)
}

func TestMaxPushWaiters(t *testing.T) {
	q := New(WithCapacity(1), WithMaxPushWaiters(1))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	// fill the queue
	err := q.Push(ctx, 1)
	require.NoError(t, err)
	// will block
	startedChan := make(chan interface{})
	go func() {
		close(startedChan)
		q.Push(ctx, 2)
	}()
	<-startedChan
	time.Sleep(10 * time.Millisecond)
	// will fail
	err = q.Push(ctx, 3)
	require.Error(t, err)
	require.Equal(t, ErrMaxWaiters, err)
}

func TestMaxPopWaiters(t *testing.T) {
	q := New(WithMaxPopWaiters(1))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	// will block
	startedChan := make(chan interface{})
	go func() {
		close(startedChan)
		q.Pop(ctx)
	}()
	<-startedChan
	time.Sleep(10 * time.Millisecond)
	// will fail
	_, err := q.Pop(ctx)
	require.Error(t, err)
	require.Equal(t, ErrMaxWaiters, err)
}
