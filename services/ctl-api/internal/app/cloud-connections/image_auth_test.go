package cloudconnections

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestImageConnectionModes(t *testing.T) {
	for name, tc := range map[string]struct {
		ecr       *app.AWSECRImageConfig
		gar       *app.GCPGARImageConfig
		acr       *app.AzureACRImageConfig
		wantError string
	}{
		"saved legacy ECR":            {ecr: &app.AWSECRImageConfig{ID: "saved", IAMRoleARN: "arn:aws:iam::123456789012:role/acme"}},
		"saved legacy GAR":            {gar: &app.GCPGARImageConfig{ID: "saved", ServiceAccountEmail: "acme@example.com"}},
		"saved legacy GAR federation": {gar: &app.GCPGARImageConfig{ID: "saved", WorkloadIdentityProvider: "provider"}},
		"saved ambient GAR":           {gar: &app.GCPGARImageConfig{ID: "saved"}},
		"mixed ECR":                   {ecr: &app.AWSECRImageConfig{IAMRoleARN: "role", CloudConnectionID: "pinned"}, wantError: "cannot be combined"},
		"mixed GAR":                   {gar: &app.GCPGARImageConfig{ServiceAccountEmail: "acme@example.com", CloudConnectionID: "pinned"}, wantError: "cannot be combined"},
		"mixed GAR federation":        {gar: &app.GCPGARImageConfig{WorkloadIdentityProvider: "provider", CloudConnectionID: "pinned"}, wantError: "cannot be combined"},
		"mixed GAR opt in account":    {gar: &app.GCPGARImageConfig{UseCloudConnection: true, ServiceAccountEmail: "acme@example.com"}, wantError: "cannot be combined"},
		"mixed GAR opt in provider":   {gar: &app.GCPGARImageConfig{UseCloudConnection: true, WorkloadIdentityProvider: "provider"}, wantError: "cannot be combined"},
		"missing ECR pin":             {ecr: &app.AWSECRImageConfig{ID: "saved"}, wantError: "no pinned cloud connection"},
		"missing GAR pin":             {gar: &app.GCPGARImageConfig{ID: "saved", UseCloudConnection: true}, wantError: "no pinned cloud connection"},
		"missing ACR pin":             {acr: &app.AzureACRImageConfig{ID: "saved", UseCloudConnection: true}, wantError: "no pinned cloud connection"},
		"mixed ACR":                   {acr: &app.AzureACRImageConfig{CloudConnectionID: "pinned", ClientID: "client"}, wantError: "cannot be combined"},
	} {
		t.Run(name, func(t *testing.T) {
			var conn *app.CloudConnection
			var err error
			switch {
			case tc.ecr != nil:
				conn, err = ECRConnection(context.Background(), nil, "org_acme", tc.ecr, "")
			case tc.gar != nil:
				conn, err = GARConnection(context.Background(), nil, "org_acme", tc.gar, "")
			case tc.acr != nil:
				conn, err = ACRConnection(context.Background(), nil, "org_acme", tc.acr, "")
			}
			require.Nil(t, conn)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestImageConnectionPins(t *testing.T) {
	for name, tc := range map[string]struct {
		pin        string
		queryError error
		wantError  string
	}{
		"resolve mapping":           {},
		"keep exact pin":            {pin: "cc_saved"},
		"missing mapping":           {queryError: gorm.ErrRecordNotFound, wantError: "no saved image repository mapping"},
		"deleted or unverified pin": {pin: "cc_saved", queryError: gorm.ErrRecordNotFound, wantError: "not verified"},
		"failed pin lookup":         {pin: "cc_saved", queryError: errors.New("database unavailable"), wantError: "database unavailable"},
	} {
		for _, platform := range []string{"ecr", "gar", "acr"} {
			t.Run(platform+"/"+name, func(t *testing.T) {
				db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
				require.NoError(t, err)
				queries := 0
				require.NoError(t, db.Callback().Query().After("gorm:query").Register("image-auth", func(tx *gorm.DB) {
					queries++
					require.Contains(t, tx.Statement.Vars, "org_acme")
					if tc.queryError != nil {
						tx.AddError(tc.queryError)
						return
					}
					switch dest := tx.Statement.Dest.(type) {
					case *app.ImageRepository:
						require.Empty(t, tc.pin)
						dest.CloudConnectionID = "cc_mapping"
					case *app.CloudConnection:
						pin := tc.pin
						if pin == "" {
							pin = "cc_mapping"
						}
						require.Contains(t, tx.Statement.Vars, pin)
						require.Contains(t, tx.Statement.Vars, app.CloudConnectionStatusVerified)
						*dest = app.CloudConnection{ID: pin, OrgID: "org_acme", Status: app.CloudConnectionStatusVerified}
						if platform == "acr" {
							dest.Platform, dest.Preset = app.CloudPlatformAzure, app.CloudConnectionPresetACR
							dest.TargetID, dest.Principal, dest.AzureTenantID = "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000003"
							dest.AzureRegistryID = "/subscriptions/" + dest.TargetID + "/resourceGroups/acme/providers/Microsoft.ContainerRegistry/registries/acmeregistry"
							dest.AzureRegistryLoginServer = "acme.azurecr.io"
						}
					}
					tx.RowsAffected = 1
				}))
				cfg := &app.ExternalImageComponentConfig{}
				switch platform {
				case "ecr":
					cfg.ImageURL = "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme"
					cfg.AWSECRImageConfig = &app.AWSECRImageConfig{CloudConnectionID: tc.pin, AWSRegion: "us-east-1"}
				case "gar":
					cfg.ImageURL = "us-central1-docker.pkg.dev/acme/repo/image"
					cfg.GCPGARImageConfig = &app.GCPGARImageConfig{UseCloudConnection: tc.pin == "", CloudConnectionID: tc.pin, GCPRegion: "us-central1", GCPProjectID: "acme"}
				case "acr":
					cfg.ImageURL = "acme.azurecr.io/repo"
					cfg.AzureACRImageConfig = &app.AzureACRImageConfig{CloudConnectionID: tc.pin, UseCloudConnection: true, RegistryURL: "acme.azurecr.io"}
				}
				ccc := &app.ComponentConfigConnection{Checksum: "original", ExternalImageComponentConfig: cfg}
				err = ResolveImageConnection(context.Background(), db, "org_acme", ccc)
				if tc.wantError != "" {
					require.ErrorContains(t, err, tc.wantError)
					require.Equal(t, "original", ccc.Checksum)
				} else {
					require.NoError(t, err)
					require.NotEqual(t, "original", ccc.Checksum)
				}
				if tc.pin != "" || tc.queryError != nil {
					require.Equal(t, 1, queries)
				} else {
					require.Equal(t, 2, queries)
				}
			})
		}
	}
}
