package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
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
	parts, err := oidc.CompactJWSParts(tokenStr)
	if err != nil {
		return "", err
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

func (s *service) verifyOIDCToken(ctx context.Context, tokenStr, issuer, audience string) (oidc.Claims, error) {
	return s.verifier.Verify(ctx, tokenStr, issuer, audience)
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

	tokenStr := strings.TrimSpace(req.Token)
	issuer, err := parseUnverifiedIssuer(tokenStr)
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
	var claims oidc.Claims
	haveClaims := false

	for i := range policies {
		policy := &policies[i]

		verified, seen := verifiedAudiences[policy.Audience]
		if !seen {
			tokenClaims, err := s.verifyOIDCToken(reqCtx, tokenStr, policy.IssuerURL, policy.Audience)
			verified = err == nil
			verifiedAudiences[policy.Audience] = verified
			if err != nil {
				s.l.Warn("oidc exchange: token verification failed",
					zap.String("org_id", req.OrgID),
					zap.String("issuer", issuer),
					zap.String("policy_id", policy.ID),
					zap.Error(err))
			} else if !haveClaims {
				claims = tokenClaims
				haveClaims = true
			}
		}
		if !verified {
			continue
		}

		if oidc.ParseConditions(policy.ClaimConditions).Match(claims) {
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
		if haveClaims {
			fields = append(fields, zap.String("sub", claims.Sub), zap.Strings("claim_keys", claims.Names()))
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

	metricTags["status"] = "ok"
	s.l.Info("oidc exchange: token issued",
		zap.String("org_id", matched.OrgID),
		zap.String("policy_id", matched.ID),
		zap.String("issuer", matched.IssuerURL),
		zap.String("sub", claims.Sub),
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
