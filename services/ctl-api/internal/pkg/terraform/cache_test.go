package terraform

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCache(ttl time.Duration) Cache {
	return Cache{
		ttl:   ttl,
		store: make(map[string]CacheEntry),
	}
}

func TestCache_MissOnEmptyCache(t *testing.T) {
	c := newTestCache(time.Minute)
	_, ok := c.Get("missing")
	assert.False(t, ok)
}

func TestCache_HitAfterSet(t *testing.T) {
	c := newTestCache(time.Minute)
	c.Set("key", "value")
	got, ok := c.Get("key")
	require.True(t, ok)
	assert.Equal(t, "value", got)
}

func TestCache_MissAfterExpiry(t *testing.T) {
	c := newTestCache(1 * time.Millisecond)
	c.Set("key", "value")
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("key")
	assert.False(t, ok)
}

func TestCache_EntryRemovedAfterExpiry(t *testing.T) {
	c := newTestCache(1 * time.Millisecond)
	c.Set("key", "value")
	time.Sleep(5 * time.Millisecond)
	c.Get("key") // triggers deletion
	c.mu.RLock()
	_, exists := c.store["key"]
	c.mu.RUnlock()
	assert.False(t, exists)
}

func TestCache_OverwriteValue(t *testing.T) {
	c := newTestCache(time.Minute)
	c.Set("key", "first")
	c.Set("key", "second")
	got, ok := c.Get("key")
	require.True(t, ok)
	assert.Equal(t, "second", got)
}

func TestCache_MultipleKeys(t *testing.T) {
	c := newTestCache(time.Minute)
	c.Set("a", "1")
	c.Set("b", "2")

	v, ok := c.Get("a")
	require.True(t, ok)
	assert.Equal(t, "1", v)

	v, ok = c.Get("b")
	require.True(t, ok)
	assert.Equal(t, "2", v)
}
