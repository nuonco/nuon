package activities

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// The gate is what stops the operator-initiated backfill from becoming "provision
// everything unconditionally". Waiving the org feature flag must waive that and
// nothing else.
//
// Every case here runs with a nil features client on purpose: consulting it would
// panic, so these also prove the bypass never reaches for it.
func TestPhoneHomeSecretSkipReasonIgnoringFeatureGate(t *testing.T) {
	const targetAccount = "123456789012"

	reachableCfg := &internal.Config{
		ManagementRegion:     "us-west-2",
		ManagementIAMRoleARN: "arn:aws:iam::766121324316:role/ctl-api-management",
	}

	for _, tc := range []struct {
		name    string
		cfg     *internal.Config
		install *app.Install
		want    string
	}{
		{
			name: "a non-aws install is still skipped",
			cfg:  reachableCfg,
			install: &app.Install{
				ID:                    "inst",
				CloudPlatformMetadata: app.CloudPlatformMetadata{TargetAccountID: targetAccount},
			},
			want: phoneHomeSkipNotAWS,
		},
		{
			// The one that matters most: target_account_id is only required at
			// creation once the flag is on, so an org that never had the flag is
			// mostly full of installs that skip here regardless of the bypass.
			name:    "an install with no target account is still skipped",
			cfg:     reachableCfg,
			install: &app.Install{ID: "inst", AWSAccount: &app.AWSAccount{}},
			want:    phoneHomeSkipNoTargetAccount,
		},
		{
			name: "an unreachable management account is still skipped",
			cfg:  &internal.Config{},
			install: &app.Install{
				ID:                    "inst",
				AWSAccount:            &app.AWSAccount{},
				CloudPlatformMetadata: app.CloudPlatformMetadata{TargetAccountID: targetAccount},
			},
			want: phoneHomeSkipNoManagement,
		},
		{
			name: "a sandbox install is still skipped",
			cfg:  reachableCfg,
			install: &app.Install{
				ID:                    "inst",
				SandboxMode:           sql.NullBool{Bool: true, Valid: true},
				AWSAccount:            &app.AWSAccount{},
				CloudPlatformMetadata: app.CloudPlatformMetadata{TargetAccountID: targetAccount},
			},
			want: phoneHomeSkipSandboxMode,
		},
		{
			// An install created before AdminForceSandboxMode flipped the org keeps an
			// explicit false that Install.AfterQuery will not override.
			name: "an install in a sandboxed org is skipped even when its own flag is false",
			cfg:  reachableCfg,
			install: &app.Install{
				ID:                    "inst",
				SandboxMode:           sql.NullBool{Bool: false, Valid: true},
				Org:                   app.Org{SandboxMode: true},
				AWSAccount:            &app.AWSAccount{},
				CloudPlatformMetadata: app.CloudPlatformMetadata{TargetAccountID: targetAccount},
			},
			want: phoneHomeSkipSandboxMode,
		},
		{
			name: "an otherwise eligible install proceeds even with the flag off",
			cfg:  reachableCfg,
			install: &app.Install{
				ID:                    "inst",
				AWSAccount:            &app.AWSAccount{},
				CloudPlatformMetadata: app.CloudPlatformMetadata{TargetAccountID: targetAccount},
			},
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &Activities{cfg: tc.cfg, l: zap.NewNop()}

			got, err := a.phoneHomeSecretSkipReason(context.Background(), tc.install, true)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
