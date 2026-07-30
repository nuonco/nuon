package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCustomCheckStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"healthy is valid", "healthy", false},
		{"degraded is valid", "degraded", false},
		{"unhealthy is valid", "unhealthy", false},
		{"unknown is valid", "unknown", false},
		{"progressing is not a valid custom check status", "progressing", true},
		{"not-applicable is not a valid custom check status", "not-applicable", true},
		{"empty status is invalid", "", true},
		{"arbitrary garbage is invalid", "OK", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCustomCheckStatus(tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBoundCustomCheckMessage(t *testing.T) {
	t.Run("short message passes through unchanged", func(t *testing.T) {
		assert.Equal(t, "all good", boundCustomCheckMessage("all good"))
	})

	t.Run("oversized message is truncated to the byte cap", func(t *testing.T) {
		msg := strings.Repeat("a", maxCustomCheckMessageBytes+500)
		got := boundCustomCheckMessage(msg)
		assert.LessOrEqual(t, len(got), maxCustomCheckMessageBytes)
		assert.Equal(t, strings.Repeat("a", maxCustomCheckMessageBytes), got)
	})

	t.Run("truncation never splits a multi-byte rune", func(t *testing.T) {
		msg := strings.Repeat("a", maxCustomCheckMessageBytes-1) + "日本語"
		got := boundCustomCheckMessage(msg)
		assert.True(t, len(got) <= maxCustomCheckMessageBytes)
		for _, r := range got {
			assert.NotEqual(t, rune(0xFFFD), r, "truncated message must not contain the invalid-UTF8 replacement rune")
		}
	})
}

func TestBoundCustomCheckDetails(t *testing.T) {
	t.Run("small details pass through unchanged", func(t *testing.T) {
		assert.Equal(t, `{"a":1}`, boundCustomCheckDetails(`{"a":1}`))
	})

	t.Run("oversized details are replaced with a truncation marker", func(t *testing.T) {
		details := strings.Repeat("a", maxCustomCheckDetailsBytes+1)
		assert.Equal(t, `{"_truncated":true}`, boundCustomCheckDetails(details))
	})
}

func TestValidateCustomCheckName(t *testing.T) {
	valid := []string{"checkout-latency", "a", "p99.latency", "queue_depth", "A1-b2.c3_d4"}
	for _, name := range valid {
		assert.NoError(t, validateCustomCheckName(name), name)
	}

	invalid := []string{
		"",
		"-leading-dash",
		"trailing-dash-",
		".leading-dot",
		"has space",
		"has/slash",
		"has\nnewline",
		"emoji💥",
		strings.Repeat("a", 101),
	}
	for _, name := range invalid {
		assert.Error(t, validateCustomCheckName(name), "%q should be rejected", name)
	}
}
