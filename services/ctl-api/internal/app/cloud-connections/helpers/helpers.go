package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/aws/smithy-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google/externalaccount"
	"gorm.io/gorm"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type Helpers struct {
	db     *gorm.DB
	cfg    *internal.Config
	issuer *oidcissuer.Issuer
}

func New(db *gorm.DB, cfg *internal.Config) (*Helpers, error) {
	if cfg.TelemetryJWKS == "" {
		return &Helpers{db: db, cfg: cfg}, nil
	}
	privateKey, keyID, _, err := oidcissuer.ParseJWKS(cfg.TelemetryJWKS)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	issuer, err := oidcissuer.New(cfg.PublicAPIURL, privateKey, keyID)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	return &Helpers{db: db, cfg: cfg, issuer: issuer}, nil
}

func (h *Helpers) Credentials(ctx context.Context, connection *app.CloudConnection, sessionName string) (*credentials.Config, error) {
	if h.issuer == nil {
		return nil, fmt.Errorf("cloud connection OIDC issuer is unavailable")
	}
	token, err := h.issuer.Mint(ctx, fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID), "sts.amazonaws.com", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	return &credentials.Config{Region: connection.DefaultRegion, AssumeRole: &credentials.AssumeRoleConfig{RoleARN: connection.Principal, SessionName: sessionName, SessionDurationSeconds: 900, WebIdentityToken: token}}, nil
}

func (h *Helpers) AzureCredential(connection *app.CloudConnection) (azcore.TokenCredential, error) {
	return AzureCredential(h.issuer, connection)
}

func AzureCredential(issuer *oidcissuer.Issuer, connection *app.CloudConnection) (azcore.TokenCredential, error) {
	if issuer == nil {
		return nil, fmt.Errorf("cloud connection OIDC issuer is unavailable")
	}
	return azidentity.NewClientAssertionCredential(connection.TenantID, connection.Principal, func(ctx context.Context) (string, error) {
		return issuer.Mint(ctx, fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID), "api://AzureADTokenExchange", 10*time.Minute)
	}, nil)
}

type gcpSubjectTokenSupplier struct {
	issuer     *oidcissuer.Issuer
	connection *app.CloudConnection
	subject    string
}

func (s *gcpSubjectTokenSupplier) SubjectToken(ctx context.Context, _ externalaccount.SupplierOptions) (string, error) {
	return s.issuer.Mint(ctx, s.subject, gcpSubjectTokenAudience(s.connection.IdentityProvider), 10*time.Minute)
}

func (h *Helpers) GCPAccessToken(ctx context.Context, connection *app.CloudConnection) (*oauth2.Token, error) {
	federated, err := GCPFederatedToken(ctx, h.issuer, connection, fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID))
	if err != nil {
		return nil, err
	}
	return GCPImpersonatedToken(ctx, http.DefaultClient, federated.AccessToken, connection.Principal)
}

func GCPFederatedToken(ctx context.Context, issuer *oidcissuer.Issuer, connection *app.CloudConnection, tokenSubject string) (*oauth2.Token, error) {
	if issuer == nil {
		return nil, fmt.Errorf("cloud connection OIDC issuer is unavailable")
	}
	config := externalaccount.Config{
		Audience:             gcpExternalAccountAudience(connection.IdentityProvider),
		SubjectTokenType:     "urn:ietf:params:oauth:token-type:jwt",
		Scopes:               []string{"https://www.googleapis.com/auth/cloud-platform"},
		SubjectTokenSupplier: &gcpSubjectTokenSupplier{issuer: issuer, connection: connection, subject: tokenSubject},
	}
	tokenSource, err := externalaccount.NewTokenSource(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create GCP federated token source: %w", err)
	}
	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("exchange Nuon OIDC token: %w", err)
	}
	return token, nil
}

func GCPImpersonatedToken(ctx context.Context, client *http.Client, federatedToken, serviceAccountEmail string) (*oauth2.Token, error) {
	body := strings.NewReader(`{"scope":["https://www.googleapis.com/auth/cloud-platform"],"lifetime":"900s"}`)
	endpoint := "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/" + url.PathEscape(serviceAccountEmail) + ":generateAccessToken"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+federatedToken)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, &gcpHTTPError{statusCode: response.StatusCode, message: fmt.Sprintf("generate GCP service account access token returned %s: %s", response.Status, strings.TrimSpace(string(message)))}
	}
	var output struct {
		AccessToken string    `json:"accessToken"`
		ExpireTime  time.Time `json:"expireTime"`
	}
	if err := json.NewDecoder(response.Body).Decode(&output); err != nil {
		return nil, fmt.Errorf("decode GCP service account access token: %w", err)
	}
	if output.AccessToken == "" {
		return nil, fmt.Errorf("GCP service account access token response was empty")
	}
	return &oauth2.Token{AccessToken: output.AccessToken, TokenType: "Bearer", Expiry: output.ExpireTime}, nil
}

