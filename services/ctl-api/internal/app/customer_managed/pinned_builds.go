package customermanaged

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func LoadPinnedSandboxBuild(ctx context.Context, db *gorm.DB, orgID, appID, buildID string) (*app.AppSandboxBuild, error) {
	var build app.AppSandboxBuild
	if err := db.WithContext(ctx).Where(app.AppSandboxBuild{ID: buildID, OrgID: orgID, AppID: appID}).First(&build).Error; err != nil {
		return nil, err
	}
	return &build, nil
}
