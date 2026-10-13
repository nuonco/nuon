package account

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type ManagedServiceAccountRequest struct {
	OrgID       string
	OwnerType   string
	OwnerID     string
	Purpose     app.ManagedServiceAccountPurpose
	InstanceKey string
	Subject     string
	Name        string
}

func (c *Client) EnsureManagedServiceAccount(ctx context.Context, req ManagedServiceAccountRequest) (*app.ManagedServiceAccount, error) {
	if req.OrgID == "" || req.OwnerType == "" || req.OwnerID == "" || req.Purpose == "" || req.InstanceKey == "" {
		return nil, fmt.Errorf("managed service account requires an org, owner, purpose and instance key")
	}

	var binding app.ManagedServiceAccount
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockManagedAccountOwner(tx, req); err != nil {
			return err
		}

		err := tx.Where(app.ManagedServiceAccount{
			OrgID: req.OrgID, OwnerType: req.OwnerType, OwnerID: req.OwnerID,
			Purpose: req.Purpose, InstanceKey: req.InstanceKey,
		}).First(&binding).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load managed service account: %w", err)
		}
		creating := errors.Is(err, gorm.ErrRecordNotFound)
		if creating {
			subject := req.Subject
			if subject == "" {
				subject = domains.NewAccountID()
			}
			var acct app.Account
			err := tx.Where(app.Account{Email: ServiceAccountEmail(subject), Subject: subject}).First(&acct).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				created, err := New(Params{DB: tx}).CreateServiceAccount(ctx, subject, req.Name)
				if err != nil {
					return err
				}
				acct = *created
			} else if err != nil {
				return fmt.Errorf("load service account for adoption: %w", err)
			}
			binding = app.ManagedServiceAccount{
				AccountID: acct.ID, OrgID: req.OrgID, OwnerType: req.OwnerType,
				OwnerID: req.OwnerID, Purpose: req.Purpose, InstanceKey: req.InstanceKey,
			}
		}

		var acct app.Account
		if err := tx.Select("id", "account_type", "subject").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Account{ID: binding.AccountID}).Take(&acct).Error; err != nil {
			return fmt.Errorf("lock managed service account: %w", err)
		}
		if acct.AccountType != app.AccountTypeService || (req.Subject != "" && acct.Subject != req.Subject) {
			return fmt.Errorf("account does not match the managed service identity")
		}

		var conflictingRoles int64
		if err := tx.Model(&app.AccountRole{}).
			Joins("JOIN roles ON roles.id = account_roles.role_id AND roles.deleted_at = 0").
			Where(app.AccountRole{AccountID: acct.ID}).
			Where("roles.org_id IS DISTINCT FROM ? OR account_roles.org_id IS DISTINCT FROM ?", req.OrgID, req.OrgID).
			Count(&conflictingRoles).Error; err != nil {
			return fmt.Errorf("validate service account tenant: %w", err)
		}
		if conflictingRoles != 0 {
			return fmt.Errorf("managed service account has roles outside its owning org")
		}
		if creating {
			if err := tx.Create(&binding).Error; err != nil {
				return fmt.Errorf("bind managed service account: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

func lockManagedAccountOwner(tx *gorm.DB, req ManagedServiceAccountRequest) error {
	orgType := plugins.TableName(tx, app.Org{})
	installType := plugins.TableName(tx, app.Install{})
	stackType := plugins.TableName(tx, app.InstallStack{})
	if req.OwnerType != orgType && req.OwnerType != installType && req.OwnerType != stackType {
		return fmt.Errorf("unsupported managed service account owner %q", req.OwnerType)
	}
	strength := "SHARE"
	if req.OwnerType == orgType {
		if req.OwnerID != req.OrgID {
			return fmt.Errorf("org-owned service account must belong to its owning org")
		}
		strength = "UPDATE"
	}
	var org struct{ ID string }
	if err := tx.Model(&app.Org{}).Select("id").
		Clauses(clause.Locking{Strength: strength}).Where(app.Org{ID: req.OrgID}).Take(&org).Error; err != nil {
		return fmt.Errorf("lock service account org: %w", err)
	}
	if req.OwnerType == orgType {
		return nil
	}

	installID := req.OwnerID
	if req.OwnerType == stackType {
		var stack struct{ InstallID string }
		if err := tx.Model(&app.InstallStack{}).Select("install_id").
			Where(app.InstallStack{ID: req.OwnerID, OrgID: req.OrgID}).Take(&stack).Error; err != nil {
			return fmt.Errorf("load service account stack: %w", err)
		}
		installID = stack.InstallID
	}
	var install struct{ ID string }
	if err := tx.Model(&app.Install{}).Scopes(scopes.WithDisableViews).Select("id").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(app.Install{ID: installID, OrgID: req.OrgID}).Take(&install).Error; err != nil {
		return fmt.Errorf("lock service account install: %w", err)
	}
	if req.OwnerType == stackType {
		var stack struct{ ID string }
		if err := tx.Model(&app.InstallStack{}).Select("id").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.InstallStack{ID: req.OwnerID, OrgID: req.OrgID, InstallID: installID}).Take(&stack).Error; err != nil {
			return fmt.Errorf("lock service account stack: %w", err)
		}
	}
	return nil
}
