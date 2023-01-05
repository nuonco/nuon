package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/powertoolsdev/api/internal/domain"
	"github.com/stretchr/testify/assert"
)

var (
	validClaim = CustomClaim{
		ExternalID:   "google-id",
		ShouldReject: false,
	}
	shouldRejectClaim = CustomClaim{
		ShouldReject: true,
	}
)

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		claim       CustomClaim
		errExpected error
	}{
		"valid claim": {
			claim: validClaim,
		},
		"should reject": {
			claim:       shouldRejectClaim,
			errExpected: errors.New("should reject was set to true"),
		},
		"external ID is required": {
			claim: CustomClaim{
				ExternalID: "",
			},
			errExpected: errors.New("subject must be set"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.claim.Validate(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestJwt(t *testing.T) {
	tests := map[string]struct {
		cfg         domain.Config
		errExpected error
	}{
		"valid": {
			cfg: domain.Config{
				AuthIssuerURL: "https://nuon.co",
				AuthAudience:  "https://nuon.co/audience",
			},
		},
		//"malformed url": {
		//cfg: api.Config{
		//AuthIssuerURL: "  https://nuon.co",
		//AuthAudience:  "https://nuon.co/audience",
		//},
		//errExpected: errors.New("parse "),
		//},
		"missing audience": {
			cfg: domain.Config{
				AuthIssuerURL: "https://nuon.co",
			},
			// NOTE(jdt): this doesn't actually error
			// errExpected: errors.New("something?"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mw, err := Jwt(test.cfg.AuthIssuerURL, test.cfg.AuthAudience)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			assert.NotNil(t, mw)
		})
	}
}
