package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workloadjwt"
)

// ensureAzurePhoneHomeIdentity records which managed identity each of this install's live
// stack versions should phone home as.
//
// Azure needs no credential published anywhere: the stack presents a token Entra already
// signed. So unlike the AWS path there is nothing to mint, encrypt, or grant — only a name
// that the rendered template and the verifier both derive the same way.
//
// Convergent like its AWS counterpart. The name is written to every eligible version and
// cleared from the rest, so a re-run after a partial failure is a no-op and a retired
// version cannot keep authenticating.
func (a *Activities) ensureAzurePhoneHomeIdentity(
	ctx context.Context, install *app.Install, ignoreFeatureGate bool,
) (*EnsureInstallPhoneHomeSecretResponse, error) {
	if skip, err := a.azurePhoneHomeSkipReason(ctx, install, ignoreFeatureGate); err != nil {
		return nil, err
	} else if skip != "" {
		return &EnsureInstallPhoneHomeSecretResponse{Skipped: true, SkipReason: skip}, nil
	}

	name := workloadjwt.AzurePhoneHomeIdentityName(install.ID)
	resp := &EnsureInstallPhoneHomeSecretResponse{IdentityName: name}

	if install.InstallStack == nil {
		return resp, nil
	}

	for i := range install.InstallStack.InstallStackVersions {
		version := &install.InstallStack.InstallStackVersions[i]

		want := name
		if !version.PhoneHomeTokenEligible() {
			want = ""
		}
		if version.PhoneHomeIdentityName == want {
			continue
		}

		if res := a.db.WithContext(ctx).
			Model(&app.InstallStackVersion{ID: version.ID}).
			Update("phone_home_identity_name", want); res.Error != nil {
			return nil, generics.TemporalGormError(res.Error)
		}
		version.PhoneHomeIdentityName = want

		if want == "" {
			resp.TokensRevoked++
		} else {
			resp.TokensMinted++
		}
	}

	a.l.Info("reconciled azure install phone home identity",
		zap.String("install_id", install.ID),
		zap.String("identity_name", name),
		zap.Int("versions_set", resp.TokensMinted),
		zap.Int("versions_cleared", resp.TokensRevoked),
	)

	return resp, nil
}

// azurePhoneHomeSkipReason returns a non-empty reason when this install must be passed
// over. Every one is a clean no-op: enforcement is opt-in per cloud and a miss must never
// fail a provision.
func (a *Activities) azurePhoneHomeSkipReason(
	ctx context.Context, install *app.Install, ignoreFeatureGate bool,
) (string, error) {
	if !ignoreFeatureGate {
		enabled, err := a.features.OrgHasFeature(ctx, install.OrgID, app.OrgFeaturePhoneHomeAuth)
		if err != nil {
			return "", fmt.Errorf("unable to check phone home auth feature: %w", err)
		}
		if !enabled {
			return phoneHomeSkipFeatureDisabled, nil
		}
	}

	// The org is checked too because AdminForceSandboxMode flips the org without
	// updating installs.sandbox_mode.
	if install.SandboxMode.Bool || install.Org.SandboxMode {
		return phoneHomeSkipSandboxMode, nil
	}

	// The verifier binds the token's subscription to this value, so rendering an
	// identity without it would mean enforcing against nothing.
	if install.CloudPlatformMetadata.TargetSubscriptionID == "" {
		a.l.Info("skipping azure phone home identity: no target subscription",
			zap.String("install_id", install.ID), zap.String("org_id", install.OrgID))

		return phoneHomeSkipNoSubscription, nil
	}

	return "", nil
}
