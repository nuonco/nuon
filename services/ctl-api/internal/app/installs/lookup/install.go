package lookup

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func ByRef(ctx context.Context, db *gorm.DB, orgID, ref string) (*app.Install, error) {
	query := db.WithContext(ctx).
		Where(&app.Install{OrgID: orgID}, "OrgID").
		Session(&gorm.Session{})
	var install app.Install
	err := query.Where(&app.Install{ID: ref}, "ID").First(&install).Error
	if err == nil {
		return &install, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var installs []app.Install
	if err := query.Where(&app.Install{Name: ref}, "Name").Limit(2).Find(&installs).Error; err != nil {
		return nil, err
	}
	switch len(installs) {
	case 0:
		return nil, gorm.ErrRecordNotFound
	case 1:
		return &installs[0], nil
	default:
		return nil, stderr.ErrUser{Err: fmt.Errorf("install name %q is ambiguous; use an install ID", ref)}
	}
}

func OnApp(ctx context.Context, db *gorm.DB, orgID, appID, ref string) (*app.Install, error) {
	return ByRef(ctx, db.Where(&app.Install{AppID: appID}, "AppID"), orgID, ref)
}
