package filecache

import (
	"container/list"
	"os"
	"path/filepath"
	"sync"
)

type Options struct {
	// Dir is the directory where cache files are stored.
	// Created if it does not exist.
	Dir string

	// MaxCount is the maximum number of files in the cache.
	// When exceeded, the least recently used files are evicted.
	MaxCount int

	// MaxBytes is the maximum total size in bytes of all cached files.
	// When exceeded, the least recently used files are evicted.
	MaxBytes int64
}

type cacheEntry struct {
	id      string
	size    int64
	element *list.Element
}

// FileCache is a file-based LRU cache. It tracks files in memory and
// evicts the least recently used when count or total size limits are exceeded.
type FileCache struct {
	mu        sync.Mutex
	dir       string
	maxCount  int
	maxBytes  int64
	entries   map[string]*cacheEntry
	order     *list.List // front = most recently used
	totalSize int64
}

// New creates a new file-based LRU cache. The directory is created
// if it does not exist.
func New(opts Options) (*FileCache, error) {
	if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
		return nil, err
	}

	fc := &FileCache{
		dir:      opts.Dir,
		maxCount: opts.MaxCount,
		maxBytes: opts.MaxBytes,
		entries:  make(map[string]*cacheEntry),
		order:    list.New(),
	}

	// Hydrate from existing files on disk so cached data survives restarts.
	dirEntries, _ := os.ReadDir(opts.Dir)
	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		id := de.Name()
		elem := fc.order.PushBack(id)
		fc.entries[id] = &cacheEntry{
			id:      id,
			size:    info.Size(),
			element: elem,
		}
		fc.totalSize += info.Size()
	}
	fc.evict()

	return fc, nil
}

// Put writes data to the cache atomically and evicts old entries if needed.
// Nil-safe: calling Put on a nil *FileCache is a no-op.
func (c *FileCache) Put(id string, data []byte) error {
	if c == nil {
		return nil
	}

	path := filepath.Join(c.dir, id)
	tmpPath := path + ".tmp"

	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	size := int64(len(data))

	// If entry already exists, remove it first so we can re-add at front
	if existing, ok := c.entries[id]; ok {
		c.totalSize -= existing.size
		c.order.Remove(existing.element)
		delete(c.entries, id)
	}

	elem := c.order.PushFront(id)
	c.entries[id] = &cacheEntry{
		id:      id,
		size:    size,
		element: elem,
	}
	c.totalSize += size

	c.evict()
	return nil
}

// Get reads data from the cache and promotes the entry to most recently used.
// Returns (nil, false) on cache miss or read error.
// Nil-safe: calling Get on a nil *FileCache always returns (nil, false).
func (c *FileCache) Get(id string) ([]byte, bool) {
	if c == nil {
		return nil, false
	}

	c.mu.Lock()
	entry, ok := c.entries[id]
	if ok {
		c.order.MoveToFront(entry.element)
	}
	c.mu.Unlock()

	if !ok {
		return nil, false
	}

	data, err := os.ReadFile(filepath.Join(c.dir, id))
	if err != nil {
		// File was removed externally; clean up tracking
		c.mu.Lock()
		if e, exists := c.entries[id]; exists {
			c.totalSize -= e.size
			c.order.Remove(e.element)
			delete(c.entries, id)
		}
		c.mu.Unlock()
		return nil, false
	}

	return data, true
}

// Len returns the number of entries currently in the cache.
// Nil-safe: calling Len on a nil *FileCache returns 0.
func (c *FileCache) Len() int {
	if c == nil {
		return 0
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

// TotalSize returns the total size in bytes of all cached entries.
// Nil-safe: calling TotalSize on a nil *FileCache returns 0.
func (c *FileCache) TotalSize() int64 {
	if c == nil {
		return 0
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalSize
}

// evict removes the least recently used entries until both count and size
// limits are satisfied. Must be called with c.mu held.
func (c *FileCache) evict() {
	for c.order.Len() > c.maxCount || (c.maxBytes > 0 && c.totalSize > c.maxBytes) {
		back := c.order.Back()
		if back == nil {
			break
		}

		id := back.Value.(string)
		entry := c.entries[id]

		os.Remove(filepath.Join(c.dir, id))
		c.totalSize -= entry.size
		c.order.Remove(back)
		delete(c.entries, id)
	}
}
