package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetcomponentRecordsCount struct {
}

// @temporal-gen activity
func (a *Activities) GetUnknownComponentCount(ctx context.Context, req GetcomponentRecordsCount) (int64, error) {
	count, err := a.getUnkownComponentCount(ctx)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (a *Activities) getUnkownComponentCount(ctx context.Context) (int64, error) {
	var count int64
	err := a.db.WithContext(ctx).Model(&app.Component{}).Where("type IS NULL").Count(&count).Error
	if err != nil {
		return -1, err
	}
	return count, nil
}
