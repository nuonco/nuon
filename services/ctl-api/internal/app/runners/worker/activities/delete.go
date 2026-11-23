package activities

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

type DeleteRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := account.New(account.Params{DB: tx}).DeleteRunnerServiceAccounts(ctx, req.RunnerID); err != nil {
			return err
		}
		if err := tx.Select(clause.Associations).Delete(&app.Runner{ID: req.RunnerID}).Error; err != nil {
			return err
		}
		return tx.Where(app.RunnerCloudIdentity{RunnerID: req.RunnerID}).Delete(&app.RunnerCloudIdentity{}).Error
	})
}
