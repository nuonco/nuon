package service

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
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

func TestInstallPhoneHomeRejectsNonStringRequestType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"request_type":123}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	require.NotPanics(t, func() {
		(&service{}).InstallPhoneHome(ctx)
	})
	require.Len(t, ctx.Errors, 1)
	assert.Contains(t, ctx.Errors[0].Error(), "request type param must be a string")
	var invalidRequest stderr.ErrInvalidRequest
	assert.ErrorAs(t, ctx.Errors[0].Err, &invalidRequest)
}

func (s *InstallsServiceTestSuite) TestAuthorizePhoneHomeRejectsPreviousToken() {
	install := s.createTestInstall()
	stackVersion := s.deps.Seeder.CreateInstallStackVersion(
		s.ctx, s.T(), install.ID, install.InstallStack.ID, install.AppConfigID,
	)

	serviceAccountEmail := account.ServiceAccountEmail(stackVersion.ID)
	serviceAccount := &app.Account{
		Email:       serviceAccountEmail,
		Subject:     serviceAccountEmail,
		AccountType: app.AccountTypeService,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(serviceAccount).Error)

	now := time.Now()
	previous := &app.Token{
		CreatedByID: s.testAcc.ID,
		AccountID:   serviceAccount.ID,
		Token:       "previous-phone-home-token",
		ExpiresAt:   now.Add(time.Hour),
		IssuedAt:    now,
		Issuer:      "test",
	}
	current := &app.Token{
		CreatedByID: s.testAcc.ID,
		AccountID:   serviceAccount.ID,
		Token:       "current-phone-home-token",
		ExpiresAt:   now.Add(time.Hour),
		IssuedAt:    now,
		Issuer:      "test",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(previous).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(current).Error)

	stackVersion.PhoneHomeTokenID = current.ID
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(stackVersion).
		Update("phone_home_token_id", current.ID).Error)

	s.testOrg.Features[string(app.OrgFeaturePhoneHomeAuth)] = true
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(s.testOrg).
		Update("features", s.testOrg.Features).Error)

	reason, err := s.installsService.authorizePhoneHome(
		s.ctx, install, stackVersion, "Bearer "+previous.Token,
	)
	assert.Equal(s.T(), phoneHomeRejectUnknownToken, reason)
	assert.Error(s.T(), err)

	reason, err = s.installsService.authorizePhoneHome(
		s.ctx, install, stackVersion, "Bearer "+current.Token,
	)
	assert.Equal(s.T(), phoneHomeAuthOK, reason)
	assert.NoError(s.T(), err)
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
			// A stack that posts no identifier at all has nothing to contradict.
			name:       "a payload without an account is not a rejection",
			install:    &app.Install{ExpectedAccountID: expected},
			props:      map[string]any{},
			wantReason: phoneHomeAuthOK,
		},
		{
			name:       "matching azure subscription passes",
			install:    &app.Install{ExpectedSubscriptionID: "AAAAAAAA-bbbb-cccc-dddd-eeeeeeeeeeee"},
			props:      map[string]any{"subscription_id": "aaaaaaaa-BBBB-cccc-dddd-eeeeeeeeeeee"},
			wantReason: phoneHomeAuthOK,
		},
		{
			name:       "differing azure subscription is rejected",
			install:    &app.Install{ExpectedSubscriptionID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"},
			props:      map[string]any{"subscription_id": "ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee"},
			wantReason: phoneHomeRejectAccountMismatch,
		},
		{
			name:       "differing gcp project is rejected",
			install:    &app.Install{ExpectedProjectID: "acme-prod"},
			props:      map[string]any{"project_id": "acme-staging"},
			wantReason: phoneHomeRejectAccountMismatch,
		},
		{
			// An install carries one cloud's identifier; the others must not false-positive.
			name:       "an unrelated cloud identifier is ignored",
			install:    &app.Install{ExpectedAccountID: expected},
			props:      map[string]any{"subscription_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"},
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
	assert.Equal(t, "invalid_identity_token", phoneHomeRejectIdentityToken)
	assert.Equal(t, "identity_mismatch", phoneHomeRejectIdentityMismatch)
	assert.Equal(t, "tenant_mismatch", phoneHomeRejectTenantMismatch)
	assert.Equal(t, "principal_mismatch", phoneHomeRejectPrincipalPinned)
	assert.Equal(t, "no_target_account", phoneHomeRejectNoTargetAccount)
	assert.Equal(t, "skipped", phoneHomeAuthSkipped)
	assert.Equal(t, "", phoneHomeAuthOK, "an empty reason must mean verified")
}
