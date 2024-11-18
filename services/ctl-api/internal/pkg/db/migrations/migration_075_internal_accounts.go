package migrations

import (
	"context"
	_ "embed"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

func (a *Migrations) migration075InternalAccounts(ctx context.Context) error {
	svcAcctNames := []string{
		"github-actions",
		"nuonctl",
		"runner-local",
		"canary",
		"integration",
	}
	for _, svcAcctName := range svcAcctNames {
		_, err := a.acctClient.FindAccount(ctx, account.ServiceAccountEmail(svcAcctName))
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err, "unable to lookup"+svcAcctName)
		}

		_, err = a.acctClient.CreateServiceAccount(ctx, svcAcctName)
		if err != nil {
			return errors.Wrap(err, "unable to create "+svcAcctName)
		}

	}

	return nil
}
