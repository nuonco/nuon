package components

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/syncerr"
)

func TestUnchangedLegacyImageSync(t *testing.T) {
	for name, tc := range map[string]struct {
		image *config.ExternalImageComponentConfig
	}{
		"ECR role":        {image: &config.ExternalImageComponentConfig{AWSECRImageConfig: &config.AWSECRConfig{IAMRoleARN: "arn:aws:iam::123456789012:role/acme", AWSRegion: "us-east-1", ImageURL: "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme", Tag: "v1"}}},
		"GAR account":     {image: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{ServiceAccountEmail: "acme@example.com", GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}}},
		"GAR ambient":     {image: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}}},
		"GAR federation":  {image: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{WorkloadIdentityProvider: "provider", GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}}},
		"ACR ambient":     {image: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1"}}},
		"ACR secret":      {image: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1", TenantID: "tenant", ClientID: "client", ClientSecretName: "secret"}}},
		"ACR certificate": {image: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1", TenantID: "tenant", ClientID: "client", ClientCertificateName: "cert"}}},
	} {
		t.Run(name, func(t *testing.T) {
			component := &config.Component{Name: "acme", Type: config.ContainerImageComponentType, ExternalImage: tc.image}
			checksum, err := build.ComponentChecksum(component)
			require.NoError(t, err)
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
			require.NoError(t, err)
			db = db.Session(&gorm.Session{SkipHooks: true})
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("legacy-sync", func(tx *gorm.DB) {
				switch dest := tx.Statement.Dest.(type) {
				case *app.Component:
					*dest = app.Component{ID: "component", Name: "acme", OrgID: "org_acme"}
				case *app.AppConfig:
					*dest = app.AppConfig{ID: "appconfig"}
				case *app.ComponentConfigConnection:
					*dest = app.ComponentConfigConnection{ID: "saved_config", Checksum: checksum, LatestBuildID: generics.NewNullString("saved_build")}
				case *app.ComponentBuild:
					*dest = app.ComponentBuild{ID: "saved_build", Status: app.ComponentBuildStatusActive}
				case *app.ImageRepository:
					tx.AddError(gorm.ErrRecordNotFound)
					return
				default:
					t.Fatalf("unexpected lookup for legacy sync: %T", dest)
				}
				tx.RowsAffected = 1
			}))
			require.NoError(t, db.Callback().Create().Before("gorm:create").Register("no-new-config", func(tx *gorm.DB) {
				t.Fatal("unchanged sync must reuse the saved config without creating a config or build")
			}))
			state := &sync.State{}
			require.NoError(t, SyncComponent(context.Background(), SyncComponentParams{DB: db, Component: component, AppID: "app", AppConfigID: "appconfig", State: state, DispatchBuilds: true, LegacyImageAuth: true}))
			require.Len(t, state.Components, 1)
			require.Equal(t, "saved_config", state.Components[0].ConfigID)
			require.Nil(t, state.Result)
		})
	}
}

func TestImageConnectionSyncErrorClassification(t *testing.T) {
	for name, tc := range map[string]struct {
		resolveErr   error
		wantSyncErr  bool
		wantInternal bool
	}{
		"user error":     {resolveErr: cloudconnections.ErrLegacyImageAuthDisabled, wantSyncErr: true},
		"database error": {resolveErr: errors.New("database unavailable"), wantInternal: true},
	} {
		t.Run(name, func(t *testing.T) {
			err := syncerr.From("component-acme", "invalid image repository cloud connection", tc.resolveErr)
			var syncErr sync.SyncErr
			var internalErr sync.SyncInternalErr
			require.Equal(t, tc.wantSyncErr, errors.As(err, &syncErr))
			require.Equal(t, tc.wantInternal, errors.As(err, &internalErr))
		})
	}
}
