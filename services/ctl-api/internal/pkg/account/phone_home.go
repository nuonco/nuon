package account

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

func (c *Client) CreateInstallStackVersion(ctx context.Context, version *app.InstallStackVersion) error {
	if version == nil || version.InstallStackID == "" || version.InstallID == "" {
		return fmt.Errorf("stack and install are required")
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stack app.InstallStack
		if err := tx.Where(app.InstallStack{ID: version.InstallStackID, InstallID: version.InstallID}).Take(&stack).Error; err != nil {
			return err
		}
		if version.OrgID != "" && version.OrgID != stack.OrgID {
			return fmt.Errorf("stack version belongs to another org")
		}
		if err := lockManagedAccountOwner(tx, ManagedServiceAccountRequest{OrgID: stack.OrgID, OwnerType: plugins.TableName(tx, app.InstallStack{}), OwnerID: stack.ID}); err != nil {
			return err
		}
		version.OrgID = stack.OrgID
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		_, err := c.withDB(tx).EnsurePhoneHomeServiceAccount(ctx, stack.OrgID, stack.ID, version.ID)
		return err
	})
}

func (c *Client) EnsurePhoneHomeServiceAccount(
	ctx context.Context,
	orgID string,
	installStackID string,
	stackVersionID string,
) (*app.ManagedServiceAccount, error) {
	var binding *app.ManagedServiceAccount
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		client := New(Params{
			DB:          tx,
			AuthzClient: authz.New(authz.Params{DB: tx}),
		})
		req := ManagedServiceAccountRequest{
			OrgID:       orgID,
			OwnerType:   plugins.TableName(tx, app.InstallStack{}),
			OwnerID:     installStackID,
			Purpose:     app.ManagedServiceAccountPurposePhoneHome,
			InstanceKey: stackVersionID,
			Subject:     stackVersionID,
			Name:        "Stack version phone home",
		}
		if err := lockManagedAccountOwner(tx, req); err != nil {
			return err
		}
		// A version that becomes eligible again gets back the account deleted
		// when it became ineligible.
		if _, err := client.restoreServiceAccount(ctx, req); err != nil {
			return err
		}
		var err error
		binding, err = client.EnsureManagedServiceAccount(ctx, req)
		if err != nil {
			return fmt.Errorf("ensure phone home service account: %w", err)
		}
		var version app.InstallStackVersion
		if err := tx.Select("id").Where(app.InstallStackVersion{ID: stackVersionID, InstallStackID: installStackID}).Take(&version).Error; err != nil {
			return err
		}
		return client.authzClient.EnsurePhoneHomeRole(ctx, orgID, stackVersionID, binding.AccountID)
	})
	if err != nil {
		return nil, err
	}
	return binding, nil
}

func (c *Client) IssuePhoneHomeToken(ctx context.Context, versionID string, duration time.Duration) (*app.Token, error) {
	if versionID == "" {
		return nil, fmt.Errorf("stack version ID is required")
	}
	var token *app.Token
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var version app.InstallStackVersion
		if err := tx.Where(app.InstallStackVersion{ID: versionID}).Take(&version).Error; err != nil {
			return err
		}
		if err := lockManagedAccountOwner(tx, ManagedServiceAccountRequest{OrgID: version.OrgID, OwnerType: plugins.TableName(tx, app.InstallStack{}), OwnerID: version.InstallStackID}); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.InstallStackVersion{ID: versionID}).Take(&version).Error; err != nil {
			return err
		}
		if !version.PhoneHomeTokenEligible() {
			return fmt.Errorf("stack version is not eligible for phone home credentials")
		}
		client := c.withDB(tx)
		binding, err := client.EnsurePhoneHomeServiceAccount(ctx, version.OrgID, version.InstallStackID, version.ID)
		if err != nil {
			return err
		}
		if version.PhoneHomeTokenID != "" {
			var existing app.Token
			if err := tx.Where(app.Token{ID: version.PhoneHomeTokenID, AccountID: binding.AccountID}).Where("expires_at > ?", time.Now()).Find(&existing).Error; err != nil {
				return err
			}
			if existing.ID != "" {
				token = &existing
				return nil
			}
		}
		token, err = client.CreateToken(ctx, binding.AccountID, duration)
		if err != nil {
			return err
		}
		return tx.Model(&version).Update("phone_home_token_id", token.ID).Error
	})
	return token, err
}

func (c *Client) RevokePhoneHomeToken(ctx context.Context, versionID string) error {
	if versionID == "" {
		return fmt.Errorf("stack version ID is required")
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var version app.InstallStackVersion
		if err := tx.Where(app.InstallStackVersion{ID: versionID}).Take(&version).Error; err != nil {
			return err
		}
		if err := lockManagedAccountOwner(tx, ManagedServiceAccountRequest{OrgID: version.OrgID, OwnerType: plugins.TableName(tx, app.InstallStack{}), OwnerID: version.InstallStackID}); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.InstallStackVersion{ID: versionID}).Take(&version).Error; err != nil {
			return err
		}
		if version.PhoneHomeTokenID != "" {
			if err := c.withDB(tx).RevokeToken(ctx, version.PhoneHomeTokenID); err != nil {
				return err
			}
		}
		return tx.Model(&version).Update("phone_home_token_id", "").Error
	})
}
