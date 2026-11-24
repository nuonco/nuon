package get

import (
	"testing"

	getter "github.com/hashicorp/go-getter"
	"github.com/stretchr/testify/require"
)

func TestGetDetectorsMakeNoRequests(t *testing.T) {
	for _, d := range GetDetectors() {
		_, bitbucket := d.(*getter.BitBucketDetector)
		require.False(t, bitbucket, "the Bitbucket detector calls api.bitbucket.org while it detects")
	}

	detected, err := getter.Detect("bitbucket.org/acme/widgets", t.TempDir(), GetDetectors())
	require.NoError(t, err)
	require.Contains(t, detected, "file://")
}
