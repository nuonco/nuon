package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPhoneHomeTokenTimeoutIsLongLived(t *testing.T) {
	assert.Greater(t, phoneHomeTokenTimeout, 9*365*24*time.Hour,
		"a phone-home token must outlive any plausible deployed stack")
}
