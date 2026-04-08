package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
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

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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

func (p *azureJWKSProvider) getProvider(tenantID string) *jwks.CachingProvider {
	p.mu.RLock()
	provider, ok := p.providers[tenantID]
	p.mu.RUnlock()
	if ok {
		return provider
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock.
	if provider, ok := p.providers[tenantID]; ok {
		return provider
	}

	issuerURL, _ := url.Parse(fmt.Sprintf("https://sts.windows.net/%s/", tenantID))
	provider = jwks.NewCachingProvider(issuerURL, azureJWKSCacheDuration)
	p.providers[tenantID] = provider
	return provider
}

var azureJWKS = newAzureJWKSProvider()

// parseUnverifiedTenantID extracts the tid claim from a JWT without signature verification.
// This is needed to determine which tenant's JWKS endpoint to use for verification.
func parseUnverifiedTenantID(tokenStr string) (string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid JWT format")
	}

	// Decode the payload (second part).
	payload, err := decodeJWTSegment(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims struct {
		TenantID string `json:"tid"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse JWT claims: %w", err)
	}
	if claims.TenantID == "" {
		return "", errors.New("JWT missing tid claim")
	}

	return claims.TenantID, nil
}

// decodeJWTSegment decodes a base64url-encoded JWT segment.
func decodeJWTSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}

// verifyAzureJWT verifies an Azure managed identity JWT and returns the parsed claims.
func (s *service) verifyAzureJWT(ctx context.Context, tokenStr string) (*azureClaims, error) {
	tenantID, err := parseUnverifiedTenantID(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tenant ID: %w", err)
	}

	provider := azureJWKS.getProvider(tenantID)
	issuerURL := fmt.Sprintf("https://sts.windows.net/%s/", tenantID)

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

	return azClaims, nil
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

	if err := s.validateRunnerAzureIdentity(reqCtx, runner, azClaims, subscriptionID); err != nil {
		s.l.Warn("runner auth azure: identity validation failed",
			zap.String("runner_id", runnerID),
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

	s.l.Info("runner auth azure: authentication successful",
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

func (s *service) validateRunnerAzureIdentity(ctx context.Context, runner *app.Runner, claims *azureClaims, subscriptionID string) error {
	install, err := s.getInstallByRunnerGroup(ctx, &runner.RunnerGroup)
	if err != nil {
		return fmt.Errorf("failed to get install for runner: %w", err)
	}

	installStack, err := s.getInstallStackWithOutputs(ctx, install.ID)
	if err != nil {
		return fmt.Errorf("failed to get install stack for install %s: %w", install.ID, err)
	}

	if installStack.InstallStackOutputs.AzureStackOutputs == nil {
		return fmt.Errorf("install %s does not have Azure stack outputs configured", install.ID)
	}

	azureOutputs := installStack.InstallStackOutputs.AzureStackOutputs

	expectedTenantID := azureOutputs.SubscriptionTenantID
	if expectedTenantID == "" {
		return fmt.Errorf("install %s does not have a subscription tenant ID in stack outputs", install.ID)
	}

	if claims.TenantID != expectedTenantID {
		return fmt.Errorf("tenant ID mismatch: got %s, expected %s", claims.TenantID, expectedTenantID)
	}

	if subscriptionID != "" {
		expectedSubscriptionID := azureOutputs.SubscriptionID
		if expectedSubscriptionID != "" && !strings.EqualFold(subscriptionID, expectedSubscriptionID) {
			return fmt.Errorf("subscription ID mismatch: got %s, expected %s", subscriptionID, expectedSubscriptionID)
		}
	}

	return nil
}
