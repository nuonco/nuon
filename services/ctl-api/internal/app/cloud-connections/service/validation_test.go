package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestValidateConnection(t *testing.T) {
	for _, tc := range []struct {
		name      string
		platform  app.CloudPlatform
		preset    app.CloudConnectionPreset
		principal string
		wantErr   string
	}{
		{"stacks", app.CloudPlatformAWS, app.CloudConnectionPresetStacks, "arn:aws:iam::123456789012:role/acme", ""},
		{"custom", app.CloudPlatformAWS, app.CloudConnectionPresetCustom, "arn:aws:iam::123456789012:role/path/acme", ""},
		{"azure rejected", app.CloudPlatformAzure, app.CloudConnectionPresetStacks, "", "only aws"},
		{"gcp rejected", app.CloudPlatformGCP, app.CloudConnectionPresetCustom, "", "only aws"},
		{"preset required", app.CloudPlatformAWS, "", "", "preset must be"},
		{"images rejected", app.CloudPlatformAWS, "images", "", "preset must be"},
		{"different account", app.CloudPlatformAWS, app.CloudConnectionPresetStacks, "arn:aws:iam::210987654321:role/acme", "IAM role ARN in account"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			connection := app.CloudConnection{Name: "acme", Platform: tc.platform, Preset: tc.preset, TargetID: "123456789012", Principal: tc.principal}
			err := validateConnection(&connection)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "us-east-1", connection.DefaultRegion)
		})
	}
}
