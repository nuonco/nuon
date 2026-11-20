package lookup

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func ByRef(ctx context.Context, db *gorm.DB, orgID, ref string) (*app.App, error) {
	if a, ok := cctx.ResolvedAppFromContext(ctx, orgID, ref); ok {
		return a, nil
	}

	var a app.App
	err := db.WithContext(ctx).
		Where(&app.App{OrgID: orgID}, "OrgID").
		Where(db.Where(&app.App{Name: ref}, "Name").Or(&app.App{ID: ref}, "ID")).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}
