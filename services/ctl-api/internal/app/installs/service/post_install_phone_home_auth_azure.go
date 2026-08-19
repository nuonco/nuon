package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workloadjwt"
)

const (
	phoneHomeRejectIdentityToken    = "invalid_identity_token"
	phoneHomeRejectIdentityMismatch = "identity_mismatch"
	phoneHomeRejectTenantMismatch   = "tenant_mismatch"
	phoneHomeRejectPrincipalPinned  = "principal_mismatch"
	phoneHomeRejectNoTargetAccount  = "no_target_account"
)

// authorizeAzurePhoneHome verifies that the caller is the managed identity this stack
// version's template was rendered with.
//
// The verified signature only proves which Entra tenant minted the token, so it carries
// no authorization on its own. The binding is subscription plus identity name: Azure
// assigns the subscription in a resource id and signs it into xms_mirid, so a token
// naming this install's subscription can only have come from an identity inside it.
func (s *service) authorizeAzurePhoneHome(
	ctx context.Context, install *app.Install, stackVersion *app.InstallStackVersion, authHeader string,
) (string, error) {
	raw := bearerToken(authHeader)
	if raw == "" {
		return phoneHomeRejectMissingToken, rejectPhoneHome(
			phoneHomeRejectMissingToken, errors.New("no bearer token presented"),
		)
	}

	// Without a pinned subscription there is nothing to bind to, and a token from any
	// Entra tenant would pass. Fail closed: the template only carries an identity once
	// the feature is on, and the feature requires a subscription at creation.
	if install.CloudPlatformMetadata.TargetSubscriptionID == "" {
		return phoneHomeRejectNoTargetAccount, rejectPhoneHome(
			phoneHomeRejectNoTargetAccount,
			fmt.Errorf("install %s has no target subscription to bind against", install.ID),
		)
	}

	// A pinned tenant is authoritative. Before the first success there is none, so the
	// token's own tenant selects the key set and the subscription check below is what
	// rejects a token minted anywhere else.
	tenantID := install.CloudPlatformMetadata.TargetTenantID
	if tenantID == "" {
		tenantID = install.CloudPlatformMetadata.ObservedTenantID
	}
	if tenantID == "" {
		var err error
		if tenantID, err = unverifiedAzureTenantID(raw); err != nil {
			return s.rejectAzureIdentityToken(install, err)
		}
	}

	issuer, err := workloadjwt.AzureIssuer(tenantID)
	if err != nil {
		return s.rejectAzureIdentityToken(install, err)
	}

	claims, err := s.workloadJWT.Verify(ctx, workloadjwt.Request{
		Token:    raw,
		Issuer:   issuer,
		Audience: workloadjwt.AzureGraphAudience,
	})
	if err != nil {
		return s.rejectAzureIdentityToken(install, err)
	}

	identity, err := workloadjwt.ParseAzureManagedIdentity(claims)
	if err != nil {
		return s.rejectAzureIdentityToken(install, err)
	}

	if reason, err := bindAzureIdentity(install, stackVersion, identity, tenantID); err != nil {
		return reason, err
	}

	s.pinAzurePhoneHomeIdentity(ctx, install, identity)

	return phoneHomeAuthOK, nil
}

