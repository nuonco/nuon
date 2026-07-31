package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBearerToken(t *testing.T) {
	for _, tc := range []struct {
		name   string
		header string
		want   string
	}{
		{"canonical", "Bearer tok_abc", "tok_abc"},
		{"lowercase scheme", "bearer tok_abc", "tok_abc"},
		{"mixed case scheme", "BeArEr tok_abc", "tok_abc"},
		{"surrounding whitespace", "  Bearer   tok_abc  ", "tok_abc"},
		{"empty header", "", ""},
		{"scheme only", "Bearer", ""},
		{"scheme and space only", "Bearer ", ""},
		{"wrong scheme", "Basic dXNlcjpwYXNz", ""},
		{"bare token with no scheme", "tok_abc", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, bearerToken(tc.header))
		})
	}
}

// The payload's account is compared against the coalesced expected identifier, so a
// backfilled install (observed only) is protected the same as one pinned at creation.
func TestCheckObservedCloudAccount(t *testing.T) {
	const expected = "123456789012"

	for _, tc := range []struct {
		name       string
		install    *app.Install
		props      map[string]any
		wantReason string
	}{
		{
			name:       "matching account passes",
			install:    &app.Install{ExpectedAccountID: expected},
			props:      map[string]any{"account_id": expected},
			wantReason: phoneHomeAuthOK,
		},
		{
			name:       "differing account is rejected",
			install:    &app.Install{ExpectedAccountID: expected},
			props:      map[string]any{"account_id": "999999999999"},
			wantReason: phoneHomeRejectAccountMismatch,
		},
		{
			// Nothing to compare against: an install predating the target field, and
			// not yet reached by the metadata backfill.
			name:       "no expected account is not a rejection",
			install:    &app.Install{},
			props:      map[string]any{"account_id": "999999999999"},
			wantReason: phoneHomeAuthOK,
		},
		{
			// GCP and Azure stacks post no account_id; rejecting them here would break
			// clouds this check does not cover.
			name:       "a payload without an account is not a rejection",
			install:    &app.Install{ExpectedAccountID: expected},
			props:      map[string]any{},
			wantReason: phoneHomeAuthOK,
		},
		{
			name:       "a non-string account is treated as absent",
			install:    &app.Install{ExpectedAccountID: expected},
			props:      map[string]any{"account_id": 123456789012},
			wantReason: phoneHomeAuthOK,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reason, err := checkObservedCloudAccount(tc.install, tc.props)

			assert.Equal(t, tc.wantReason, reason)
			if tc.wantReason == phoneHomeAuthOK {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

// Every rejection reaches the caller as the same opaque 401. Distinguishing "no such
// install" from "wrong token" would confirm an install's existence to a caller holding
// nothing but a leaked phone_home_id — the leak this feature exists to close.
func TestRejectPhoneHomeIsOpaqueButKeepsTheReason(t *testing.T) {
	detailed := "token belongs to acct_123, not stack version isv_456"

	err := rejectPhoneHome(phoneHomeRejectWrongAccount, assert.AnError)
	assert.NotContains(t, err.Error(), detailed)

	var authErr errPhoneHomeAuth
	assert.ErrorAs(t, err, &authErr)
	assert.Equal(t, phoneHomeRejectWrongAccount, authErr.reason,
		"the reason must survive for the metric tag even though the client never sees it")
}

// Guards the tag values against silent drift: these strings are what dashboards and
// alerts key on, and revoked_stack_version in particular is the only signal that the
// revocation policy cut off a stack still in use.
func TestPhoneHomeRejectionReasonsAreStable(t *testing.T) {
	assert.Equal(t, "missing_token", phoneHomeRejectMissingToken)
	assert.Equal(t, "unknown_token", phoneHomeRejectUnknownToken)
	assert.Equal(t, "expired_token", phoneHomeRejectExpiredToken)
	assert.Equal(t, "wrong_account", phoneHomeRejectWrongAccount)
	assert.Equal(t, "revoked_stack_version", phoneHomeRejectRevokedVersion)
	assert.Equal(t, "version_expired", phoneHomeRejectVersionExpired)
	assert.Equal(t, "account_mismatch", phoneHomeRejectAccountMismatch)
	assert.Equal(t, "skipped", phoneHomeAuthSkipped)
	assert.Equal(t, "", phoneHomeAuthOK, "an empty reason must mean verified")
}
