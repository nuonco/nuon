package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workloadjwt"
)

const (
	azureJWKSCacheDuration  = 5 * time.Minute
	azureManagementAudience = "https://management.azure.com/"
)

type RunnerAuthAzureRequest struct {
	Token    string `json:"token" validate:"required"`
	RunnerID string `json:"runner_id,omitempty"`
}

type RunnerAuthAzureResponse struct {
	Authenticated bool   `json:"authenticated"`
	TenantID      string `json:"tenant_id,omitempty"`
	PrincipalID   string `json:"principal_id,omitempty"`
	RunnerID      string `json:"runner_id,omitempty"`
	Token         string `json:"token,omitempty"`
}

// azureClaims holds the custom claims we extract from Azure managed identity JWTs.
type azureClaims struct {
	TenantID string `json:"tid"`
	ObjectID string `json:"oid"`
	Subject  string `json:"sub"`
	XMSMirID string `json:"xms_mirid"`
}

func (a *azureClaims) Validate(_ context.Context) error {
	if a.TenantID == "" {
		return errors.New("missing tid claim")
	}
	if a.ObjectID == "" {
		return errors.New("missing oid claim")
	}
	return nil
}

// azureJWKSProvider caches JWKS providers per tenant to avoid repeated discovery.
type azureJWKSProvider struct {
	mu        sync.RWMutex
	providers map[string]*jwks.CachingProvider
}

func newAzureJWKSProvider() *azureJWKSProvider {
	return &azureJWKSProvider{
		providers: make(map[string]*jwks.CachingProvider),
	}
}

func (p *azureJWKSProvider) getProvider(tenantID string) (*jwks.CachingProvider, error) {
	p.mu.RLock()
	provider, ok := p.providers[tenantID]
	p.mu.RUnlock()
	if ok {
		return provider, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock.
	if provider, ok := p.providers[tenantID]; ok {
		return provider, nil
	}

	issuer, err := workloadjwt.AzureIssuer(tenantID)
	if err != nil {
		return nil, err
	}
	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer url: %w", err)
	}

	provider = jwks.NewCachingProvider(issuerURL, azureJWKSCacheDuration)
	p.providers[tenantID] = provider
	return provider, nil
}

var azureJWKS = newAzureJWKSProvider()

// parseUnverifiedTenantID extracts the tid claim from a compact JWT without signature
// verification, only to select which tenant's JWKS to use. The value must be a GUID so
// it cannot reshape the discovery URL, and is re-checked against verified claims after.
func parseUnverifiedTenantID(tokenStr string) (string, error) {
	claims, err := workloadjwt.UnverifiedClaims(tokenStr)
	if err != nil {
		return "", err
	}

	tenantID, ok := workloadjwt.StringClaim(claims, "tid")
	if !ok {
		return "", errors.New("JWT missing tid claim")
	}

	if _, err := workloadjwt.AzureIssuer(tenantID); err != nil {
		return "", err
	}

	return tenantID, nil
}

// verifyAzureJWT verifies an Azure managed identity JWT and returns the parsed claims.
func (s *service) verifyAzureJWT(ctx context.Context, tokenStr string) (*azureClaims, error) {
	tokenStr = strings.TrimSpace(tokenStr)
	if err := workloadjwt.RequireCompactJWT(tokenStr); err != nil {
		return nil, err
	}

	tenantID, err := parseUnverifiedTenantID(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tenant ID: %w", err)
	}

	provider, err := azureJWKS.getProvider(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant JWKS: %w", err)
	}

	issuerURL, err := workloadjwt.AzureIssuer(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to build issuer: %w", err)
	}

	customClaimsFn := func() validator.CustomClaims {
		return &azureClaims{}
	}

	tokenValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL,
		[]string{azureManagementAudience},
		validator.WithAllowedClockSkew(time.Minute),
		validator.WithCustomClaims(customClaimsFn),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT validator: %w", err)
	}

	validToken, err := tokenValidator.ValidateToken(ctx, tokenStr)
	if err != nil {
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	validatedClaims, ok := validToken.(*validator.ValidatedClaims)
	if !ok {
		return nil, errors.New("unexpected token validation response")
	}

	azClaims, ok := validatedClaims.CustomClaims.(*azureClaims)
	if !ok {
		return nil, errors.New("unexpected custom claims type")
	}

	// Key selection and verified claims must name the same tenant. Without this,
	// a confused parser could fetch keys for one tenant while accepting tid/oid
	// for another (NUO-1188).
	if err := bindAzureKeyTenant(tenantID, azClaims); err != nil {
		return nil, err
	}

	return azClaims, nil
}

