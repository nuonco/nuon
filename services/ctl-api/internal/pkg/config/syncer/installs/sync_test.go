package installs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// existingToConfig feeds both the drift diff and the immutability check, so a missing
// identifier here does not fail loudly — it presents as an install that is forever out
// of sync with a config no update can satisfy.
func TestExistingToConfigEchoesTargetIdentifiers(t *testing.T) {
	t.Run("aws reads the target account from cloud platform metadata", func(t *testing.T) {
		cfg := existingToConfig(&app.Install{
			Name:                  "inst",
			AWSAccount:            &app.AWSAccount{Region: "us-west-2"},
			CloudPlatformMetadata: app.CloudPlatformMetadata{TargetAccountID: "123456789012"},
		})

		require.NotNil(t, cfg.AWSAccount)
		assert.Equal(t, "us-west-2", cfg.AWSAccount.Region)
		assert.Equal(t, "123456789012", cfg.AWSAccount.AccountID)
	})

	t.Run("gcp prefers the target project over the account record", func(t *testing.T) {
		cfg := existingToConfig(&app.Install{
			GCPAccount:            &app.GCPAccount{ProjectID: "legacy-proj", Region: "us-central1"},
			CloudPlatformMetadata: app.CloudPlatformMetadata{TargetProjectID: "target-proj"},
		})

		require.NotNil(t, cfg.GCPAccount)
		assert.Equal(t, "target-proj", cfg.GCPAccount.ProjectID)
	})

	t.Run("azure prefers the target subscription over the account record", func(t *testing.T) {
		cfg := existingToConfig(&app.Install{
			AzureAccount:          &app.AzureAccount{Location: "eastus", SubscriptionID: "legacy-sub"},
			CloudPlatformMetadata: app.CloudPlatformMetadata{TargetSubscriptionID: "target-sub"},
		})

		require.NotNil(t, cfg.AzureAccount)
		assert.Equal(t, "target-sub", cfg.AzureAccount.SubscriptionID)
	})

	// Installs predating CloudPlatformMetadata have an empty target, and Azure/GCP
	// carry the identifier on the account record instead.
	t.Run("falls back to the account record when no target is set", func(t *testing.T) {
		cfg := existingToConfig(&app.Install{
			GCPAccount:   &app.GCPAccount{ProjectID: "legacy-proj"},
			AzureAccount: &app.AzureAccount{Location: "eastus", SubscriptionID: "legacy-sub"},
		})

		require.NotNil(t, cfg.GCPAccount)
		require.NotNil(t, cfg.AzureAccount)
		assert.Equal(t, "legacy-proj", cfg.GCPAccount.ProjectID)
		assert.Equal(t, "legacy-sub", cfg.AzureAccount.SubscriptionID)
	})
}

// The round trip is what actually matters: a config declaring the same account the
// install already targets must report no drift, and a changed one must be refused
// rather than silently dropped.
func TestExistingToConfigRoundTripsWithoutDrift(t *testing.T) {
	existing := &app.Install{
		Name:                  "inst",
		AWSAccount:            &app.AWSAccount{Region: "us-west-2"},
		CloudPlatformMetadata: app.CloudPlatformMetadata{TargetAccountID: "123456789012"},
	}
	upstream := existingToConfig(existing)

	same := &config.Install{
		Name:       "inst",
		AWSAccount: &config.AWSAccount{Region: "us-west-2", AccountID: "123456789012"},
	}
	require.NoError(t, same.CheckImmutableTargetAccount(upstream))

	d, err := same.Diff(upstream)
	require.NoError(t, err)
	assert.False(t, d.Summary().HasChanged, "an unchanged account id must not report drift")

	changed := &config.Install{
		Name:       "inst",
		AWSAccount: &config.AWSAccount{Region: "us-west-2", AccountID: "999999999999"},
	}
	err = changed.CheckImmutableTargetAccount(upstream)
	require.Error(t, err, "a changed account id must be refused, not silently ignored")
	assert.Contains(t, err.Error(), "aws_account.account_id")
	assert.Contains(t, err.Error(), "immutable")
}
