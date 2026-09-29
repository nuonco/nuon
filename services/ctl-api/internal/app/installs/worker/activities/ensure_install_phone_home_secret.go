package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/secretsmanager"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// why: phoneHomeTokenTimeout is deliberately far longer than the 90-day runner token
// expiry. A deployed stack can be updated years after it was applied — a console
// parameter change, drift remediation, a customer CFN pipeline — and each of those
// re-invokes the phone-home Lambda. A shorter expiry would silently strand it. The
// expiry remains a real backstop, and the middleware checks it unconditionally.
const phoneHomeTokenTimeout = time.Hour * 24 * 365 * 10

const (
	phoneHomeSkipFeatureDisabled = "org feature phone-home-auth is disabled"
	phoneHomeSkipNotAWS          = "install is not an AWS install"
	phoneHomeSkipNoTargetAccount = "install has no target account id"
	phoneHomeSkipNoManagement    = "control plane cannot reach management secrets manager"
	phoneHomeSkipSandboxMode     = "install is in sandbox mode"

	phoneHomeSkipNoSubscription = "install has no target subscription id"
)

type EnsureInstallPhoneHomeSecretRequest struct {
	InstallID string `json:"install_id" validate:"required"`

	IgnoreOrgFeatureGate bool `json:"ignore_org_feature_gate,omitempty"`
}

type EnsureInstallPhoneHomeSecretResponse struct {
	Skipped    bool   `json:"skipped"`
	SkipReason string `json:"skip_reason,omitempty"`

	SecretARN    string `json:"secret_arn,omitempty"`
	SecretRegion string `json:"secret_region,omitempty"`

	TokensMinted  int  `json:"tokens_minted"`
	TokensRevoked int  `json:"tokens_revoked"`
	SecretWritten bool `json:"secret_written"`

	IdentityName string `json:"identity_name,omitempty"`
}

