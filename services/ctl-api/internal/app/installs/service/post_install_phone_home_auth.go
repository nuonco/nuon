package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	phoneHomeAuthOK = ""

	phoneHomeRejectMissingToken    = "missing_token"
	phoneHomeRejectUnknownToken    = "unknown_token"
	phoneHomeRejectExpiredToken    = "expired_token"
	phoneHomeRejectWrongAccount    = "wrong_account"
	phoneHomeRejectRevokedVersion  = "revoked_stack_version"
	phoneHomeRejectVersionExpired  = "version_expired"
	phoneHomeRejectAccountMismatch = "account_mismatch"

	phoneHomeAuthSkipped = "skipped"
)

type errPhoneHomeAuth struct {
	reason string
	err    error
}

func (e errPhoneHomeAuth) Error() string { return e.err.Error() }
func (e errPhoneHomeAuth) Unwrap() error { return e.err }

// why: rejectPhoneHome is the only 401 the caller ever sees. The description is
// deliberately uniform: distinguishing "no such install" from "wrong token" would
// confirm the existence of installs to an unauthenticated caller holding nothing but a
// leaked phone_home_id, which is the exact leak this feature closes.
func rejectPhoneHome(reason string, err error) error {
	return errPhoneHomeAuth{
		reason: reason,
		err: stderr.ErrAuthentication{
			Err:         err,
			Description: "phone home authentication failed",
		},
	}
}

// why: authorizePhoneHome verifies that this request was made by the stack version it claims
// to be, returning the reason to record.
//
// Enforcement requires three things to line up: the org is enrolled, this version has a
// token minted for it, and that token has not been tombstoned. Anything less is a skip
// rather than a rejection, so enabling the flag cannot break a stack version that
// predates minting.
//
// A tombstoned version is the exception — it is rejected, not skipped. Treating a
// deliberately revoked token as "never minted" would silently reopen the hole
// revocation just closed.
func (s *service) authorizePhoneHome(
	ctx context.Context, install *app.Install, stackVersion *app.InstallStackVersion, authHeader string,
) (string, error) {
	enabled, err := s.featuresClient.OrgHasFeature(ctx, install.OrgID, app.OrgFeaturePhoneHomeAuth)
	if err != nil {
		return "", fmt.Errorf("unable to check phone home auth feature: %w", err)
	}

	revoked := stackVersion.PhoneHomeTokenRevokedAt != nil
	switch {
	case !enabled:
		return phoneHomeAuthSkipped, nil
	case revoked:
		return phoneHomeRejectRevokedVersion, rejectPhoneHome(
			phoneHomeRejectRevokedVersion,
			fmt.Errorf("phone home token for stack version %s was revoked", stackVersion.ID),
		)
	case stackVersion.PhoneHomeIdentityName == "" && stackVersion.PhoneHomeTokenID == "":
		return phoneHomeAuthSkipped, nil
	}

	if stackVersion.Status.Status == app.InstallStackVersionStatusExpired {
		return phoneHomeRejectVersionExpired, errPhoneHomeAuth{
			reason: phoneHomeRejectVersionExpired,
			err: stderr.ErrConflict{
				Err:         fmt.Errorf("install stack version %s has expired", stackVersion.ID),
				Description: "this install stack version has expired and needs to be reprovisioned",
			},
		}
	}

	if stackVersion.PhoneHomeIdentityName != "" {
		return s.authorizeAzurePhoneHome(ctx, install, stackVersion, authHeader)
	}

	raw := bearerToken(authHeader)
	if raw == "" {
		return phoneHomeRejectMissingToken, rejectPhoneHome(
			phoneHomeRejectMissingToken, errors.New("no bearer token presented"),
		)
	}

	var token app.Token
	res := s.db.WithContext(ctx).Where(&app.Token{
		ID:    stackVersion.PhoneHomeTokenID,
		Token: raw,
	}).First(&token)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return phoneHomeRejectUnknownToken, rejectPhoneHome(
			phoneHomeRejectUnknownToken, errors.New("token not recognized"),
		)
	}
	if res.Error != nil {
		return "", fmt.Errorf("unable to look up phone home token: %w", res.Error)
	}
	if time.Now().After(token.ExpiresAt) {
		return phoneHomeRejectExpiredToken, rejectPhoneHome(
			phoneHomeRejectExpiredToken, errors.New("token is expired"),
		)
	}

	var acct app.Account
	if res := s.db.WithContext(ctx).
		Where(app.Account{ID: token.AccountID}).
		First(&acct); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return phoneHomeRejectWrongAccount, rejectPhoneHome(
				phoneHomeRejectWrongAccount, errors.New("token has no account"),
			)
		}

		return "", fmt.Errorf("unable to resolve token account: %w", res.Error)
	}

	want := account.ServiceAccountEmail(stackVersion.ID)
	if acct.Email != want && acct.Subject != want {
		return phoneHomeRejectWrongAccount, rejectPhoneHome(
			phoneHomeRejectWrongAccount,
			fmt.Errorf("token belongs to %s, not stack version %s", acct.ID, stackVersion.ID),
		)
	}

	return phoneHomeAuthOK, nil
}

func bearerToken(header string) string {
	const prefix = "bearer "

	header = strings.TrimSpace(header)
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}

	return strings.TrimSpace(header[len(prefix):])
}

func checkObservedCloudAccount(install *app.Install, props map[string]any) (string, error) {
	for _, field := range []struct {
		payloadKey string
		expected   string
	}{
		{"account_id", install.ExpectedAccountID},
		{"subscription_id", install.ExpectedSubscriptionID},
		{"project_id", install.ExpectedProjectID},
	} {
		observed, _ := props[field.payloadKey].(string)
		if observed == "" || field.expected == "" || strings.EqualFold(observed, field.expected) {
			continue
		}

		return phoneHomeRejectAccountMismatch, rejectPhoneHome(
			phoneHomeRejectAccountMismatch,
			fmt.Errorf("payload %s %s does not match the account expected for this install",
				field.payloadKey, observed),
		)
	}

	return phoneHomeAuthOK, nil
}

func (s *service) recordPhoneHomeAuthResult(ctx context.Context, install *app.Install, verified bool) {
	auth := app.PhoneHomeAuth{}
	if install.PhoneHomeAuth != nil {
		auth = *install.PhoneHomeAuth
	}

	now := time.Now()
	if verified {
		auth.LastVerifiedAt = &now
	} else {
		auth.LastRejectedAt = &now
	}

	if res := s.db.WithContext(ctx).
		Model(&app.Install{ID: install.ID}).
		Update("phone_home_auth", auth); res.Error != nil {
		s.l.Warn("unable to record phone home auth result",
			zap.String("install_id", install.ID), zap.Error(res.Error))
	}
}
