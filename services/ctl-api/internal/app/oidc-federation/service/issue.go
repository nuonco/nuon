package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// issueFederatedToken mints the short-lived API token an exchange returns.
// Duration is clamped to the policy maximum.
func (s *service) issueFederatedToken(tx *gorm.DB, policy *app.OIDCTrustPolicy) (*app.Token, error) {
	duration := policy.TokenDurationSeconds
	if duration <= 0 {
		duration = app.OIDCTrustPolicyDefaultTokenDuration
	}
	if duration > app.OIDCTrustPolicyMaxTokenDuration {
		duration = app.OIDCTrustPolicyMaxTokenDuration
	}

	now := time.Now()
	token := app.Token{
		CreatedByID: policy.ServiceAccountID,
		Name:        policy.Name,
		Role:        policy.Role,
		OrgID:       policy.OrgID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeFederated,
		ExpiresAt:   now.Add(time.Duration(duration) * time.Second),
		IssuedAt:    now,
		Issuer:      policy.IssuerURL,
		AccountID:   policy.ServiceAccountID,
		SourceType:  app.TokenSourceTypeOIDCTrustPolicy,
		SourceID:    policy.ID,
	}
	if err := tx.Create(&token).Error; err != nil {
		return nil, fmt.Errorf("creating federated token: %w", err)
	}

	return &token, nil
}
