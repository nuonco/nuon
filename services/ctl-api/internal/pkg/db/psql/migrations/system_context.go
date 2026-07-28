package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

// created by 01-create-internal-accounts, which runs before everything else
const migrationServiceAccount = "nuonctl"

// systemCtx attaches an acting account to ctx. BeforeCreate hooks pull created_by_id off
// the context, which a migration has none of, and the empty value fails the accounts FK.
//
// The lookup is unscoped: created_by_id only needs the account row to exist, and this
// account is soft deleted on installs where someone removed it.
func (m *Migrations) systemCtx(ctx context.Context, db *gorm.DB) (context.Context, error) {
	var acct app.Account
	res := db.WithContext(ctx).
		Unscoped().
		Where("email = ?", account.ServiceAccountEmail(migrationServiceAccount)).
		First(&acct)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find %s service account: %w", migrationServiceAccount, res.Error)
	}

	return context.WithValue(ctx, keys.AccountIDCtxKey, acct.ID), nil
}
