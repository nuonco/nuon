package filecache

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FileCacheSuite struct {
	suite.Suite
	dir   string
	cache *FileCache
}

func TestFileCacheSuite(t *testing.T) {
	suite.Run(t, new(FileCacheSuite))
}

func (s *FileCacheSuite) SetupTest() {
	dir, err := os.MkdirTemp("", "filecache-test-*")
	require.NoError(s.T(), err)
	s.dir = dir

	cache, err := New(Options{
		Dir:      dir,
		MaxCount: 5,
		MaxBytes: 1024,
	})
	require.NoError(s.T(), err)
	s.cache = cache
}

func (s *FileCacheSuite) TearDownTest() {
	os.RemoveAll(s.dir)
}

func (s *FileCacheSuite) TestPutAndGet() {
	data := []byte("hello world")
	err := s.cache.Put("key1", data)
	require.NoError(s.T(), err)

	got, ok := s.cache.Get("key1")
	assert.True(s.T(), ok)
	assert.Equal(s.T(), data, got)
}

func (s *FileCacheSuite) TestGetMiss() {
	got, ok := s.cache.Get("nonexistent")
	assert.False(s.T(), ok)
	assert.Nil(s.T(), got)
}

func (s *FileCacheSuite) TestOverwrite() {
	err := s.cache.Put("key1", []byte("first"))
	require.NoError(s.T(), err)

	err = s.cache.Put("key1", []byte("second"))
	require.NoError(s.T(), err)

	got, ok := s.cache.Get("key1")
	assert.True(s.T(), ok)
	assert.Equal(s.T(), []byte("second"), got)
	assert.Equal(s.T(), 1, s.cache.Len())
}

func (s *FileCacheSuite) TestEvictionByCount() {
	// MaxCount is 5; insert 7 entries
	for i := 0; i < 7; i++ {
		err := s.cache.Put(fmt.Sprintf("key%d", i), []byte("x"))
		require.NoError(s.T(), err)
	}

	assert.Equal(s.T(), 5, s.cache.Len())

	// Oldest entries (key0, key1) should be evicted
	_, ok := s.cache.Get("key0")
	assert.False(s.T(), ok)
	_, ok = s.cache.Get("key1")
	assert.False(s.T(), ok)

	// Newest entries should still be present
	_, ok = s.cache.Get("key6")
	assert.True(s.T(), ok)
	_, ok = s.cache.Get("key5")
	assert.True(s.T(), ok)
}

func (s *FileCacheSuite) TestEvictionBySize() {
	// MaxBytes is 1024; insert entries that exceed it
	bigData := make([]byte, 300)
	for i := range bigData {
		bigData[i] = byte('a')
	}

	for i := 0; i < 5; i++ {
		err := s.cache.Put(fmt.Sprintf("key%d", i), bigData)
		require.NoError(s.T(), err)
	}

	// 5 * 300 = 1500 > 1024, so eviction should have happened
	assert.LessOrEqual(s.T(), s.cache.TotalSize(), int64(1024))
	assert.Less(s.T(), s.cache.Len(), 5)

	// Most recent should still be present
	_, ok := s.cache.Get("key4")
	assert.True(s.T(), ok)
}

func (s *FileCacheSuite) TestEvictionRemovesFiles() {
	for i := 0; i < 7; i++ {
		err := s.cache.Put(fmt.Sprintf("key%d", i), []byte("x"))
		require.NoError(s.T(), err)
	}

	// Evicted files should be deleted from disk
	_, err := os.Stat(filepath.Join(s.dir, "key0"))
	assert.True(s.T(), os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(s.dir, "key1"))
	assert.True(s.T(), os.IsNotExist(err))

	// Remaining files should exist on disk
	_, err = os.Stat(filepath.Join(s.dir, "key6"))
	assert.NoError(s.T(), err)
}

func (s *FileCacheSuite) TestLRUOrder() {
	// Fill cache to max count (5)
	for i := 0; i < 5; i++ {
		err := s.cache.Put(fmt.Sprintf("key%d", i), []byte("x"))
		require.NoError(s.T(), err)
	}

	// Access key0 to make it most recently used
	_, ok := s.cache.Get("key0")
	assert.True(s.T(), ok)

	// Insert a new entry to trigger eviction
	err := s.cache.Put("key5", []byte("x"))
	require.NoError(s.T(), err)

	// key1 should be evicted (oldest after key0 was promoted)
	_, ok = s.cache.Get("key1")
	assert.False(s.T(), ok)

	// key0 should survive (was promoted by Get)
	_, ok = s.cache.Get("key0")
	assert.True(s.T(), ok)
}

func (s *FileCacheSuite) TestExternalFileRemoval() {
	err := s.cache.Put("key1", []byte("data"))
	require.NoError(s.T(), err)

	// Externally delete the file
	os.Remove(filepath.Join(s.dir, "key1"))

	// Get should return miss and clean up tracking
	got, ok := s.cache.Get("key1")
	assert.False(s.T(), ok)
	assert.Nil(s.T(), got)
	assert.Equal(s.T(), 0, s.cache.Len())
}

func (s *FileCacheSuite) TestDirectoryCreation() {
	newDir := filepath.Join(s.dir, "nested", "subdir")
	cache, err := New(Options{
		Dir:      newDir,
		MaxCount: 10,
		MaxBytes: 4096,
	})
	require.NoError(s.T(), err)

	err = cache.Put("key1", []byte("data"))
	require.NoError(s.T(), err)

	got, ok := cache.Get("key1")
	assert.True(s.T(), ok)
	assert.Equal(s.T(), []byte("data"), got)
}

func (s *FileCacheSuite) TestNilSafety() {
	var nilCache *FileCache

	err := nilCache.Put("key1", []byte("data"))
	assert.NoError(s.T(), err)

	got, ok := nilCache.Get("key1")
	assert.False(s.T(), ok)
	assert.Nil(s.T(), got)

	assert.Equal(s.T(), 0, nilCache.Len())
	assert.Equal(s.T(), int64(0), nilCache.TotalSize())
}

func (s *FileCacheSuite) TestConcurrentAccess() {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i%10)
			data := []byte(fmt.Sprintf("data-%d", i))

			_ = s.cache.Put(key, data)
			s.cache.Get(key)
		}(i)
	}
	wg.Wait()

	// Cache should be in a consistent state
	assert.LessOrEqual(s.T(), s.cache.Len(), 5)
}

func (s *FileCacheSuite) TestAtomicWrite() {
	// Verify no .tmp files remain after Put
	err := s.cache.Put("key1", []byte("data"))
	require.NoError(s.T(), err)

	_, err = os.Stat(filepath.Join(s.dir, "key1.tmp"))
	assert.True(s.T(), os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(s.dir, "key1"))
	assert.NoError(s.T(), err)
}

func (s *FileCacheSuite) TestTotalSize() {
	err := s.cache.Put("key1", []byte("abc"))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), s.cache.TotalSize())

	err = s.cache.Put("key2", []byte("defgh"))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(8), s.cache.TotalSize())
}
