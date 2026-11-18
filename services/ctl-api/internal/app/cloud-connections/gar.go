package cloudconnections

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google/externalaccount"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func GARProvider(connection *app.CloudConnection) string {
	return fmt.Sprintf("projects/%s/locations/global/workloadIdentityPools/nuon-%s/providers/nuon", connection.GCPProjectNumber, connection.ID)
}

func garHasLegacyAuth(cfg *app.GCPGARImageConfig) bool {
	return cfg.ServiceAccountEmail != "" || cfg.WorkloadIdentityProvider != ""
}

func GARConnection(ctx context.Context, db *gorm.DB, orgID string, cfg *app.GCPGARImageConfig, imageURL string, legacy bool) (*app.CloudConnection, error) {
	if cfg == nil {
		return nil, fmt.Errorf("GAR image configuration is required")
	}
	if garHasLegacyAuth(cfg) {
		if cfg.CloudConnectionID != "" {
			return nil, stderr.NewInvalidRequest(errors.New("GAR cloud connection access cannot be combined with service_account_email or workload_identity_provider"))
		}
		// The legacy flag still covers ECR and ACR. A component-chosen GAR
		// identity is the shared control-plane credential impersonating whatever
		// service account the config names, so it stays refused.
		return nil, ErrLegacyImageAuthDisabled
	}
	if cfg.ID != "" && cfg.CloudConnectionID == "" {
		return legacyImageAuth(legacy, ErrLegacyImageAuthDisabled)
	}
	parts := strings.Split(imageURL, "/")
	if len(parts) < 4 || parts[0] != cfg.GCPRegion+"-docker.pkg.dev" || parts[1] != cfg.GCPProjectID {
		return legacyImageAuth(legacy, stderr.NewInvalidRequest(errors.New("GAR image URL must match the configured project and location")))
	}
	repository := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", parts[1], cfg.GCPRegion, parts[2])
	connection, err := imageConnection(ctx, db, app.CloudConnection{
		ID: cfg.CloudConnectionID, OrgID: orgID, Platform: app.CloudPlatformGCP,
		Preset: app.CloudConnectionPresetGAR, TargetID: cfg.GCPProjectID, DefaultRegion: cfg.GCPRegion,
	}, repository)
	if errors.Is(err, errNoImageRepositoryMapping) {
		return legacyImageAuth(legacy, err)
	}
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
