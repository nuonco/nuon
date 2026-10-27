package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcverify"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
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

var errAlreadyExchanged = errors.New("oidc token already exchanged")

// genericExchangeError is returned for every auth-path failure so responses
// don't reveal whether an org, policy, or issuer is configured.
func genericExchangeError() stderr.ErrAuthentication {
	return stderr.ErrAuthentication{
		Err:         errors.New("authentication failed"),
		Description: "failed to verify OIDC token",
	}
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
	issuer, err := oidcverify.UnverifiedIssuer(tokenStr)
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
	var matched []*app.OIDCTrustPolicy
	var claims map[string]any

	for i := range policies {
		policy := &policies[i]

		verified, seen := verifiedAudiences[policy.Audience]
		if !seen {
			tokenClaims, err := s.verifier.Verify(reqCtx, tokenStr, policy.IssuerURL, policy.Audience)
			verified = err == nil
			verifiedAudiences[policy.Audience] = verified
			if err != nil {
				s.l.Warn("oidc exchange: token verification failed",
					zap.String("org_id", req.OrgID),
					zap.String("issuer", issuer),
					zap.String("policy_id", policy.ID),
					zap.Error(err))
			} else if claims == nil {
				claims = tokenClaims
			}
		}
		if !verified {
			continue
		}

		if matchClaims(policy.ClaimConditions, claims) {
			matched = append(matched, policy)
		}
	}

	// More than one match is ambiguous: an older broad policy must not silently
	// win over a newer narrow one and hand out a different role.
	if len(matched) > 1 {
		ids := make([]string, len(matched))
		for i, policy := range matched {
			ids[i] = policy.ID
		}
		s.l.Warn("oidc exchange: token matched multiple trust policies",
			zap.String("org_id", req.OrgID),
			zap.String("issuer", issuer),
			zap.Strings("policy_ids", ids))
		ctx.Error(genericExchangeError())
		return
	}

	if len(matched) == 0 {
		fields := []zap.Field{
			zap.String("org_id", req.OrgID),
			zap.String("issuer", issuer),
		}
		// Log incoming claims to help diagnose why no policy's conditions matched,
		// without weakening the generic client-facing 401.
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
	policy := matched[0]
	metricTags["policy_id"] = policy.ID

	now := time.Now()
	sub, _ := claims["sub"].(string)
	jti, _ := claims["jti"].(string)

	var sourceExpiry time.Time
	if exp, ok := claims["exp"].(float64); ok {
		sourceExpiry = time.Unix(int64(exp), 0)
	}

	tokenHash := sha256.Sum256([]byte(tokenStr))

	var token *app.Token
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		exchanged := app.OIDCExchangedToken{
			OrgID:     req.OrgID,
			IssuerURL: issuer,
			TokenHash: hex.EncodeToString(tokenHash[:]),
			ExpiresAt: sourceExpiry,
		}
		if err := tx.Create(&exchanged).Error; err != nil {
			if request.IsDuplicateKey(err) {
				return errAlreadyExchanged
			}
			return fmt.Errorf("recording exchanged token: %w", err)
		}

		issued, err := s.issueFederatedToken(tx, policy)
		if err != nil {
			return err
		}
		token = issued
		return nil
	})
	if errors.Is(err, errAlreadyExchanged) {
		s.l.Warn("oidc exchange: token was already exchanged",
			zap.String("org_id", req.OrgID),
			zap.String("issuer", issuer),
			zap.String("jti", jti))
		ctx.Error(genericExchangeError())
		return
	}
	if err != nil {
		s.l.Error("oidc exchange: failed to create token",
			zap.String("policy_id", policy.ID),
			zap.Error(err))
		ctx.Error(stderr.ErrSystem{
			Err:         errors.New("internal error"),
			Description: "failed to issue authentication token",
		})
		return
	}

	if err := s.db.WithContext(ctx).
		Model(&app.OIDCTrustPolicy{}).
		Where(app.OIDCTrustPolicy{ID: policy.ID}).
		Update("last_used_at", now).Error; err != nil {
		s.l.Warn("oidc exchange: failed to update last_used_at", zap.String("policy_id", policy.ID), zap.Error(err))
	}

	metricTags["status"] = "ok"
	s.l.Info("oidc exchange: token issued",
		zap.String("org_id", policy.OrgID),
		zap.String("policy_id", policy.ID),
		zap.String("issuer", policy.IssuerURL),
		zap.String("sub", sub),
		zap.String("role", policy.Role))

	ctx.JSON(http.StatusOK, ExchangeOIDCTokenResponse{
		Authenticated: true,
		Token:         token.Token,
		ExpiresAt:     token.ExpiresAt,
		OrgID:         policy.OrgID,
		TrustPolicyID: policy.ID,
		Role:          policy.Role,
	})
}
