package migrations

import "context"

func (a *Migrations) migration048AppSlackWebhookURL(ctx context.Context) error {
	sql := `
ALTER TABLE apps DROP COLUMN IF EXISTS slack_webhook_url;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