// bindAzureIdentity is the authorization step: it decides whether a token Entra has
// already vouched for belongs to this install. Kept free of IO so every rejection path is
// directly testable.
func bindAzureIdentity(
	install *app.Install,
	stackVersion *app.InstallStackVersion,
	identity *workloadjwt.AzureManagedIdentity,
	expectedTenantID string,
) (string, error) {
	// Guards against a pinned tenant being bypassed by a token whose own tid selected the
	// key set it was verified against.
	if !strings.EqualFold(identity.TenantID, expectedTenantID) {
		return phoneHomeRejectTenantMismatch, rejectPhoneHome(
			phoneHomeRejectTenantMismatch,
			fmt.Errorf("token tenant %s is not the tenant expected for this install", identity.TenantID),
		)
	}

	// The load-bearing check. Azure assigns the subscription in a resource id and signs it
	// into xms_mirid, so a token naming this subscription came from inside it.
	if !strings.EqualFold(identity.SubscriptionID, install.CloudPlatformMetadata.TargetSubscriptionID) {
		return phoneHomeRejectAccountMismatch, rejectPhoneHome(
			phoneHomeRejectAccountMismatch,
			fmt.Errorf("token subscription %s does not match the subscription expected for this install",
				identity.SubscriptionID),
		)
	}

	// Scopes the credential to one install: every install in a subscription renders a
	// differently named identity, so a neighbour's token cannot post here.
	if stackVersion.PhoneHomeIdentityName == "" ||
		!strings.EqualFold(identity.Name, stackVersion.PhoneHomeIdentityName) {
		return phoneHomeRejectIdentityMismatch, rejectPhoneHome(
			phoneHomeRejectIdentityMismatch,
			fmt.Errorf("token identity %q is not %q", identity.Name, stackVersion.PhoneHomeIdentityName),
		)
	}

	if pinned := install.CloudPlatformMetadata.ObservedPhoneHomePrincipalID; pinned != "" &&
		!strings.EqualFold(pinned, identity.PrincipalID) {
		return phoneHomeRejectPrincipalPinned, rejectPhoneHome(
			phoneHomeRejectPrincipalPinned,
			fmt.Errorf("token principal %s is not the principal pinned for this install", identity.PrincipalID),
		)
	}

	return phoneHomeAuthOK, nil
}

// rejectAzureIdentityToken keeps verification detail out of the response. The underlying
// error quotes the presented token's own claims and, for a discovery failure, Entra's reply
// verbatim -- echoing either back to an unauthenticated caller tells them how far they got.
func (s *service) rejectAzureIdentityToken(install *app.Install, err error) (string, error) {
	s.l.Warn("rejected azure phone home identity token",
		zap.String("install_id", install.ID), zap.Error(err))

	return phoneHomeRejectIdentityToken, rejectPhoneHome(
		phoneHomeRejectIdentityToken, errors.New("identity token verification failed"),
	)
}

// pinAzurePhoneHomeIdentity records the tenant and principal the first verified call came
// from. Best-effort: a failure here must not reject a call that already verified, and the
// next success retries it.
func (s *service) pinAzurePhoneHomeIdentity(
	ctx context.Context, install *app.Install, identity *workloadjwt.AzureManagedIdentity,
) {
	cpm := install.CloudPlatformMetadata
	if cpm.ObservedTenantID == identity.TenantID &&
		cpm.ObservedSubscriptionID == identity.SubscriptionID &&
		cpm.ObservedPhoneHomePrincipalID == identity.PrincipalID {
		return
	}

	cpm.ObservedTenantID = identity.TenantID
	cpm.ObservedSubscriptionID = identity.SubscriptionID
	cpm.ObservedPhoneHomePrincipalID = identity.PrincipalID

	if res := s.db.WithContext(ctx).
		Model(&app.Install{ID: install.ID}).
		Update("cloud_platform_metadata", cpm); res.Error != nil {
		s.l.Warn("unable to pin azure phone home identity",
			zap.String("install_id", install.ID), zap.Error(res.Error))

		return
	}
	install.CloudPlatformMetadata = cpm
}

// unverifiedAzureTenantID reads tid from an unverified token, only to pick which tenant's
// keys to check the signature with. The value is re-checked against the verified claims
// afterwards, and authorization never rests on it.
func unverifiedAzureTenantID(token string) (string, error) {
	claims, err := workloadjwt.UnverifiedClaims(token)
	if err != nil {
		return "", err
	}

	tenantID, ok := workloadjwt.StringClaim(claims, "tid")
	if !ok {
		return "", errors.New("token is missing the tid claim")
	}

	return tenantID, nil
}
