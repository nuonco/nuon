package terraform

import (
	"sync"
	"time"
)

type CacheEntry struct {
	Value     string
	Timestamp time.Time
}

type Cache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	store map[string]CacheEntry
}

func (cache *Cache) Get(key string) (string, bool) {
	cache.mu.RLock()
	entry, ok := cache.store[key]
	cache.mu.RUnlock()

	if !ok {
		return "", false
	}

	if time.Since(entry.Timestamp) > cache.ttl {
		cache.mu.Lock()
		delete(cache.store, key)
		cache.mu.Unlock()
		return "", false
	}

	return entry.Value, true
}

func (cache *Cache) Set(key string, value string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.store[key] = CacheEntry{
		Value:     value,
		Timestamp: time.Now(),
	}
}