type gcpHTTPError struct {
	statusCode int
	message    string
}

func (e *gcpHTTPError) Error() string {
	return e.message
}

func GCPAccessDenied(err error) bool {
	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) && retrieveErr.Response != nil {
		return retrieveErr.Response.StatusCode == http.StatusBadRequest || retrieveErr.Response.StatusCode == http.StatusUnauthorized || retrieveErr.Response.StatusCode == http.StatusForbidden
	}
	var httpErr *gcpHTTPError
	return errors.As(err, &httpErr) && (httpErr.statusCode == http.StatusUnauthorized || httpErr.statusCode == http.StatusForbidden)
}

func gcpExternalAccountAudience(identityProvider string) string {
	provider := strings.TrimPrefix(identityProvider, "https://iam.googleapis.com/")
	provider = strings.TrimPrefix(provider, "//iam.googleapis.com/")
	return "//iam.googleapis.com/" + strings.TrimPrefix(provider, "/")
}

func gcpSubjectTokenAudience(identityProvider string) string {
	provider := strings.TrimPrefix(identityProvider, "https://iam.googleapis.com/")
	provider = strings.TrimPrefix(provider, "//iam.googleapis.com/")
	return "https://iam.googleapis.com/" + strings.TrimPrefix(provider, "/")
}

func (h *Helpers) ECRCredentials(ctx context.Context, connection *app.CloudConnection, sessionName string) (*credentials.Config, error) {
	oidc, err := h.Credentials(ctx, connection, sessionName)
	if err != nil {
		return nil, err
	}
	var legacy *credentials.Config
	mode, err := chooseAuthMode(connection.AuthMode,
		func() error { _, err := credentials.Fetch(ctx, oidc); return err },
		func() error {
			var err error
			legacy, err = h.legacyCredentials(connection, sessionName)
			if err != nil {
				return err
			}
			_, err = credentials.Fetch(ctx, legacy)
			return err
		},
	)
	if err != nil {
		return nil, err
	}
	selected := oidc
	if mode == app.CloudConnectionAuthModeLegacy {
		selected = legacy
	}
	if mode != connection.AuthMode {
		if err := h.db.WithContext(ctx).Model(&app.CloudConnection{}).Where(app.CloudConnection{ID: connection.ID, OrgID: connection.OrgID}).Update("auth_mode", mode).Error; err != nil {
			return nil, fmt.Errorf("cache cloud connection auth mode: %w", err)
		}
		connection.AuthMode = mode
	}
	return selected, nil
}

func (h *Helpers) legacyCredentials(connection *app.CloudConnection, sessionName string) (*credentials.Config, error) {
	assume := &credentials.AssumeRoleConfig{RoleARN: connection.Principal, SessionName: sessionName, SessionDurationSeconds: 900}
	switch {
	case h.cfg.IsAWS() && h.cfg.ManagementIAMRoleARN != "":
		assume.TwoStepConfig = &assumerole.TwoStepConfig{IAMRoleARN: h.cfg.ManagementIAMRoleARN}
	case h.cfg.IsGCP():
		assume.UseGCPOIDC = true
	default:
		return nil, fmt.Errorf("legacy AWS cloud connection authentication is unavailable on this control plane")
	}
	return &credentials.Config{Region: connection.DefaultRegion, AssumeRole: assume}, nil
}

func chooseAuthMode(current app.CloudConnectionAuthMode, oidc, legacy func() error) (app.CloudConnectionAuthMode, error) {
	if current == app.CloudConnectionAuthModeOIDC {
		if err := oidc(); err != nil {
			return "", err
		}
		return app.CloudConnectionAuthModeOIDC, nil
	}
	if err := oidc(); err == nil {
		return app.CloudConnectionAuthModeOIDC, nil
	} else if !accessDenied(err) {
		return "", err
	}
	if err := legacy(); err != nil {
		return "", err
	}
	return app.CloudConnectionAuthModeLegacy, nil
}

func accessDenied(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "AccessDenied" || apiErr.ErrorCode() == "AccessDeniedException")
}
