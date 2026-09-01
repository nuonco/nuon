package activities

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommitStatusContext(t *testing.T) {
	require.Equal(t, "acme/payments/production", CommitStatusContext("acme", "payments", "production"))
	require.Equal(t, "nuon", CommitStatusContext("", "", ""))
	require.Len(t, CommitStatusContext(strings.Repeat("x", 300), "payments", "production"), maxCommitStatusContextLen)
}
