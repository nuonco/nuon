package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestSecretIsRevealable(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := map[string]struct {
		secret app.TriggerSecret
		want   bool
	}{
		"active":               {secret: app.TriggerSecret{Secret: "value"}, want: true},
		"active with expiry":   {secret: app.TriggerSecret{Secret: "value", ExpiresAt: &future}, want: true},
		"revoked":              {secret: app.TriggerSecret{Secret: "value", RevokedAt: &past}, want: false},
		"expired":              {secret: app.TriggerSecret{Secret: "value", ExpiresAt: &past}, want: false},
		"expiring exactly now": {secret: app.TriggerSecret{Secret: "value", ExpiresAt: &now}, want: false},
		"scrubbed":             {secret: app.TriggerSecret{Secret: ""}, want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, secretIsRevealable(&tt.secret, now))
		})
	}
}
