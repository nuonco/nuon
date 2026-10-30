package cloudconnections

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google/externalaccount"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func GARProvider(connection *app.CloudConnection) string {
	return fmt.Sprintf("projects/%s/locations/global/workloadIdentityPools/nuon-%s/providers/nuon", connection.GCPProjectNumber, connection.ID)
}

func GARConnection(ctx context.Context, db *gorm.DB, orgID string, cfg *app.GCPGARImageConfig, imageURL string) (*app.CloudConnection, error) {
	if cfg == nil {
		return nil, fmt.Errorf("GAR image configuration is required")
	}
	if cfg.ServiceAccountEmail != "" || cfg.WorkloadIdentityProvider != "" {
		return nil, fmt.Errorf("gcp_gar service_account_email and workload_identity_provider are no longer supported; configure access on the Cloud Connection instead")
	}
	if cfg.ID != "" && cfg.CloudConnectionID == "" {
		return nil, fmt.Errorf("GAR image configuration has no pinned cloud connection; sync the app configuration before building")
	}
	parts := strings.Split(imageURL, "/")
	if len(parts) < 4 || parts[0] != cfg.GCPRegion+"-docker.pkg.dev" || parts[1] != cfg.GCPProjectID {
		return nil, fmt.Errorf("GAR image URL must match the configured project and location")
	}
	repository := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", parts[1], cfg.GCPRegion, parts[2])
	connection, err := imageConnection(ctx, db, app.CloudConnection{
		ID: cfg.CloudConnectionID, OrgID: orgID, Platform: app.CloudPlatformGCP,
		Preset: app.CloudConnectionPresetGAR, TargetID: cfg.GCPProjectID, DefaultRegion: cfg.GCPRegion,
	}, repository)
	if err != nil {
		return nil, err
	}
	cfg.CloudConnectionID = connection.ID
	return connection, nil
}

type garSubjectTokenSupplier struct {
	issuer   *oidcissuer.Issuer
	subject  string
	audience string
}

func (s garSubjectTokenSupplier) SubjectToken(ctx context.Context, _ externalaccount.SupplierOptions) (string, error) {
	return s.issuer.Mint(ctx, s.subject, s.audience, 10*time.Minute)
}

func GARTokenSource(ctx context.Context, issuer *oidcissuer.Issuer, connection *app.CloudConnection, subject string) (oauth2.TokenSource, error) {
	if issuer == nil {
		return nil, fmt.Errorf("cloud connection OIDC issuer is unavailable")
	}
	audience := "//iam.googleapis.com/" + GARProvider(connection)
	return externalaccount.NewTokenSource(ctx, externalaccount.Config{
		Audience:                       audience,
		SubjectTokenType:               "urn:ietf:params:oauth:token-type:jwt",
		Scopes:                         []string{"https://www.googleapis.com/auth/cloud-platform"},
		ServiceAccountImpersonationURL: "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/" + connection.Principal + ":generateAccessToken",
		SubjectTokenSupplier:           garSubjectTokenSupplier{issuer: issuer, subject: subject, audience: audience},
	})
}
