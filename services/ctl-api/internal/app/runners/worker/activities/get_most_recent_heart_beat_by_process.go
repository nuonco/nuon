package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetMostRecentHeartBeatByProcessRequest struct {
	RunnerID  string `validate:"required"`
	ProcessID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) GetMostRecentHeartBeatByProcess(ctx context.Context, req GetMostRecentHeartBeatByProcessRequest) (*app.RunnerHeartBeat, error) {
	var hb app.RunnerHeartBeat
	res := a.chDB.WithContext(ctx).
		Where(app.RunnerHeartBeat{
			RunnerID:  req.RunnerID,
			ProcessID: req.ProcessID,
		}).
		Order("created_at desc").
		Limit(1).
		First(&hb)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrap(res.Error, "unable to get heart beats by process")
	}

	return &hb, nil
}
