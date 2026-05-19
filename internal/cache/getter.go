package cache

import (
	"context"
	"sync"

	"golang.org/x/sync/singleflight"
)

type MemoizedGetter[V any] struct {
	get   func(context.Context, string) (V, error)
	group singleflight.Group

	mu    sync.RWMutex
	items map[string]V
}

func NewMemoizedGetter[V any](get func(context.Context, string) (V, error)) *MemoizedGetter[V] {
	return &MemoizedGetter[V]{
		get:   get,
		items: make(map[string]V),
	}
}

func (cache *MemoizedGetter[V]) Get(ctx context.Context, key string) (V, error) {
	if value, ok := cache.lookup(key); ok {
		return value, nil
	}

	resultCh := cache.group.DoChan(key, func() (any, error) {
		if value, ok := cache.lookup(key); ok {
			return value, nil
		}

		value, err := cache.get(ctx, key)
		if err != nil {
			return nil, err
		}

		cache.store(key, value)
		return value, nil
	})

	select {
	case result := <-resultCh:
		if result.Err != nil {
			var zero V
			return zero, result.Err
		}

		return result.Val.(V), nil
	case <-ctx.Done():
		var zero V
		return zero, ctx.Err()
	}
}

func (cache *MemoizedGetter[V]) lookup(key string) (V, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	value, ok := cache.items[key]
	return value, ok
}

func (cache *MemoizedGetter[V]) store(key string, value V) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.items[key] = value
}
