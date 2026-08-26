package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseIncludeDiff(t *testing.T) {
	require.Equal(t, map[string]bool{"git": true, "full": true}, parseIncludeDiff("git,full"))
	require.Equal(t, map[string]bool{"config": true}, parseIncludeDiff(" CONFIG "))
	require.Empty(t, parseIncludeDiff(""))
	require.Empty(t, parseIncludeDiff(",,"))
}
