package migrations

import (
	"context"
	_ "embed"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (m *Migrations) Migration088AccountsEmailsNotEmpty(ctx context.Context, db *gorm.DB) error {
	err := m.DeleteAccountsAndSetUUIDAsEmail(ctx, db)
	if err != nil {
		return err
	}

	qry := `DO $$
BEGIN
    -- Check if constraint exists before adding
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'accounts_email_not_empty' 
        AND conrelid = 'accounts'::regclass
    ) THEN
        ALTER TABLE accounts
            ALTER COLUMN email SET NOT NULL,
            ADD CONSTRAINT accounts_email_not_empty CHECK (LENGTH(TRIM(email)) > 0);
    END IF;
END $$;`
	if res := db.WithContext(ctx).
		Exec(qry); res.Error != nil {
		return res.Error
	}
	return nil
}

func (m *Migrations) GetAccountsWithEmptyEmails(ctx context.Context, db *gorm.DB) ([]app.Account, error) {
	var accounts []app.Account
	res := db.WithContext(ctx).
		Where("email IS NULL OR email = ''").
		Find(&accounts)

	if res.Error != nil {
		return nil, res.Error
	}
	return accounts, nil
}

func (m *Migrations) DeleteAccountsAndSetUUIDAsEmail(ctx context.Context, db *gorm.DB) error {
	accounts, err := m.GetAccountsWithEmptyEmails(ctx, db)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		res := db.WithContext(ctx).
			Model(&app.Account{}).
			Where("id = ?", account.ID).
			Update("email", "deleted_"+uuid.NewString())

		if res.Error != nil {
			return res.Error
		}

		// soft delete the account
		res = db.WithContext(ctx).
			Delete(&app.Account{
				ID: account.ID,
			})

		if res.Error != nil {
			return res.Error
		}
	}

	return nil
}
