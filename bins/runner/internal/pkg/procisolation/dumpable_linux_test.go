//go:build linux

package procisolation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestApplyMarksProcessNonDumpable(t *testing.T) {
	require.NoError(t, Apply())

	dumpable, err := unix.PrctlRetInt(unix.PR_GET_DUMPABLE, 0, 0, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, dumpable)
}
