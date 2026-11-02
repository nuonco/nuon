package helpers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type mockImageAuthActivities struct{ mock.Mock }

func (m *mockImageAuthActivities) GetECRAccessToken(ctx context.Context, req *sharedactivities.GetECRAccessTokenRequest) (*sharedactivities.ECRAccessToken, error) {
	args := m.Called(req)
	return &sharedactivities.ECRAccessToken{Username: "AWS", Password: "acme-token", ServerAddress: "123456789012.dkr.ecr.us-east-1.amazonaws.com"}, args.Error(0)
}

func (m *mockImageAuthActivities) GetGARAccessToken(ctx context.Context, req *sharedactivities.GetGARAccessTokenRequest) (*sharedactivities.GARAccessToken, error) {
	args := m.Called(req)
	return &sharedactivities.GARAccessToken{Username: "oauth2accesstoken", Password: "acme-token"}, args.Error(0)
}

func TestLegacyImageAuthorization(t *testing.T) {
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
		for result, tokenError := range map[string]error{"success": nil, "denied": errors.New("legacy authorization denied")} {
			t.Run(name+"/"+result, func(t *testing.T) {
				acts := &mockImageAuthActivities{}
				acts.On("GetECRAccessToken", mock.MatchedBy(func(req *sharedactivities.GetECRAccessTokenRequest) bool {
					require.Equal(t, "us-east-1", req.Credentials.Region)
					role := req.Credentials.AssumeRole
					require.Equal(t, "arn:aws:iam::123456789012:role/acme", role.RoleARN)
					require.Equal(t, tc.gcpOIDC, role.UseGCPOIDC)
					if tc.twoStep {
						require.Equal(t, tc.managementRole, role.TwoStepConfig.IAMRoleARN)
					} else {
						require.Nil(t, role.TwoStepConfig)
					}
					return true
				})).Return(tokenError).Once()
				h := &Helpers{cfg: &internal.Config{CloudProvider: tc.cloud, ManagementIAMRoleARN: tc.managementRole, LegacyImageAuthEnabled: true}, imageAuth: acts}
				auth, err := h.ECRAuthorization(context.Background(), "org_acme", &app.AWSECRImageConfig{ID: "saved", AWSRegion: "us-east-1", IAMRoleARN: "arn:aws:iam::123456789012:role/acme"}, "image")
				if tokenError != nil {
					require.ErrorIs(t, err, tokenError)
					require.Nil(t, auth)
				} else {
					require.NoError(t, err)
					require.Equal(t, "acme-token", auth.RegistryToken)
				}
				acts.AssertExpectations(t)
			})
		}
	}
	for name, tc := range map[string]struct{ account, provider string }{
		"ambient":                       {},
		"impersonation":                 {account: "acme@example.com"},
		"federation":                    {provider: "provider"},
		"federation with impersonation": {account: "acme@example.com", provider: "provider"},
	} {
		t.Run(name, func(t *testing.T) {
			acts := &mockImageAuthActivities{}
			acts.On("GetGARAccessToken", &sharedactivities.GetGARAccessTokenRequest{ServiceAccountEmail: tc.account, WorkloadIdentityProvider: tc.provider}).Return(nil).Once()
			h := &Helpers{cfg: &internal.Config{LegacyImageAuthEnabled: true}, imageAuth: acts}
			token, err := h.GARAuthorization(context.Background(), "org_acme", &app.GCPGARImageConfig{ID: "saved", ServiceAccountEmail: tc.account, WorkloadIdentityProvider: tc.provider}, "image")
			require.NoError(t, err)
			require.Equal(t, "acme-token", token.AccessToken)
			acts.AssertExpectations(t)
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
			acts := &mockImageAuthActivities{}
			h := &Helpers{cfg: &internal.Config{}, imageAuth: acts}
			require.ErrorIs(t, authorize(h), cloudconnections.ErrLegacyImageAuthDisabled)
			acts.AssertNotCalled(t, "GetECRAccessToken", mock.Anything)
			acts.AssertNotCalled(t, "GetGARAccessToken", mock.Anything)
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
				acts := &mockImageAuthActivities{}
				h := &Helpers{cfg: &internal.Config{LegacyImageAuthEnabled: true}, db: db, imageAuth: acts}
				switch platform {
				case "ecr":
					_, err = h.ECRAuthorization(context.Background(), "org_acme", &app.AWSECRImageConfig{ID: "saved", CloudConnectionID: "cc_saved", AWSRegion: "us-east-1"}, "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme")
				case "gar":
					_, err = h.GARAuthorization(context.Background(), "org_acme", &app.GCPGARImageConfig{ID: "saved", CloudConnectionID: "cc_saved", GCPRegion: "us-central1", GCPProjectID: "acme"}, "us-central1-docker.pkg.dev/acme/repo/image")
				case "acr":
					_, err = h.ACRAuthorization(context.Background(), "org_acme", &app.AzureACRImageConfig{ID: "saved", CloudConnectionID: "cc_saved", RegistryURL: "acme.azurecr.io"}, "acme.azurecr.io/repo")
				}
				require.ErrorContains(t, err, tc.want)
				acts.AssertNotCalled(t, "GetECRAccessToken", mock.Anything)
				acts.AssertNotCalled(t, "GetGARAccessToken", mock.Anything)
			})
		}
	}
}
