package plans

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecompressPlan_Raw(t *testing.T) {
	// Test with a simple JSON object
	original := []byte(`{"key": "value", "number": 42}`)

	// Compress
	encoded, err := CompressPlan(original)
	require.NoError(t, err)

	// Decompress
	decompressed, err := DecompressPlan(encoded)
	require.NoError(t, err)

	// Compare the results
	assert.Equal(t, original, decompressed)
}

func TestCompressPlan_Large(t *testing.T) {
	// Create a large byte array
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// Compress
	encoded, err := CompressPlan(largeData)
	require.NoError(t, err)

	// Decompress
	decompressed, err := DecompressPlan(encoded)
	require.NoError(t, err)

	// Compare the results
	assert.Equal(t, largeData, decompressed)
}
