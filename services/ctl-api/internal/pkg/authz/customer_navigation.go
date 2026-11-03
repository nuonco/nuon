package authz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (c *Client) ReadCustomerNavigation(ctx context.Context, read func(*gorm.DB, []string) error) error {
	credential, ok := cctx.VerifiedCredentialFromContext(ctx)
	limits := permissions.CredentialLimits{Profile: credential.Profile, API: credential.Audience, Ceiling: credential.Ceiling}
	denied := stderr.ErrAuthorization{Err: fmt.Errorf("customer navigation denied"), Description: "access denied"}
	if !ok || !customerCredentialAllowed(credential) || limits.Validate() != nil {
		return denied
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var token app.Token
		if err := tx.Select("org_id").Where(app.Token{ID: credential.TokenID, AccountID: credential.AccountID}).Take(&token).Error; err != nil {
			return denied
		}
		ceiling, reason, err := loadCustomerCredential(tx, credential, token.OrgID, time.Now())
		if err != nil {
			return err
		}
		if reason != ReasonAllowed {
			return denied
		}
		orgIDs := make([]string, 0, len(credential.Ceiling))
		for _, grant := range credential.Ceiling {
			if token.OrgID == "" || token.OrgID == grant.Scope.OrgID {
				orgIDs = append(orgIDs, grant.Scope.OrgID)
			}
		}
		slices.Sort(orgIDs)
		visible := make([]string, 0, len(orgIDs))
		for _, orgID := range slices.Compact(orgIDs) {
			grants, err := loadScopedGrants(tx, credential.AccountID, orgID)
			if errors.Is(err, errNoScopedGrants) || errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			if err != nil {
				return err
			}
			if grants.API != credential.Audience {
				return denied
			}
			candidates := append(slices.Clone(grants.Grants), credential.Ceiling...)
			candidates = append(candidates, ceiling...)
			seen := make(map[permissions.Scope]bool)
			allowed := false
			for _, candidate := range candidates {
				scope := candidate.Scope
				if scope.OrgID != orgID || seen[scope] {
					continue
				}
				seen[scope] = true
				if err := validateBindingScope(scope); err != nil {
					return denied
				}
				target, err := resolveTarget(tx, scope)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				if err != nil {
					return err
				}
				for _, grant := range grants.Grants {
					for permission := range grant.Permissions {
						one := permissions.PermissionSet{permission: {}}
						if permissions.ContainsGrant(grants.Grants, target, one) &&
							permissions.ContainsGrant(credential.Ceiling, target, one) &&
							permissions.ContainsGrant(ceiling, target, one) {
							allowed = true
							break
						}
					}
					if allowed {
						break
					}
				}
				if allowed {
					break
				}
			}
			if allowed {
				visible = append(visible, orgID)
			}
		}
		return read(tx, visible)
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
}