// EnsureInstallPhoneHomeSecret reconciles an install's phone-home credentials.
//
// It is a convergent reconciler, not a create hook: desired state is derived from
// Postgres and driven into Secrets Manager on every call, so the same code serves a
// fresh provision, a fleet backfill, and a token refresh. Callers never say what
// changed. Every operation is expressed as "mutate Postgres, then reconcile":
//
//   - a new stack version arrives with an empty PhoneHomeTokenID   -> minted here
//   - refresh: clear PhoneHomeTokenID, leave the tombstone nil     -> re-minted here
//     under the same phone_home_id, so no template re-render and no customer re-apply
//   - revoke: set PhoneHomeTokenRevokedAt                          -> dropped here
//
// The database is the source of truth and the secret is a projection of it, so a
// partial failure is always recoverable by running this again, and the map can never
// hold an entry no token backs.
//
// @temporal-gen-v2 activity
// @by-field InstallID
// @start-to-close-timeout 2m
func (a *Activities) EnsureInstallPhoneHomeSecret(
	ctx context.Context, req *EnsureInstallPhoneHomeSecretRequest,
) (*EnsureInstallPhoneHomeSecretResponse, error) {
	if err := a.v.StructCtx(ctx, req); err != nil {
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	var install app.Install
	if res := a.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("Org").
		Preload("InstallStack.InstallStackVersions").
		Where(app.Install{ID: req.InstallID}).
		First(&install); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	if install.AzureAccount != nil {
		return a.ensureAzurePhoneHomeIdentity(ctx, &install, req.IgnoreOrgFeatureGate)
	}

	if skip, err := a.phoneHomeSecretSkipReason(ctx, &install, req.IgnoreOrgFeatureGate); err != nil {
		return nil, err
	} else if skip != "" {
		return &EnsureInstallPhoneHomeSecretResponse{Skipped: true, SkipReason: skip}, nil
	}

	resp := &EnsureInstallPhoneHomeSecretResponse{}

	tokens, err := a.reconcileStackVersionPhoneHomeTokens(ctx, &install, resp)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(tokens)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal phone home token map: %w", err)
	}

	secret, err := a.secretsSvc.EnsureSecret(ctx, secretsmanager.EnsureSecretInput{
		Name:        secretsmanager.PhoneHomeSecretName(install.ID),
		Value:       string(payload),
		Description: fmt.Sprintf("Nuon phone home tokens for install %s", install.ID),
		KMSKeyARN:   a.cfg.AWSPhoneHomeCMKARN,
		Tags: secretsmanager.PhoneHomeSecretTags(
			install.OrgID, install.ID, a.cfg.RunnerAPIURL, string(a.cfg.Env),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to ensure phone home secret: %w", err)
	}

	resp.SecretARN = secret.ARN
	resp.SecretRegion = secret.Region
	resp.SecretWritten = secret.Wrote

	policy, err := secretsmanager.PhoneHomeResourcePolicy(
		install.CloudPlatformMetadata.TargetAccountID,
		stacks.PhoneHomeRoleName(install.ID),
	)
	if err != nil {
		return nil, err
	}

	if err := a.secretsSvc.PutResourcePolicy(ctx, secret.ARN, policy); err != nil {
		// why: An invalid target account never becomes valid, and the default policy would
		// retry forever, stranding the provisioning step.
		if secretsmanager.IsPermanentInputError(err) {
			return nil, temporal.NewNonRetryableApplicationError(
				fmt.Sprintf("unable to grant cross account read to target account %s",
					install.CloudPlatformMetadata.TargetAccountID),
				"phone_home_resource_policy_rejected",
				err,
			)
		}

		return nil, fmt.Errorf("unable to grant cross account read: %w", err)
	}

	if err := a.persistInstallPhoneHomeAuth(ctx, &install, secret); err != nil {
		return nil, err
	}

	a.l.Info("reconciled install phone home secret",
		zap.String("install_id", install.ID),
		zap.String("secret_arn", secret.ARN),
		zap.Int("tokens", len(tokens)),
		zap.Int("minted", resp.TokensMinted),
		zap.Int("revoked", resp.TokensRevoked),
		zap.Bool("secret_written", resp.SecretWritten),
	)

	return resp, nil
}

// why: phoneHomeSecretSkipReason returns a non-empty reason when this install must be
// passed over. Every one of these is a clean no-op rather than an error: the feature
// is opt-in per org, and a miss must never fail a provision.
//
// Note the two cloud axes are distinct and easy to conflate. Whether the *install*
// is on AWS decides which secret backend the customer's Lambda can read; whether
// *Nuon* is on AWS only decides how ctl-api authenticates to it, and is handled
// entirely inside Config.ManagementSecretsCreds. Gating on cfg.IsAWS() would
// silently disable phone-home auth for every AWS install on a GCP-hosted control
// plane.
func (a *Activities) phoneHomeSecretSkipReason(
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

	if install.SandboxMode.Bool || install.Org.SandboxMode {
		return phoneHomeSkipSandboxMode, nil
	}

	if install.AWSAccount == nil {
		return phoneHomeSkipNotAWS, nil
	}

	if install.CloudPlatformMetadata.TargetAccountID == "" {
		a.l.Info("skipping phone home secret: no target account",
			zap.String("install_id", install.ID), zap.String("org_id", install.OrgID))

		return phoneHomeSkipNoTargetAccount, nil
	}

	if a.cfg.ManagementSecretsCreds() == nil {
		a.l.Warn("skipping phone home secret: no path to management secrets manager",
			zap.String("install_id", install.ID),
			zap.String("cloud_provider", a.cfg.CloudProvider),
		)

		return phoneHomeSkipNoManagement, nil
	}

	return "", nil
}

func (a *Activities) reconcileStackVersionPhoneHomeTokens(
	ctx context.Context, install *app.Install, resp *EnsureInstallPhoneHomeSecretResponse,
) (map[string]string, error) {
	tokens := map[string]string{}

	if install.InstallStack == nil {
		return tokens, nil
	}

	live, err := a.livePhoneHomeTokens(ctx, install.InstallStack.InstallStackVersions)
	if err != nil {
		return nil, err
	}

	for i := range install.InstallStack.InstallStackVersions {
		version := &install.InstallStack.InstallStackVersions[i]

		if version.PhoneHomeID == "" {
			continue
		}

		if !version.PhoneHomeTokenEligible() {
			if version.PhoneHomeTokenID != "" {
				if err := a.revokePhoneHomeToken(ctx, version); err != nil {
					return nil, err
				}
				resp.TokensRevoked++
			}

			continue
		}

		if value, ok := live[version.PhoneHomeTokenID]; ok {
			tokens[version.PhoneHomeID] = value

			continue
		}

		value, err := a.mintPhoneHomeToken(ctx, version)
		if err != nil {
			return nil, err
		}
		tokens[version.PhoneHomeID] = value
		resp.TokensMinted++
	}

	return tokens, nil
}

func (a *Activities) livePhoneHomeTokens(
	ctx context.Context, versions []app.InstallStackVersion,
) (map[string]string, error) {
	ids := make([]string, 0, len(versions))
	for i := range versions {
		if versions[i].PhoneHomeTokenID != "" {
			ids = append(ids, versions[i].PhoneHomeTokenID)
		}
	}

	live := map[string]string{}
	if len(ids) == 0 {
		return live, nil
	}

	var rows []app.Token
	if res := a.db.WithContext(ctx).
		Where("id IN ?", ids).
		Where("expires_at > ?", time.Now()).
		Find(&rows); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	for _, row := range rows {
		live[row.ID] = row.Token
	}

	return live, nil
}

func (a *Activities) mintPhoneHomeToken(ctx context.Context, version *app.InstallStackVersion) (string, error) {
	if _, err := a.acctClient.EnsureServiceAccount(ctx, version.ID, ""); err != nil {
		return "", fmt.Errorf("unable to ensure stack version service account: %w", err)
	}

	token, err := a.acctClient.CreateToken(ctx, account.ServiceAccountEmail(version.ID), phoneHomeTokenTimeout)
	if err != nil {
		return "", fmt.Errorf("unable to create phone home token: %w", err)
	}

	if res := a.db.WithContext(ctx).
		Model(&app.InstallStackVersion{ID: version.ID}).
		Update("phone_home_token_id", token.ID); res.Error != nil {
		return "", generics.TemporalGormError(res.Error)
	}
	version.PhoneHomeTokenID = token.ID

	return token.Token, nil
}

func (a *Activities) revokePhoneHomeToken(ctx context.Context, version *app.InstallStackVersion) error {
	if res := a.db.WithContext(ctx).
		Where(app.Token{ID: version.PhoneHomeTokenID}).
		Delete(&app.Token{}); res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}

	if res := a.db.WithContext(ctx).
		Model(&app.InstallStackVersion{ID: version.ID}).
		Update("phone_home_token_id", ""); res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}
	version.PhoneHomeTokenID = ""

	return nil
}

func (a *Activities) persistInstallPhoneHomeAuth(
	ctx context.Context, install *app.Install, secret *secretsmanager.EnsureSecretOutput,
) error {
	auth := app.PhoneHomeAuth{CreatedAt: time.Now()}
	if install.PhoneHomeAuth != nil {
		auth = *install.PhoneHomeAuth
	}

	if auth.SecretARN == secret.ARN &&
		auth.SecretRegion == secret.Region &&
		auth.KMSKeyARN == a.cfg.AWSPhoneHomeCMKARN {
		return nil
	}

	auth.SecretARN = secret.ARN
	auth.SecretRegion = secret.Region
	auth.KMSKeyARN = a.cfg.AWSPhoneHomeCMKARN

	if res := a.db.WithContext(ctx).
		Model(&app.Install{ID: install.ID}).
		Update("phone_home_auth", auth); res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}
	install.PhoneHomeAuth = &auth

	return nil
}
