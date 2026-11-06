package controlplanejob

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	ecr "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/pkg/oci/metadata"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

type mockImageAuthorizer struct{ mock.Mock }

func (m *mockImageAuthorizer) ECRAuthorization(ctx context.Context, org string, cfg *app.AWSECRImageConfig, image string) (*ecr.Authorization, error) {
	args := m.Called(org, cfg, image)
	return &ecr.Authorization{Username: "AWS", RegistryToken: "acme-token", ServerAddress: "https://123456789012.dkr.ecr.us-east-1.amazonaws.com"}, args.Error(0)
}

func (m *mockImageAuthorizer) GARAuthorization(ctx context.Context, org string, cfg *app.GCPGARImageConfig, image string) (*oauth2.Token, error) {
	args := m.Called(org, cfg, image)
	return &oauth2.Token{AccessToken: "acme-token"}, args.Error(0)
}

func (m *mockImageAuthorizer) ACRAuthorization(ctx context.Context, org string, cfg *app.AzureACRImageConfig, image string) (*metadata.RegistryAuth, error) {
	args := m.Called(org, cfg, image)
	if err := args.Error(0); err != nil || cfg.CloudConnectionID == "" {
		return nil, err
	}
	return &metadata.RegistryAuth{Username: "acme", Password: "acme-token"}, nil
}

func TestImageSourceAuthorization(t *testing.T) {
	for name, tc := range map[string]struct {
		registry configs.OCIRegistryType
		image    string
		method   string
	}{
		"ECR": {configs.OCIRegistryTypeECR, "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme", "ECRAuthorization"},
		"GAR": {configs.OCIRegistryTypeGAR, "us-central1-docker.pkg.dev/acme/repo/image", "GARAuthorization"},
		"ACR": {configs.OCIRegistryTypeACR, "acme.azurecr.io/repo", "ACRAuthorization"},
	} {
		for mode, mc := range map[string]struct{ legacy, fail, foreign, disabled bool }{
			"pinned": {}, "legacy": {legacy: true}, "authorization failure": {fail: true}, "foreign owner": {foreign: true},
			"legacy disabled": {legacy: true, disabled: true},
		} {
			for jobName, metadataJob := range map[string]bool{"build": false, "metadata": true} {
				t.Run(name+"/"+mode+"/"+jobName, func(t *testing.T) {
					cfg := &app.ExternalImageComponentConfig{ImageURL: tc.image}
					var authConfig any
					switch tc.registry {
					case configs.OCIRegistryTypeECR:
						cfg.AWSECRImageConfig = &app.AWSECRImageConfig{ID: "saved", AWSRegion: "us-east-1", CloudConnectionID: "cc_saved"}
						if mc.legacy {
							cfg.AWSECRImageConfig.CloudConnectionID = ""
							cfg.AWSECRImageConfig.IAMRoleARN = "arn:aws:iam::123456789012:role/acme"
						}
						authConfig = cfg.AWSECRImageConfig
					case configs.OCIRegistryTypeGAR:
						cfg.GCPGARImageConfig = &app.GCPGARImageConfig{ID: "saved", GCPRegion: "us-central1", CloudConnectionID: "cc_saved"}
						if mc.legacy {
							cfg.GCPGARImageConfig.CloudConnectionID = ""
							cfg.GCPGARImageConfig.ServiceAccountEmail = "acme@example.com"
						}
						authConfig = cfg.GCPGARImageConfig
					case configs.OCIRegistryTypeACR:
						cfg.AzureACRImageConfig = &app.AzureACRImageConfig{ID: "saved", RegistryURL: "acme.azurecr.io", CloudConnectionID: "cc_saved"}
						if mc.legacy {
							cfg.AzureACRImageConfig.CloudConnectionID = ""
						}
						authConfig = cfg.AzureACRImageConfig
					}
					db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
					require.NoError(t, err)
					require.NoError(t, db.Callback().Query().After("gorm:query").Register("image-source", func(tx *gorm.DB) {
						switch dest := tx.Statement.Dest.(type) {
						case *app.RunnerJob:
							*dest = app.RunnerJob{ID: "job", OrgID: "org_acme", OwnerID: "build", OwnerType: "component_builds"}
						case *app.ComponentBuild:
							require.Contains(t, tx.Statement.Vars, "org_acme")
							require.Contains(t, tx.Statement.Vars, "build")
							if mc.foreign {
								tx.AddError(gorm.ErrRecordNotFound)
								return
							}
							*dest = app.ComponentBuild{ID: "build", OrgID: "org_acme", ComponentConfigConnection: app.ComponentConfigConnection{ExternalImageComponentConfig: cfg}}
						}
						tx.RowsAffected = 1
					}))
					authorizer := &mockImageAuthorizer{}
					if !mc.foreign {
						var authError error
						if mc.fail {
							authError = errors.New("connection authorization denied")
						}
						if mc.disabled {
							authError = cloudconnections.ErrLegacyImageAuthDisabled
						}
						authorizer.On(tc.method, "org_acme", authConfig, tc.image).Return(authError).Once()
					}
					a := &Activities{db: db, cloudConnectionsHelpers: authorizer}
					source := &configs.OCIRegistryRepository{RegistryType: tc.registry, Repository: tc.image}
					plan := &plantypes.CompositePlan{}
					if metadataJob {
						plan.FetchImageMetadataPlan = &plantypes.FetchImageMetadataPlan{Registry: source}
					} else {
						plan.BuildPlan = &plantypes.BuildPlan{ContainerImagePullPlan: &plantypes.ContainerImagePullPlan{RepoCfg: source}}
					}
					switch tc.registry {
					case configs.OCIRegistryTypeECR, configs.OCIRegistryTypeGAR, configs.OCIRegistryTypeACR:
						err = a.authorizeImageSource(context.Background(), "job", plan)
					}
					switch {
					case mc.disabled:
						require.ErrorIs(t, err, cloudconnections.ErrLegacyImageAuthDisabled)
						require.Equal(t, tc.registry, source.RegistryType)
						require.Nil(t, source.OCIAuth)
					case mc.fail || mc.foreign:
						require.Error(t, err)
						require.Equal(t, tc.registry, source.RegistryType)
						require.Nil(t, source.OCIAuth)
					case mc.legacy && tc.registry == configs.OCIRegistryTypeACR:
						require.NoError(t, err)
						require.Equal(t, configs.OCIRegistryTypeACR, source.RegistryType)
					default:
						require.NoError(t, err)
						require.Equal(t, configs.OCIRegistryTypePrivateOCI, source.RegistryType)
						require.Equal(t, "acme-token", source.OCIAuth.Password)
						if tc.registry == configs.OCIRegistryTypeECR {
							require.Equal(t, "123456789012.dkr.ecr.us-east-1.amazonaws.com", source.LoginServer)
							require.Equal(t, "acme", source.Repository)
						}
					}
					authorizer.AssertExpectations(t)
				})
			}
		}
	}
}
