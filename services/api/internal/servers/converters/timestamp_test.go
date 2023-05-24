package converters

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalToUTC(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// Create a specific date and time in the given location
	localTime := time.Date(2023, time.May, 23, 18, 04, 05, 06, location)

	zuluTimeActual := TimeToDatetime(localTime)
	assert.Equal(t, int32(22), zuluTimeActual.Hours) // most likely field for time zone bugs
	assert.Equal(t, int32(4), zuluTimeActual.Minutes)
	assert.Equal(t, int32(5), zuluTimeActual.Seconds)
	assert.Equal(t, int32(6), zuluTimeActual.Nanos)
	assert.Equal(t, int32(2023), zuluTimeActual.Year)
	assert.Equal(t, int32(time.May), zuluTimeActual.Month)
	assert.Equal(t, int32(23), zuluTimeActual.Day)
}