// bindAzureKeyTenant requires the verified tid to identify the tenant whose JWKS
// verified the signature. iss is already constrained to that tenant by the validator.
func bindAzureKeyTenant(keyTenantID string, claims *azureClaims) error {
	if claims == nil {
		return errors.New("missing azure claims")
	}
	if !strings.EqualFold(claims.TenantID, keyTenantID) {
		return fmt.Errorf("verified tid %q does not match key tenant %q", claims.TenantID, keyTenantID)
	}
	return nil
}

// extractRunnerIDFromXMSMirID parses the Azure managed identity resource ID to extract the runner ID.
// The format is: /subscriptions/{sub}/resourcegroups/{rg}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/{name}
// The identity name is expected to contain or be the runner ID.
// For system-assigned identities (e.g. VMSS), xms_mirid points to the compute resource
// rather than a ManagedIdentity resource — those are not supported here.
func extractRunnerIDFromXMSMirID(xmsMirID string) (runnerID string, subscriptionID string, err error) {
	if xmsMirID == "" {
		return "", "", errors.New("empty xms_mirid claim")
	}

	lower := strings.ToLower(xmsMirID)
	parts := strings.Split(strings.TrimPrefix(lower, "/"), "/")

	// Expected: subscriptions/{sub}/resourcegroups/{rg}/providers/microsoft.managedidentity/userassignedidentities/{name}
	if len(parts) < 8 {
		return "", "", fmt.Errorf("unexpected xms_mirid format: %s", xmsMirID)
	}

	if parts[0] != "subscriptions" {
		return "", "", fmt.Errorf("unexpected xms_mirid format, expected 'subscriptions' prefix: %s", xmsMirID)
	}

	subscriptionID = parts[1]

	// Only extract runner ID from user-assigned managed identity resources.
	// System-assigned identities (e.g. VMSS, VM) have a different resource type
	// and the last segment is not a runner ID.
	if !strings.Contains(lower, "microsoft.managedidentity/userassignedidentities") {
		return "", subscriptionID, fmt.Errorf("xms_mirid is not a user-assigned managed identity: %s", xmsMirID)
	}

	// The identity name is the last segment.
	originalParts := strings.Split(strings.TrimPrefix(xmsMirID, "/"), "/")
	identityName := originalParts[len(originalParts)-1]

	return identityName, subscriptionID, nil
}

// extractRunnerIDFromClaims extracts the runner ID from Azure JWT claims.
// It first tries xms_mirid, then falls back to the request-provided runner_id.
func extractRunnerIDFromClaims(claims *azureClaims, requestRunnerID string) (runnerID string, subscriptionID string, err error) {
	if claims.XMSMirID != "" {
		runnerID, subscriptionID, err = extractRunnerIDFromXMSMirID(claims.XMSMirID)
		if err == nil && runnerID != "" {
			return runnerID, subscriptionID, nil
		}
	}

	// Fallback to request-provided runner ID.
	if requestRunnerID != "" {
		return requestRunnerID, subscriptionID, nil
	}

	return "", "", errors.New("unable to determine runner ID: xms_mirid claim missing or invalid and no runner_id provided")
}

