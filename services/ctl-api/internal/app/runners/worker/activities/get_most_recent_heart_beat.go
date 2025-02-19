package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetMostRecentHeartBeatRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen activity
// @by-id RunnerID
func (a *Activities) GetMostRecentHeartBeatRequest(ctx context.Context, req GetMostRecentHeartBeatRequest) (*app.RunnerHeartBeat, error) {
	hb, err := a.getMostRecentHeartBeat(ctx, req.RunnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner heart beat: %w", err)
	}

	return hb, nil
}

func (a *Activities) getMostRecentHeartBeat(ctx context.Context, runnerID string) (*app.RunnerHeartBeat, error) {
	var hb app.RunnerHeartBeat
	res := a.chDB.WithContext(ctx).
		Where(app.RunnerHeartBeat{
			RunnerID: runnerID,
		}).
		Order("created_at desc").
		Limit(1).
		First(&hb)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrap(res.Error, "unable to get heart beats")
	}

	return &hb, nil
}
