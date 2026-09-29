package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testManagementRoleARN = "arn:aws:iam::766121324316:role/ctl-api-management"
	testPhoneHomeRoleARN  = "arn:aws:iam::766121324316:role/ctl-api-phone-home-secrets"
)

func TestManagementSecretsCreds(t *testing.T) {
	for _, tc := range []struct {
		name        string
		cfg         Config
		wantNil     bool
		wantRoleARN string
		wantGCPOIDC bool
	}{
		{
			name: "aws hosted falls back to the management role",
			cfg: Config{
				CloudProvider:        "aws",
				ManagementRegion:     "us-west-2",
				ManagementIAMRoleARN: testManagementRoleARN,
			},
			wantRoleARN: testManagementRoleARN,
			wantGCPOIDC: false,
		},
		{
			name: "an explicit phone home role wins over the management role",
			cfg: Config{
				CloudProvider:              "aws",
				ManagementRegion:           "us-west-2",
				ManagementIAMRoleARN:       testManagementRoleARN,
				AWSPhoneHomeSecretsRoleARN: testPhoneHomeRoleARN,
			},
			wantRoleARN: testPhoneHomeRoleARN,
			wantGCPOIDC: false,
		},
		{
			name: "gcp hosted federates into the nuon aws account",
			cfg: Config{
				CloudProvider:              "gcp",
				ManagementRegion:           "us-west-2",
				AWSPhoneHomeSecretsRoleARN: testPhoneHomeRoleARN,
			},
			wantRoleARN: testPhoneHomeRoleARN,
			wantGCPOIDC: true,
		},
		{
			name: "gcp hosted without a configured role has no path",
			cfg: Config{
				CloudProvider:        "gcp",
				ManagementRegion:     "us-west-2",
				ManagementIAMRoleARN: testManagementRoleARN,
			},
			wantNil: true,
		},
		{
			name: "azure hosted has no federation path to aws",
			cfg: Config{
				CloudProvider:    "azure",
				ManagementRegion: "us-west-2",
			},
			wantNil: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			creds := tc.cfg.ManagementSecretsCreds()

			if tc.wantNil {
				assert.Nil(t, creds)
				return
			}

			require.NotNil(t, creds)
			require.NotNil(t, creds.AssumeRole)
			assert.Equal(t, tc.wantRoleARN, creds.AssumeRole.RoleARN)
			assert.Equal(t, tc.wantGCPOIDC, creds.AssumeRole.UseGCPOIDC)
			assert.Equal(t, tc.cfg.ManagementRegion, creds.Region)
		})
	}
}
