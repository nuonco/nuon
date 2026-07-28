package migrations

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

// created by 01-create-internal-accounts, which runs before everything else
const migrationServiceAccount = "nuonctl"

// systemCtx attaches an acting account to ctx. BeforeCreate hooks pull created_by_id off
// the context, which a migration has none of, and the empty value fails the accounts FK.
func (m *Migrations) systemCtx(ctx context.Context) (context.Context, error) {
	acct, err := m.acctClient.FindAccount(ctx, account.ServiceAccountEmail(migrationServiceAccount))
	if err != nil {
		return nil, fmt.Errorf("unable to find %s service account: %w", migrationServiceAccount, err)
	}

	return context.WithValue(ctx, keys.AccountIDCtxKey, acct.ID), nil
}
