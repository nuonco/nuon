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

// why: Waiving the org feature flags must waive those and nothing else. Every case runs with a
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
			name: "an otherwise eligible install proceeds even with the flag off",
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

func TestAzurePhoneHomeSkipReasonsAreStable(t *testing.T) {
	assert.Equal(t, "install has no target subscription id", phoneHomeSkipNoSubscription)
}

func TestAzureUsesTheSharedPhoneHomeAuthFlag(t *testing.T) {
	assert.Equal(t, app.OrgFeature("phone-home-auth"), app.OrgFeaturePhoneHomeAuth)
}
