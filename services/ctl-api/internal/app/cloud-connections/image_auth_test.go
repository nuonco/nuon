package cloudconnections

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func TestImageConnectionModes(t *testing.T) {
	const garImage, acrImage = "us-central1-docker.pkg.dev/acme/repo/image", "acme.azurecr.io/repo"
	for name, tc := range map[string]struct {
		ecr       *app.AWSECRImageConfig
		gar       *app.GCPGARImageConfig
		acr       *app.AzureACRImageConfig
		image     string
		legacy    bool
		wantError string
		wantIs    error
	}{
		"legacy ECR":                    {ecr: &app.AWSECRImageConfig{IAMRoleARN: "arn:aws:iam::123456789012:role/acme"}, legacy: true},
		"legacy ECR disabled":           {ecr: &app.AWSECRImageConfig{IAMRoleARN: "arn:aws:iam::123456789012:role/acme"}, wantIs: ErrLegacyImageAuthDisabled},
		"legacy GAR impersonation":      {gar: &app.GCPGARImageConfig{ServiceAccountEmail: "acme@example.com"}, legacy: true, wantIs: ErrLegacyImageAuthDisabled},
		"legacy GAR federation":         {gar: &app.GCPGARImageConfig{WorkloadIdentityProvider: "provider"}, legacy: true, wantIs: ErrLegacyImageAuthDisabled},
		"legacy GAR disabled":           {gar: &app.GCPGARImageConfig{ServiceAccountEmail: "acme@example.com"}, wantIs: ErrLegacyImageAuthDisabled},
		"saved ambient GAR":             {gar: &app.GCPGARImageConfig{ID: "saved"}, legacy: true},
		"saved ambient GAR disabled":    {gar: &app.GCPGARImageConfig{ID: "saved"}, wantIs: ErrLegacyImageAuthDisabled},
		"saved ambient ACR":             {acr: &app.AzureACRImageConfig{ID: "saved"}, legacy: true},
		"saved ambient ACR disabled":    {acr: &app.AzureACRImageConfig{ID: "saved"}, wantIs: ErrLegacyImageAuthDisabled},
		"unmatched GAR image":           {gar: &app.GCPGARImageConfig{GCPRegion: "us-east1", GCPProjectID: "acme"}, image: garImage, legacy: true},
		"unmatched GAR image disabled":  {gar: &app.GCPGARImageConfig{GCPRegion: "us-east1", GCPProjectID: "acme"}, image: garImage, wantError: "must match the configured project"},
		"unmatched ACR image":           {acr: &app.AzureACRImageConfig{RegistryURL: "other.azurecr.io"}, image: acrImage, legacy: true},
		"unmatched ACR image disabled":  {acr: &app.AzureACRImageConfig{RegistryURL: "other.azurecr.io"}, image: acrImage, wantError: "must match registry_url"},
		"ACR app registration":          {acr: &app.AzureACRImageConfig{TenantID: "tenant", ClientID: "client", ClientSecretName: "secret"}, legacy: true},
		"ACR app registration disabled": {acr: &app.AzureACRImageConfig{TenantID: "tenant", ClientID: "client", ClientCertificateName: "cert"}, wantIs: ErrLegacyImageAuthDisabled},
		"mixed ECR":                     {ecr: &app.AWSECRImageConfig{IAMRoleARN: "role", CloudConnectionID: "pinned"}, legacy: true, wantError: "cannot be combined"},
		"mixed GAR":                     {gar: &app.GCPGARImageConfig{ServiceAccountEmail: "acme@example.com", CloudConnectionID: "pinned"}, legacy: true, wantError: "cannot be combined"},
		"mixed GAR federation":          {gar: &app.GCPGARImageConfig{WorkloadIdentityProvider: "provider", CloudConnectionID: "pinned"}, legacy: true, wantError: "cannot be combined"},
		"mixed ACR":                     {acr: &app.AzureACRImageConfig{CloudConnectionID: "pinned", ClientID: "client"}, legacy: true, wantError: "cannot be combined"},
		"missing ECR pin":               {ecr: &app.AWSECRImageConfig{ID: "saved"}, legacy: true, wantError: "no pinned cloud connection"},
	} {
		t.Run(name, func(t *testing.T) {
			var conn *app.CloudConnection
			var err error
			switch {
			case tc.ecr != nil:
				conn, err = ECRConnection(context.Background(), nil, "org_acme", tc.ecr, tc.image, tc.legacy)
			case tc.gar != nil:
				conn, err = GARConnection(context.Background(), nil, "org_acme", tc.gar, tc.image, tc.legacy)
			case tc.acr != nil:
				conn, err = ACRConnection(context.Background(), nil, "org_acme", tc.acr, tc.image, tc.legacy)
			}
			require.Nil(t, conn)
			switch {
			case tc.wantIs != nil:
				require.ErrorIs(t, err, tc.wantIs)
			case tc.wantError != "":
				require.ErrorContains(t, err, tc.wantError)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestImageConnectionPins(t *testing.T) {
	for name, tc := range map[string]struct {
		pin        string
		queryError error
		legacy     bool
		ambient    bool
		mixedCase  bool
		wantError  string
	}{
		"resolve mapping":                   {},
		"resolve mapping with legacy":       {legacy: true},
		"keep exact pin":                    {pin: "cc_saved"},
		"mixed case registry":               {mixedCase: true},
		"missing mapping":                   {queryError: gorm.ErrRecordNotFound, wantError: "no saved image repository mapping"},
		"missing mapping with legacy":       {queryError: gorm.ErrRecordNotFound, legacy: true, ambient: true, wantError: "no saved image repository mapping"},
		"failed mapping lookup with legacy": {queryError: errors.New("database unavailable"), legacy: true, wantError: "database unavailable"},
		"deleted or unverified pin":         {pin: "cc_saved", queryError: gorm.ErrRecordNotFound, legacy: true, wantError: "not verified"},
		"failed pin lookup":                 {pin: "cc_saved", queryError: errors.New("database unavailable"), legacy: true, wantError: "database unavailable"},
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
					cfg.GCPGARImageConfig = &app.GCPGARImageConfig{CloudConnectionID: tc.pin, GCPRegion: "us-central1", GCPProjectID: "acme"}
				case "acr":
					cfg.ImageURL = "acme.azurecr.io/repo"
					registryURL := "acme.azurecr.io"
					if tc.mixedCase {
						registryURL = "Acme.azurecr.io"
					}
					cfg.AzureACRImageConfig = &app.AzureACRImageConfig{CloudConnectionID: tc.pin, RegistryURL: registryURL}
				}
				ccc := &app.ComponentConfigConnection{Checksum: "original", ExternalImageComponentConfig: cfg}
				err = ResolveImageConnection(context.Background(), db, "org_acme", ccc, tc.legacy)
				switch {
				case tc.ambient && platform != "ecr":
					require.NoError(t, err)
					require.Equal(t, "original", ccc.Checksum)
				case tc.wantError != "":
					require.ErrorContains(t, err, tc.wantError)
					require.Equal(t, "original", ccc.Checksum)
				default:
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

func TestImageConnectionErrorClassification(t *testing.T) {
	for name, tc := range map[string]struct {
		pin        string
		iamRole    bool
		legacy     bool
		queryError error
		wantError  bool
		wantUser   bool
	}{
		"missing mapping":           {queryError: gorm.ErrRecordNotFound, wantUser: true},
		"deleted or unverified pin": {pin: "cc_saved", queryError: gorm.ErrRecordNotFound, wantUser: true},
		"failed mapping lookup":     {queryError: errors.New("database unavailable"), wantError: true},
		"failed pin lookup":         {pin: "cc_saved", queryError: errors.New("database unavailable"), wantError: true},
		"legacy disabled":           {iamRole: true, wantUser: true},
		"legacy enabled":            {iamRole: true, legacy: true},
	} {
		t.Run(name, func(t *testing.T) {
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
			require.NoError(t, err)
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("classification", func(tx *gorm.DB) {
				if tc.queryError != nil {
					tx.AddError(tc.queryError)
					return
				}
				switch dest := tx.Statement.Dest.(type) {
				case *app.ImageRepository:
					dest.CloudConnectionID = "cc_mapping"
				case *app.CloudConnection:
					*dest = app.CloudConnection{ID: "cc_mapping", OrgID: "org_acme", Status: app.CloudConnectionStatusVerified}
				}
				tx.RowsAffected = 1
			}))
			ecrCfg := &app.AWSECRImageConfig{AWSRegion: "us-east-1"}
			if tc.iamRole {
				ecrCfg.IAMRoleARN = "arn:aws:iam::123456789012:role/acme"
			} else {
				ecrCfg.CloudConnectionID = tc.pin
			}
			ccc := &app.ComponentConfigConnection{ExternalImageComponentConfig: &app.ExternalImageComponentConfig{
				ImageURL:          "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme",
				AWSECRImageConfig: ecrCfg,
			}}
			err = ResolveImageConnection(context.Background(), db, "org_acme", ccc, tc.legacy)
			var invalidReq stderr.ErrInvalidRequest
			if tc.wantUser {
				require.ErrorAs(t, err, &invalidReq)
			} else if tc.wantError {
				require.Error(t, err)
				require.NotErrorAs(t, err, &invalidReq)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
