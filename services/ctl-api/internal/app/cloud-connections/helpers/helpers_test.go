package helpers

import (
	"errors"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestChooseAuthMode(t *testing.T) {
	accessDenied := &smithy.GenericAPIError{Code: "AccessDenied", Message: "denied"}
	otherErr := errors.New("network failed")
	tests := map[string]struct {
		current         app.CloudConnectionAuthMode
		oidcErr         error
		legacyErr       error
		want            app.CloudConnectionAuthMode
		wantErr         string
		wantOIDCCalls   int
		wantLegacyCalls int
	}{
		"legacy to OIDC success never falls back": {current: app.CloudConnectionAuthModeLegacy, want: app.CloudConnectionAuthModeOIDC, wantOIDCCalls: 1},
		"OIDC mode never falls back":              {current: app.CloudConnectionAuthModeOIDC, oidcErr: accessDenied, wantErr: "denied", wantOIDCCalls: 1},
		"access denied falls back once":           {current: app.CloudConnectionAuthModeLegacy, oidcErr: accessDenied, want: app.CloudConnectionAuthModeLegacy, wantOIDCCalls: 1, wantLegacyCalls: 1},
		"other errors propagate":                  {current: app.CloudConnectionAuthModeLegacy, oidcErr: otherErr, wantErr: "network failed", wantOIDCCalls: 1},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			oidcCalls, legacyCalls := 0, 0
			mode, err := chooseAuthMode(test.current, func() error { oidcCalls++; return test.oidcErr }, func() error { legacyCalls++; return test.legacyErr })
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want, mode)
			}
			assert.Equal(t, test.wantOIDCCalls, oidcCalls)
			assert.Equal(t, test.wantLegacyCalls, legacyCalls)
		})
	}
}
