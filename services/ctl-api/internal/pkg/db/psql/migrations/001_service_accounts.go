package migrations

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"gorm.io/gorm"
)

func (m *Migrations) migration01InternalAccounts(ctx context.Context, db *gorm.DB) error {
	svcAcctNames := []string{
		"github-actions",
		"nuonctl",
		"runner-local",
		"canary",
		"integration",
	}
	for _, svcAcctName := range svcAcctNames {
		_, err := m.acctClient.FindAccount(ctx, account.ServiceAccountEmail(svcAcctName))
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err, "unable to lookup"+svcAcctName)
		}

		_, err = m.acctClient.CreateServiceAccount(ctx, svcAcctName)
		if err != nil {
			return errors.Wrap(err, "unable to create "+svcAcctName)
		}

	}

	return nil
}
