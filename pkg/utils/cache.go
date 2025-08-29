package utils

import "container/list"

// LRUCache is a simple fixed-size LRU cache.
type LRUCache[K comparable, V any] struct {
	capacity int
	items    map[K]*list.Element
	order    *list.List
}

type lruEntry[K comparable, V any] struct {
	key   K
	value V
}

// NewLRUCache creates LRU cache with given capacity; if capacity <=0, cache is unbounded.
func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
		order:    list.New(),
	}
}

// Get returns value by key and marks it as most recently used.
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		return el.Value.(lruEntry[K, V]).value, true
	}
	var zero V
	return zero, false
}

// Set saves value for key, evicting oldest item if needed.
func (c *LRUCache[K, V]) Set(key K, value V) {
	if el, ok := c.items[key]; ok {
		el.Value = lruEntry[K, V]{key: key, value: value}
		c.order.MoveToFront(el)
		return
	}
	el := c.order.PushFront(lruEntry[K, V]{key: key, value: value})
	c.items[key] = el
	if c.capacity > 0 && c.order.Len() > c.capacity {
		last := c.order.Back()
		if last != nil {
			c.order.Remove(last)
			kv := last.Value.(lruEntry[K, V])
			delete(c.items, kv.key)
		}
	}
}

// MemoizedWithSingleArgument returns memoized version of function using provided cache map.
func MemoizedWithSingleArgument[K comparable, V any](cache map[K]V, fn func(K) V) func(K) V {
	return func(arg K) V {
		if val, ok := cache[arg]; ok {
			return val
		}
		res := fn(arg)
		cache[arg] = res
		return res
	}
}

// MemoizeWithLRU wraps function with memoization backed by LRUCache.
func MemoizeWithLRU[K comparable, V any](cache *LRUCache[K, V], fn func(K) V) func(K) V {
	return func(arg K) V {
		if val, ok := cache.Get(arg); ok {
			return val
		}
		res := fn(arg)
		cache.Set(arg, res)
		return res
	}
}
