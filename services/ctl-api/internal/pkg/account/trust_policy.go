package account

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (c *Client) CreateTrustPolicy(ctx context.Context, policy *app.OIDCTrustPolicy) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		client := c.withDB(tx)
		acct, err := client.CreateOrgServiceAccount(ctx, policy.OrgID, policy.OrgID+"-oidc-"+domains.NewAccountID(), "", app.RoleType(policy.Role))
		if err != nil {
			return err
		}
		policy.ServiceAccountID = acct.ID
		return tx.Create(policy).Error
	})
}

func (c *Client) UpdateTrustPolicy(ctx context.Context, policy *app.OIDCTrustPolicy) error {
	if policy == nil || policy.ID == "" {
		return fmt.Errorf("trust policy is required")
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := lockTrustPolicy(tx, policy.OrgID, policy.ID)
		if err != nil {
			return err
		}
		if current.ServiceAccountID != policy.ServiceAccountID || !current.UpdatedAt.Equal(policy.UpdatedAt) {
			return fmt.Errorf("trust policy changed during update; retry the request")
		}
		if current.Role != policy.Role {
			if err := c.withDB(tx).SetOrgRole(ctx, current.OrgID, current.ServiceAccountID, app.RoleType(policy.Role)); err != nil {
				return err
			}
		}
		if !policy.Enabled {
			if err := RevokeDerivedCredentials(tx, app.TokenSourceTypeOIDCTrustPolicy, []string{policy.ID}); err != nil {
				return err
			}
		}
		return tx.Model(policy).Select("name", "enabled", "issuer_url", "audience", "claim_conditions", "role", "token_duration_seconds", "updated_at").Updates(policy).Error
	})
}

func (c *Client) DeleteTrustPolicy(ctx context.Context, orgID, policyID string) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		policy, err := lockTrustPolicy(tx, orgID, policyID)
		if err != nil {
			return err
		}
		if err := RevokeDerivedCredentials(tx, app.TokenSourceTypeOIDCTrustPolicy, []string{policy.ID}); err != nil {
			return err
		}
		if err := tx.Delete(policy).Error; err != nil {
			return err
		}
		return deleteServiceAccount(tx, policy.ServiceAccountID)
	})
}

func lockTrustPolicy(tx *gorm.DB, orgID, policyID string) (*app.OIDCTrustPolicy, error) {
	if policyID == "" {
		return nil, fmt.Errorf("trust policy ID is required")
	}
	if err := lockOrg(tx, orgID); err != nil {
		return nil, err
	}
	var policy app.OIDCTrustPolicy
	if err := tx.Where(app.OIDCTrustPolicy{ID: policyID, OrgID: orgID}).Take(&policy).Error; err != nil {
		return nil, err
	}
	// Updates and deletes go on to change the account's roles or delete it under
	// FOR UPDATE; taking that strength now avoids upgrading a weaker lock.
	if err := tx.Select("id").Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(app.Account{ID: policy.ServiceAccountID}).Take(&app.Account{}).Error; err != nil {
		return nil, err
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.OIDCTrustPolicy{ID: policyID, OrgID: orgID, ServiceAccountID: policy.ServiceAccountID}).Take(&policy).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}
