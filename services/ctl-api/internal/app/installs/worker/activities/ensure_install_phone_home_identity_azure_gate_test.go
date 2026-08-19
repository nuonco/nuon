package activities

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Waiving the org feature flags must waive those and nothing else. Every case runs with a
// nil features client on purpose: consulting it would panic, so these also prove the
// bypass never reaches for it.
func TestAzurePhoneHomeSkipReasonIgnoringFeatureGate(t *testing.T) {
	const targetSubscription = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	for _, tc := range []struct {
		name    string
		install *app.Install
		want    string
	}{
		{
			// Without it the verifier would have nothing to bind the token's
			// subscription against, so rendering an identity would enforce nothing.
			name:    "an install with no target subscription is skipped",
			install: &app.Install{ID: "inst", AzureAccount: &app.AzureAccount{}},
			want:    phoneHomeSkipNoSubscription,
		},
		{
			name: "a sandbox install is skipped",
			install: &app.Install{
				ID:                    "inst",
				SandboxMode:           sql.NullBool{Bool: true, Valid: true},
				AzureAccount:          &app.AzureAccount{},
				CloudPlatformMetadata: app.CloudPlatformMetadata{TargetSubscriptionID: targetSubscription},
			},
			want: phoneHomeSkipSandboxMode,
		},
		{
			// An install created before AdminForceSandboxMode flipped the org keeps an
			// explicit false that Install.AfterQuery will not override.
			name: "an install in a sandboxed org is skipped even when its own flag is false",
			install: &app.Install{
				ID:                    "inst",
				SandboxMode:           sql.NullBool{Bool: false, Valid: true},
				Org:                   app.Org{SandboxMode: true},
				AzureAccount:          &app.AzureAccount{},
				CloudPlatformMetadata: app.CloudPlatformMetadata{TargetSubscriptionID: targetSubscription},
			},
			want: phoneHomeSkipSandboxMode,
		},
		{
			name: "an otherwise eligible install proceeds even with the flags off",
			install: &app.Install{
				ID:                    "inst",
				AzureAccount:          &app.AzureAccount{},
				CloudPlatformMetadata: app.CloudPlatformMetadata{TargetSubscriptionID: targetSubscription},
			},
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &Activities{l: zap.NewNop()}

			got, err := a.azurePhoneHomeSkipReason(context.Background(), tc.install, true)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// Skip reasons are metric tags and the backfill's progress output, so drift is silent.
func TestAzurePhoneHomeSkipReasonsAreStable(t *testing.T) {
	assert.Equal(t, "org feature phone-home-auth-azure is disabled", phoneHomeSkipAzureDisabled)
	assert.Equal(t, "install has no target subscription id", phoneHomeSkipNoSubscription)
}
