package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

type ExchangeOIDCTokenRequest struct {
	OrgID string `json:"org_id" validate:"required"`
	Token string `json:"token" validate:"required"`
}

type ExchangeOIDCTokenResponse struct {
	Authenticated bool      `json:"authenticated"`
	Token         string    `json:"token,omitempty"`
	ExpiresAt     time.Time `json:"expires_at,omitzero"`
	OrgID         string    `json:"org_id,omitempty"`
	TrustPolicyID string    `json:"trust_policy_id,omitempty"`
	Role          string    `json:"role,omitempty"`
}

// genericExchangeError is returned for every auth-path failure so responses
// don't reveal whether an org, policy, or issuer is configured.
func genericExchangeError() stderr.ErrAuthentication {
	return stderr.ErrAuthentication{
		Err:         errors.New("authentication failed"),
		Description: "failed to verify OIDC token",
	}
}

// parseUnverifiedIssuer extracts the iss claim from a JWT without signature
// verification. It is only used to select among stored trust policies; the
// issuer used for verification and JWKS fetching always comes from the policy.
func parseUnverifiedIssuer(tokenStr string) (string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims struct {
		Issuer string `json:"iss"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse JWT claims: %w", err)
	}
	if claims.Issuer == "" {
		return "", errors.New("JWT missing iss claim")
	}

	return claims.Issuer, nil
}

// parseTokenClaims decodes the payload of an already-verified JWT into a
// generic claim map for condition matching.
func parseTokenClaims(tokenStr string) (map[string]any, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	claims := map[string]any{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return claims, nil
}

// verifyOIDCToken verifies the token's signature, issuer, audience, and time
// claims against a policy's stored configuration.
func (s *service) verifyOIDCToken(ctx context.Context, tokenStr, issuer, audience string) error {
	provider, err := s.jwks.getProvider(issuer)
	if err != nil {
		return err
	}

	tokenValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuer,
		[]string{audience},
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to create JWT validator: %w", err)
	}

	if _, err := tokenValidator.ValidateToken(ctx, tokenStr); err != nil {
		return fmt.Errorf("JWT validation failed: %w", err)
	}

	return nil
}

// @ID						ExchangeOIDCToken
// @Summary				exchange an OIDC token for a Nuon API token
// @Description			Exchanges an OIDC ID token (e.g. from GitHub Actions) for a short-lived Nuon API token. The token must match an enabled OIDC trust policy in the target org: its signature is verified against the policy issuer's JWKS, and its issuer, audience, and claims must satisfy the policy. No Nuon credentials are required to call this endpoint.
// @Param					req	body	ExchangeOIDCTokenRequest	true	"Input"
// @Tags					oidc_federation
// @Accept					json
// @Produce				json
// @Success				200	{object}	ExchangeOIDCTokenResponse
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/oidc/token [POST]
func (s *service) ExchangeOIDCToken(ctx *gin.Context) {
	start := time.Now()
	metricTags := map[string]string{
		"status": "error",
	}
	defer func() {
		if s.mw != nil {
			s.mw.Timing("oidc.exchange.latency", time.Since(start), metrics.ToTags(metricTags))
		}
	}()

	if !s.cfg.OIDCFederationEnabled {
		metricTags["status"] = "disabled"
		ctx.Error(genericExchangeError())
		return
	}

	var req ExchangeOIDCTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request format")))
		return
	}
	if err := s.v.Struct(req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request: missing required fields")))
		return
	}
	metricTags["org_id"] = req.OrgID

	issuer, err := parseUnverifiedIssuer(req.Token)
	if err != nil {
		s.l.Warn("oidc exchange: failed to parse token issuer", zap.String("org_id", req.OrgID), zap.Error(err))
		ctx.Error(genericExchangeError())
		return
	}

	var policies []app.OIDCTrustPolicy
	res := s.db.WithContext(ctx).
		Where(app.OIDCTrustPolicy{
			OrgID:     req.OrgID,
			IssuerURL: issuer,
			Enabled:   true,
		}).
		Order("created_at ASC").
		Find(&policies)
	if res.Error != nil {
		s.l.Error("oidc exchange: failed to load trust policies", zap.String("org_id", req.OrgID), zap.Error(res.Error))
		ctx.Error(stderr.ErrSystem{
			Err:         errors.New("internal error"),
			Description: "failed to process token exchange",
		})
		return
	}
	if len(policies) == 0 {
		s.l.Warn("oidc exchange: no enabled trust policies for issuer",
			zap.String("org_id", req.OrgID),
			zap.String("issuer", issuer))
		ctx.Error(genericExchangeError())
		return
	}

	reqCtx := ctx.Request.Context()

	verifiedAudiences := map[string]bool{}
	var matched *app.OIDCTrustPolicy
	var claims map[string]any

	for i := range policies {
		policy := &policies[i]

		verified, seen := verifiedAudiences[policy.Audience]
		if !seen {
			err := s.verifyOIDCToken(reqCtx, req.Token, policy.IssuerURL, policy.Audience)
			verified = err == nil
			verifiedAudiences[policy.Audience] = verified
			if err != nil {
				s.l.Warn("oidc exchange: token verification failed",
					zap.String("org_id", req.OrgID),
					zap.String("issuer", issuer),
					zap.String("policy_id", policy.ID),
					zap.Error(err))
			}
		}
		if !verified {
			continue
		}

		if claims == nil {
			claims, err = parseTokenClaims(req.Token)
			if err != nil {
				s.l.Warn("oidc exchange: failed to parse verified token claims", zap.Error(err))
				ctx.Error(genericExchangeError())
				return
			}
		}

		if matchClaims(policy.ClaimConditions, claims) {
			matched = policy
			break
		}
	}

	if matched == nil {
		fields := []zap.Field{
			zap.String("org_id", req.OrgID),
			zap.String("issuer", issuer),
		}
		// When at least one policy verified the token, claims are populated;
		// logging the sub and claim keys lets operators see why no policy's
		// conditions matched without weakening the generic client-facing 401.
		if claims != nil {
			sub, _ := claims["sub"].(string)
			keys := make([]string, 0, len(claims))
			for k := range claims {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			fields = append(fields, zap.String("sub", sub), zap.Strings("claim_keys", keys))
		}
		s.l.Warn("oidc exchange: no trust policy matched", fields...)
		ctx.Error(genericExchangeError())
		return
	}
	metricTags["policy_id"] = matched.ID

	duration := matched.TokenDurationSeconds
	if duration <= 0 {
		duration = app.OIDCTrustPolicyDefaultTokenDuration
	}
	if duration > app.OIDCTrustPolicyMaxTokenDuration {
		duration = app.OIDCTrustPolicyMaxTokenDuration
	}

	now := time.Now()
	token := app.Token{
		CreatedByID: matched.ServiceAccountID,
		Name:        matched.Name,
		Role:        matched.Role,
		OrgID:       matched.OrgID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeFederated,
		ExpiresAt:   now.Add(time.Duration(duration) * time.Second),
		IssuedAt:    now,
		Issuer:      matched.IssuerURL,
		AccountID:   matched.ServiceAccountID,
	}
	if res := s.db.WithContext(ctx).Create(&token); res.Error != nil {
		s.l.Error("oidc exchange: failed to create token",
			zap.String("policy_id", matched.ID),
			zap.Error(res.Error))
		ctx.Error(stderr.ErrSystem{
			Err:         errors.New("internal error"),
			Description: "failed to issue authentication token",
		})
		return
	}

	if err := s.db.WithContext(ctx).
		Model(&app.OIDCTrustPolicy{}).
		Where(app.OIDCTrustPolicy{ID: matched.ID}).
		Update("last_used_at", now).Error; err != nil {
		s.l.Warn("oidc exchange: failed to update last_used_at", zap.String("policy_id", matched.ID), zap.Error(err))
	}

	sub, _ := claims["sub"].(string)
	metricTags["status"] = "ok"
	s.l.Info("oidc exchange: token issued",
		zap.String("org_id", matched.OrgID),
		zap.String("policy_id", matched.ID),
		zap.String("issuer", matched.IssuerURL),
		zap.String("sub", sub),
		zap.String("role", matched.Role))

	ctx.JSON(http.StatusOK, ExchangeOIDCTokenResponse{
		Authenticated: true,
		Token:         token.Token,
		ExpiresAt:     token.ExpiresAt,
		OrgID:         matched.OrgID,
		TrustPolicyID: matched.ID,
		Role:          matched.Role,
	})
}
