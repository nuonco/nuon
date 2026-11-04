package account

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

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
		var err error
		binding, err = client.EnsureManagedServiceAccount(ctx, ManagedServiceAccountRequest{
			OrgID:       orgID,
			OwnerType:   plugins.TableName(tx, app.InstallStack{}),
			OwnerID:     installStackID,
			Purpose:     app.ManagedServiceAccountPurposePhoneHome,
			InstanceKey: stackVersionID,
			Subject:     stackVersionID,
			Name:        "Stack version phone home",
		})
		if err != nil {
			return fmt.Errorf("ensure phone home service account: %w", err)
		}

		return client.authzClient.EnsureManagedServiceAccountRole(ctx, binding.AccountID, &app.Role{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypePhoneHome,
			Version:     app.RoleVersionV2,
			Title:       "Phone home",
			Description: "Scoped access to one install stack version callback.",
			Policies: []app.Policy{{
				OrgID: generics.NewNullString(orgID),
				Name:  app.PolicyNamePhoneHome,
				Permissions: pgtype.Hstore{
					permissions.APIPermissionKey:                                         generics.ToPtr(string(permissions.APIStack)),
					permissions.Object(orgID, permissions.KindPhoneHome, stackVersionID): permissions.PermissionCreate.ToStrPtr(),
				},
			}},
		})
	})
	if err != nil {
		return nil, err
	}
	return binding, nil
}
