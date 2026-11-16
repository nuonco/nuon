package helpers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func TestLegacyECRCredentialsConfig(t *testing.T) {
	for name, tc := range map[string]struct {
		cloud          string
		managementRole string
		twoStep        bool
		gcpOIDC        bool
	}{
		"AWS management chain": {managementRole: "arn:aws:iam::111111111111:role/acme-management", twoStep: true},
		"AWS direct":           {},
		"GCP OIDC":             {cloud: "gcp", managementRole: "unused", gcpOIDC: true},
		"Azure":                {cloud: "azure", managementRole: "unused"},
	} {
		t.Run(name, func(t *testing.T) {
			cfg := legacyECRCredentialsConfig(
				&internal.Config{CloudProvider: tc.cloud, ManagementIAMRoleARN: tc.managementRole},
				&app.AWSECRImageConfig{AWSRegion: "us-east-1", IAMRoleARN: "arn:aws:iam::123456789012:role/acme"},
			)
			require.Equal(t, "us-east-1", cfg.Region)
			require.Equal(t, "arn:aws:iam::123456789012:role/acme", cfg.AssumeRole.RoleARN)
			require.Equal(t, tc.gcpOIDC, cfg.AssumeRole.UseGCPOIDC)
			if tc.twoStep {
				require.Equal(t, tc.managementRole, cfg.AssumeRole.TwoStepConfig.IAMRoleARN)
			} else {
				require.Nil(t, cfg.AssumeRole.TwoStepConfig)
			}
		})
	}
}

func TestLegacyImageAuthorizationDisabled(t *testing.T) {
	for name, authorize := range map[string]func(*Helpers) error{
		"ECR role": func(h *Helpers) error {
			_, err := h.ECRAuthorization(context.Background(), "org_acme", &app.AWSECRImageConfig{ID: "saved", AWSRegion: "us-east-1", IAMRoleARN: "arn:aws:iam::123456789012:role/acme"}, "image")
			return err
		},
		"GAR impersonation": func(h *Helpers) error {
			_, err := h.GARAuthorization(context.Background(), "org_acme", &app.GCPGARImageConfig{ID: "saved", ServiceAccountEmail: "acme@example.com"}, "image")
			return err
		},
		"GAR ambient": func(h *Helpers) error {
			_, err := h.GARAuthorization(context.Background(), "org_acme", &app.GCPGARImageConfig{ID: "saved"}, "image")
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			h := &Helpers{cfg: &internal.Config{}, imageAuth: &sharedactivities.Activities{}}
			require.ErrorIs(t, authorize(h), cloudconnections.ErrLegacyImageAuthDisabled)
		})
	}
}

func TestConnectionAuthorizationNeverFallsBack(t *testing.T) {
	for name, tc := range map[string]struct {
		queryError error
		want       string
	}{
		"deleted or unverified": {queryError: gorm.ErrRecordNotFound, want: "not verified"},
		"lookup failed":         {queryError: errors.New("database unavailable"), want: "database unavailable"},
		"issuer unavailable":    {want: "OIDC issuer is unavailable"},
	} {
		for _, platform := range []string{"ecr", "gar", "acr"} {
			t.Run(platform+"/"+name, func(t *testing.T) {
				db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
				require.NoError(t, err)
				require.NoError(t, db.Callback().Query().After("gorm:query").Register("pinned", func(tx *gorm.DB) {
					require.Contains(t, tx.Statement.Vars, "cc_saved")
					require.Contains(t, tx.Statement.Vars, "org_acme")
					if tc.queryError != nil {
						tx.AddError(tc.queryError)
						return
					}
					connection := tx.Statement.Dest.(*app.CloudConnection)
					*connection = app.CloudConnection{ID: "cc_saved", OrgID: "org_acme", Platform: app.CloudPlatformAWS, Status: app.CloudConnectionStatusVerified}
					if platform == "acr" {
						connection.Platform, connection.Preset = app.CloudPlatformAzure, app.CloudConnectionPresetACR
						connection.TargetID, connection.Principal, connection.AzureTenantID = "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000003"
						connection.AzureRegistryID = "/subscriptions/" + connection.TargetID + "/resourceGroups/acme/providers/Microsoft.ContainerRegistry/registries/acmeregistry"
						connection.AzureRegistryLoginServer = "acme.azurecr.io"
					}
					tx.RowsAffected = 1
				}))
				h := &Helpers{cfg: &internal.Config{LegacyImageAuthEnabled: true}, db: db, imageAuth: &sharedactivities.Activities{}}
				switch platform {
				case "ecr":
					_, err = h.ECRAuthorization(context.Background(), "org_acme", &app.AWSECRImageConfig{ID: "saved", CloudConnectionID: "cc_saved", AWSRegion: "us-east-1"}, "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme")
				case "gar":
					_, err = h.GARAuthorization(context.Background(), "org_acme", &app.GCPGARImageConfig{ID: "saved", CloudConnectionID: "cc_saved", GCPRegion: "us-central1", GCPProjectID: "acme"}, "us-central1-docker.pkg.dev/acme/repo/image")
				case "acr":
					_, err = h.ACRAuthorization(context.Background(), "org_acme", &app.AzureACRImageConfig{ID: "saved", CloudConnectionID: "cc_saved", RegistryURL: "acme.azurecr.io"}, "acme.azurecr.io/repo")
				}
				require.ErrorContains(t, err, tc.want)
			})
		}
	}
}
