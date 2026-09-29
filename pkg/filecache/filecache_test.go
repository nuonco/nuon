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

func (s *FileCacheSuite) TestHas() {
	err := s.cache.Put("key1", []byte("data"))
	require.NoError(s.T(), err)

	assert.True(s.T(), s.cache.Has("key1"))
	assert.False(s.T(), s.cache.Has("nonexistent"))
}

func (s *FileCacheSuite) TestHasPromotesLRU() {
	for i := 0; i < 5; i++ {
		err := s.cache.Put(fmt.Sprintf("key%d", i), []byte("x"))
		require.NoError(s.T(), err)
	}

	assert.True(s.T(), s.cache.Has("key0"))
	err := s.cache.Put("key5", []byte("x"))
	require.NoError(s.T(), err)

	assert.False(s.T(), s.cache.Has("key1"))
	assert.True(s.T(), s.cache.Has("key0"))
}

func (s *FileCacheSuite) TestHasNilSafety() {
	var nilCache *FileCache
	assert.False(s.T(), nilCache.Has("key1"))
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
	for i := 0; i < 7; i++ {
		err := s.cache.Put(fmt.Sprintf("key%d", i), []byte("x"))
		require.NoError(s.T(), err)
	}

	assert.Equal(s.T(), 5, s.cache.Len())

	_, ok := s.cache.Get("key0")
	assert.False(s.T(), ok)
	_, ok = s.cache.Get("key1")
	assert.False(s.T(), ok)

	_, ok = s.cache.Get("key6")
	assert.True(s.T(), ok)
	_, ok = s.cache.Get("key5")
	assert.True(s.T(), ok)
}

func (s *FileCacheSuite) TestEvictionBySize() {
	bigData := make([]byte, 300)
	for i := range bigData {
		bigData[i] = byte('a')
	}

	for i := 0; i < 5; i++ {
		err := s.cache.Put(fmt.Sprintf("key%d", i), bigData)
		require.NoError(s.T(), err)
	}

	assert.LessOrEqual(s.T(), s.cache.TotalSize(), int64(1024))
	assert.Less(s.T(), s.cache.Len(), 5)

	_, ok := s.cache.Get("key4")
	assert.True(s.T(), ok)
}

func (s *FileCacheSuite) TestEvictionRemovesFiles() {
	for i := 0; i < 7; i++ {
		err := s.cache.Put(fmt.Sprintf("key%d", i), []byte("x"))
		require.NoError(s.T(), err)
	}

	_, err := os.Stat(filepath.Join(s.dir, "key0"))
	assert.True(s.T(), os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(s.dir, "key1"))
	assert.True(s.T(), os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(s.dir, "key6"))
	assert.NoError(s.T(), err)
}

func (s *FileCacheSuite) TestLRUOrder() {
	for i := 0; i < 5; i++ {
		err := s.cache.Put(fmt.Sprintf("key%d", i), []byte("x"))
		require.NoError(s.T(), err)
	}

	_, ok := s.cache.Get("key0")
	assert.True(s.T(), ok)

	err := s.cache.Put("key5", []byte("x"))
	require.NoError(s.T(), err)

	_, ok = s.cache.Get("key1")
	assert.False(s.T(), ok)

	_, ok = s.cache.Get("key0")
	assert.True(s.T(), ok)
}

func (s *FileCacheSuite) TestExternalFileRemoval() {
	err := s.cache.Put("key1", []byte("data"))
	require.NoError(s.T(), err)

	os.Remove(filepath.Join(s.dir, "key1"))

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

	assert.LessOrEqual(s.T(), s.cache.Len(), 5)
}

func (s *FileCacheSuite) TestAtomicWrite() {
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