// @ID						RunnerAuthAzure
// @Summary				Authenticate a runner using Azure managed identity JWT
// @Description			Validates runner identity by verifying an Azure IMDS JWT token
// @Param					req	body	RunnerAuthAzureRequest	true	"Azure JWT token"
// @Tags					runners/auth
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	RunnerAuthAzureResponse
// @Router					/v1/runner-auth/azure [POST]
func (s *service) RunnerAuthAzure(ctx *gin.Context) {
	start := time.Now()
	metricTags := map[string]string{
		"cloud_provider": "azure",
		"auth_method":    "jwt",
		"status":         "error",
	}
	defer func() {
		if s.mw != nil {
			s.mw.Timing("runner.auth.latency", time.Since(start), metrics.ToTags(metricTags))
		}
	}()

	var req RunnerAuthAzureRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.l.Warn("runner auth azure: failed to parse request", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request format")))
		ctx.Abort()
		return
	}

	if err := s.v.Struct(req); err != nil {
		s.l.Warn("runner auth azure: request validation failed", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request: missing required fields")))
		ctx.Abort()
		return
	}

	reqCtx := ctx.Request.Context()

	azClaims, err := s.verifyAzureJWT(reqCtx, req.Token)
	if err != nil {
		s.l.Warn("runner auth azure: JWT verification failed", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "failed to verify Azure identity token",
		})
		ctx.Abort()
		return
	}

	runnerID, subscriptionID, err := extractRunnerIDFromClaims(azClaims, req.RunnerID)
	if err != nil {
		s.l.Warn("runner auth azure: failed to extract runner ID",
			zap.String("xms_mirid", azClaims.XMSMirID),
			zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "unable to determine runner identity",
		})
		ctx.Abort()
		return
	}

	runner, err := s.getRunnerWithGroup(reqCtx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.l.Warn("runner auth azure: runner not found", zap.String("runner_id", runnerID))
		} else {
			s.l.Error("runner auth azure: failed to get runner", zap.String("runner_id", runnerID), zap.Error(err))
		}
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "runner not recognized",
		})
		ctx.Abort()
		return
	}
	metricTags["runner_id"] = runner.ID

	install, err := s.getInstallByRunnerGroup(reqCtx, &runner.RunnerGroup)
	if err != nil {
		s.l.Warn("runner auth azure: failed to get install for runner",
			zap.String("runner_id", runnerID),
			zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "runner not associated with an install",
		})
		ctx.Abort()
		return
	}
	metricTags["install_id"] = install.ID
	metricTags["install_name"] = install.Name
	metricTags["org_id"] = install.OrgID

	if err := s.validateRunnerAzureIdentity(reqCtx, install, azClaims, subscriptionID); err != nil {
		s.l.Warn("runner auth azure: identity validation failed",
			zap.String("runner_id", runnerID),
			zap.String("install_id", install.ID),
			zap.String("install_name", install.Name),
			zap.String("org_id", install.OrgID),
			zap.String("tenant_id", azClaims.TenantID),
			zap.String("principal_id", azClaims.ObjectID),
			zap.Error(err))
		ctx.Error(stderr.ErrAuthorization{
			Err:         errors.New("authorization failed"),
			Description: "runner identity does not match expected configuration",
		})
		ctx.Abort()
		return
	}

	token, err := s.createRunnerToken(reqCtx, runner.ID)
	if err != nil {
		s.l.Error("runner auth azure: failed to create token", zap.String("runner_id", runnerID), zap.Error(err))
		ctx.Error(stderr.ErrSystem{
			Err:         errors.New("internal error"),
			Description: "failed to issue authentication token",
		})
		ctx.Abort()
		return
	}

	metricTags["status"] = "ok"
	s.l.Info("runner auth azure: authentication successful",
		zap.String("install_id", install.ID),
		zap.String("install_name", install.Name),
		zap.String("org_id", install.OrgID),
		zap.String("runner_id", runner.ID),
		zap.String("tenant_id", azClaims.TenantID),
		zap.String("principal_id", azClaims.ObjectID))

	ctx.JSON(http.StatusOK, RunnerAuthAzureResponse{
		Authenticated: true,
		TenantID:      azClaims.TenantID,
		PrincipalID:   azClaims.ObjectID,
		RunnerID:      runner.ID,
		Token:         token,
	})
}

func (s *service) validateRunnerAzureIdentity(ctx context.Context, install *app.Install, claims *azureClaims, subscriptionID string) error {
	installStack, err := s.getInstallStackWithOutputs(ctx, install.ID)
	if err != nil {
		return fmt.Errorf("failed to get install stack for install %s: %w", install.ID, err)
	}

	if installStack.InstallStackOutputs.AzureStackOutputs == nil {
		return fmt.Errorf("install %s does not have Azure stack outputs configured", install.ID)
	}

	return matchAzureRunnerClaims(install.ID, installStack.InstallStackOutputs.AzureStackOutputs, claims, subscriptionID)
}

func matchAzureRunnerClaims(installID string, outputs *app.AzureStackOutputs, claims *azureClaims, subscriptionID string) error {
	expectedTenantID := outputs.SubscriptionTenantID
	if expectedTenantID == "" {
		return fmt.Errorf("install %s does not have a subscription tenant ID in stack outputs", installID)
	}

	if claims.TenantID != expectedTenantID {
		return fmt.Errorf("tenant ID mismatch: got %s, expected %s", claims.TenantID, expectedTenantID)
	}

	if subscriptionID != "" {
		expectedSubscriptionID := outputs.SubscriptionID
		if expectedSubscriptionID != "" && !strings.EqualFold(subscriptionID, expectedSubscriptionID) {
			return fmt.Errorf("subscription ID mismatch: got %s, expected %s", subscriptionID, expectedSubscriptionID)
		}
	}

	expectedPrincipalID := outputs.RunnerIdentityPrincipalID
	if expectedPrincipalID == "" {
		return fmt.Errorf("install %s does not have a runner identity principal id in stack outputs", installID)
	}
	if !strings.EqualFold(claims.ObjectID, expectedPrincipalID) {
		return fmt.Errorf("principal id mismatch: got %s, expected %s", claims.ObjectID, expectedPrincipalID)
	}

	return nil
}
