package lookup

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func ByRef(ctx context.Context, db *gorm.DB, orgID, ref string) (*app.Install, error) {
	var install app.Install
	err := db.WithContext(ctx).
		Where(&app.Install{OrgID: orgID}, "OrgID").
		Where(clause.Or(clause.Eq{Column: "name", Value: ref}, clause.Eq{Column: "id", Value: ref})).
		First(&install).Error
	if err != nil {
		return nil, err
	}
	return &install, nil
}

func OnApp(ctx context.Context, db *gorm.DB, orgID, appID, ref string) (*app.Install, error) {
	return ByRef(ctx, db.Where(&app.Install{AppID: appID}, "AppID"), orgID, ref)
}
