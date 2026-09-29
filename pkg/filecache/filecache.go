package filecache

import (
	"container/list"
	"os"
	"path/filepath"
	"sync"
)

type Options struct {
	Dir string

	MaxCount int

	MaxBytes int64
}

type cacheEntry struct {
	id      string
	size    int64
	element *list.Element
}

type FileCache struct {
	mu        sync.Mutex
	dir       string
	maxCount  int
	maxBytes  int64
	entries   map[string]*cacheEntry
	order     *list.List
	totalSize int64
}

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

func (c *FileCache) Has(id string) bool {
	if c == nil {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[id]
	if ok {
		c.order.MoveToFront(entry.element)
	}
	return ok
}

func (c *FileCache) Len() int {
	if c == nil {
		return 0
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

func (c *FileCache) TotalSize() int64 {
	if c == nil {
		return 0
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalSize
}

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
