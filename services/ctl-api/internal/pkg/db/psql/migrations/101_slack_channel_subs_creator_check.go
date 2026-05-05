package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration101SlackChannelSubsCreatorCheck enforces that every
// slack_channel_subscriptions row has at least one creator identity recorded —
// either a Slack user (created via slash command) or a Nuon account (created
// via dashboard / API). GORM doesn't model multi-column CHECK constraints
// cleanly, so it lives here.
func (m *Migrations) Migration101SlackChannelSubsCreatorCheck(ctx context.Context, db *gorm.DB) error {
	qry := `DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'slack_channel_subscriptions_creator_present'
        AND conrelid = 'slack_channel_subscriptions'::regclass
    ) THEN
        ALTER TABLE slack_channel_subscriptions
            ADD CONSTRAINT slack_channel_subscriptions_creator_present
            CHECK (created_by_slack_user_id IS NOT NULL OR created_by_account_id IS NOT NULL);
    END IF;
END $$;`
	if res := db.WithContext(ctx).Exec(qry); res.Error != nil {
		return res.Error
	}
	return nil
}
