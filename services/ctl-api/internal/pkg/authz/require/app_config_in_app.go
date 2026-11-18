package require

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// AppConfigInApp returns not found unless appConfigID is a config of appID in
// the selected org. Config create bodies carry app_config_id, which AppInOrg
// never sees. An empty appConfigID is allowed because several creates treat
// it as optional.
func AppConfigInApp(ctx context.Context, db *gorm.DB, appID, appConfigID string) error {
	if appConfigID == "" {
		return nil
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return err
	}
	if orgID == "" || appID == "" {
		return errAppConfigNotFound()
	}

	var n int64
	err = db.WithContext(ctx).
		Model(&app.AppConfig{}).
		Where(app.AppConfig{OrgID: orgID, AppID: appID}).
		Where("id = ?", appConfigID).
		Limit(1).
		Count(&n).Error
	if err != nil {
		return fmt.Errorf("unable to authorize app config: %w", err)
	}
	if n < 1 {
		return errAppConfigNotFound()
	}

	return nil
}

func errAppConfigNotFound() error {
	return stderr.ErrNotFound{
		Err:         errors.New("app config not found"),
		Description: "app config not found",
	}
}
