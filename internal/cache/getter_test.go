package cache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemoizedGetterCachesSuccessfulValuesByKey(t *testing.T) {
	underlying := &recordingGetter[int]{values: map[string]int{
		"key-1": 10,
		"key-2": 20,
	}}
	cache := NewMemoizedGetter(underlying.get)

	value, err := cache.Get(context.Background(), "key-1")
	require.NoError(t, err)
	require.Equal(t, 10, value)
	value, err = cache.Get(context.Background(), "key-1")
	require.NoError(t, err)
	require.Equal(t, 10, value)
	value, err = cache.Get(context.Background(), "key-2")
	require.NoError(t, err)
	require.Equal(t, 20, value)

	require.Equal(t, 1, underlying.callCount("key-1"))
	require.Equal(t, 1, underlying.callCount("key-2"))
}

func TestMemoizedGetterDoesNotCacheErrors(t *testing.T) {
	underlying := &recordingGetter[int]{
		values: map[string]int{"key": 42},
		errors: map[string][]error{
			"key": {errors.New("read failed"), nil},
		},
	}
	cache := NewMemoizedGetter(underlying.get)

	_, err := cache.Get(context.Background(), "key")
	require.EqualError(t, err, "read failed")
	value, err := cache.Get(context.Background(), "key")
	require.NoError(t, err)
	require.Equal(t, 42, value)

	require.Equal(t, 2, underlying.callCount("key"))
}

func TestMemoizedGetterCoalescesConcurrentMisses(t *testing.T) {
	underlying := &recordingGetter[int]{values: map[string]int{"key": 42}}
	underlying.block()
	cache := NewMemoizedGetter(underlying.get)

	const goroutineCount = 5
	var wg sync.WaitGroup
	errs := make(chan error, goroutineCount)
	values := make(chan int, goroutineCount)
	for range goroutineCount {
		wg.Add(1)
		go func() {
			defer wg.Done()

			value, err := cache.Get(context.Background(), "key")
			values <- value
			errs <- err
		}()
	}

	underlying.waitForBlockedCall(t)
	require.Eventually(t, func() bool {
		return underlying.callCount("key") == 1
	}, time.Second, 10*time.Millisecond)

	underlying.release()
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	close(values)
	for value := range values {
		require.Equal(t, 42, value)
	}
	require.Equal(t, 1, underlying.callCount("key"))
}

func TestMemoizedGetterWaiterCanCancelDuringConcurrentMiss(t *testing.T) {
	underlying := &recordingGetter[int]{values: map[string]int{"key": 42}}
	underlying.block()
	cache := NewMemoizedGetter(underlying.get)

	firstErr := make(chan error, 1)
	go func() {
		_, err := cache.Get(context.Background(), "key")
		firstErr <- err
	}()
	underlying.waitForBlockedCall(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := cache.Get(ctx, "key")
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, underlying.callCount("key"))

	underlying.release()
	require.NoError(t, <-firstErr)
}

type recordingGetter[V any] struct {
	mu sync.Mutex

	values map[string]V
	errors map[string][]error
	calls  map[string]int

	started     chan struct{}
	startedOnce sync.Once
	releaseCh   chan struct{}
}

func (getter *recordingGetter[V]) get(_ context.Context, key string) (V, error) {
	getter.mu.Lock()
	if getter.calls == nil {
		getter.calls = make(map[string]int)
	}
	getter.calls[key]++
	callIndex := getter.calls[key] - 1
	var err error
	if callIndex < len(getter.errors[key]) {
		err = getter.errors[key][callIndex]
	}
	value := getter.values[key]
	started := getter.started
	releaseCh := getter.releaseCh
	getter.mu.Unlock()

	if started != nil {
		getter.startedOnce.Do(func() {
			close(started)
		})
		<-releaseCh
	}
	if err != nil {
		var zero V
		return zero, err
	}

	return value, nil
}

func (getter *recordingGetter[V]) block() {
	getter.started = make(chan struct{})
	getter.releaseCh = make(chan struct{})
}

func (getter *recordingGetter[V]) waitForBlockedCall(t *testing.T) {
	t.Helper()

	select {
	case <-getter.started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for blocked getter call")
	}
}

func (getter *recordingGetter[V]) release() {
	close(getter.releaseCh)
}

func (getter *recordingGetter[V]) callCount(key string) int {
	getter.mu.Lock()
	defer getter.mu.Unlock()

	return getter.calls[key]
}
