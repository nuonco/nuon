package account

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// TODO(jm): this entire file should probably live in `pkg/account`
func (m *Client) FetchAccount(ctx context.Context, acctID string) (*app.Account, error) {
	acct := app.Account{}
	res := m.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		First(&acct, "id = ?", acctID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to fetch account %s: %w", acctID, res.Error)
	}

	return &acct, nil
}
