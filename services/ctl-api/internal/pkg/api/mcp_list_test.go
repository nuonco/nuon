package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPListPage(t *testing.T) {
	t.Parallel()

	limit, offset, err := MCPListPage(0, 0)
	require.NoError(t, err)
	assert.Equal(t, MCPDefaultListLimit, limit)
	assert.Equal(t, 0, offset)

	limit, offset, err = MCPListPage(50, 40)
	require.NoError(t, err)
	assert.Equal(t, 50, limit)
	assert.Equal(t, 40, offset)

	_, _, err = MCPListPage(101, 0)
	require.Error(t, err)

	_, _, err = MCPListPage(-1, 0)
	require.Error(t, err)

	_, _, err = MCPListPage(10, -1)
	require.Error(t, err)
}

func TestMCPClipList(t *testing.T) {
	t.Parallel()

	page, hasMore := MCPClipList([]int{1, 2, 3}, 3)
	assert.Equal(t, []int{1, 2, 3}, page)
	assert.False(t, hasMore)

	page, hasMore = MCPClipList([]int{1, 2, 3, 4}, 3)
	assert.Equal(t, []int{1, 2, 3}, page)
	assert.True(t, hasMore)
}

func TestMCPNextOffset(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, MCPNextOffset(20, 20, false))
	assert.Equal(t, 40, MCPNextOffset(20, 20, true))
}
