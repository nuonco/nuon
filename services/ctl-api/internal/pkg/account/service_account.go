package account

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (c *Client) CreateServiceAccount(ctx context.Context, svcAcctID, name string) (*app.Account, error) {
	if svcAcctID == "" {
		return nil, errors.New("service account subject is required")
	}
	email := ServiceAccountEmail(svcAcctID)
	acct := app.Account{
		Email:       email,
		Subject:     svcAcctID,
		Name:        name,
		AccountType: app.AccountTypeService,
	}
	res := c.db.WithContext(ctx).
		Create(&acct)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create account")
	}

	return &acct, nil
}

// EnsureServiceAccount returns the service account for svcAcctID, creating it if it
// does not exist yet. CreateServiceAccount is a bare insert that conflicts on the
// unique email, so callers that may run against entities predating service-account
// creation need this instead.
func (c *Client) EnsureServiceAccount(ctx context.Context, svcAcctID, name string) (*app.Account, error) {
	if svcAcctID == "" {
		return nil, errors.New("service account subject is required")
	}
	acct, err := c.FindAccount(ctx, ServiceAccountEmail(svcAcctID))
	if err == nil {
		if acct.AccountType != app.AccountTypeService || acct.Subject != svcAcctID {
			return nil, errors.New("account does not match the service identity")
		}
		return acct, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "unable to look up service account")
	}

	return c.CreateServiceAccount(ctx, svcAcctID, name)
}

func (c *Client) restoreServiceAccount(ctx context.Context, req ManagedServiceAccountRequest) (*app.Account, error) {
	if req.OrgID == "" || req.Subject == "" {
		return nil, errors.New("restoration requires an org and service account subject")
	}
	if err := validateContextOrg(ctx, req.OrgID); err != nil {
		return nil, err
	}
	var acct app.Account
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Account{Email: ServiceAccountEmail(req.Subject), Subject: req.Subject}).
			Order("deleted_at = 0 DESC, deleted_at DESC").Limit(1).Find(&acct).Error; err != nil {
			return errors.Wrap(err, "load service account for restoration")
		}
		if acct.ID == "" {
			created, err := c.withDB(tx).CreateServiceAccount(ctx, req.Subject, req.Name)
			if err != nil {
				return err
			}
			acct = *created
			return nil
		}
		if acct.AccountType != app.AccountTypeService {
			return errors.New("account does not match the service identity")
		}
		var binding app.ManagedServiceAccount
		if err := tx.Unscoped().Where(app.ManagedServiceAccount{AccountID: acct.ID}).Find(&binding).Error; err != nil {
			return err
		}
		if binding.AccountID != "" && (binding.OrgID != req.OrgID || binding.OwnerType != req.OwnerType ||
			binding.OwnerID != req.OwnerID || binding.Purpose != req.Purpose || binding.InstanceKey != req.InstanceKey) {
			return errors.New("service identity belongs to a different owner")
		}
		if err := validateServiceAccountOrg(tx, acct.ID, req.OrgID); err != nil {
			return err
		}
		if acct.DeletedAt != 0 {
			if err := tx.Unscoped().Model(&app.Account{}).Where(app.Account{ID: acct.ID}).Update("deleted_at", 0).Error; err != nil {
				return err
			}
			acct.DeletedAt = 0
		}
		if binding.DeletedAt != 0 {
			if err := tx.Unscoped().Model(&app.ManagedServiceAccount{}).
				Where(app.ManagedServiceAccount{AccountID: acct.ID}).Update("deleted_at", 0).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return &acct, err
}
