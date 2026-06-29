package activities

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ResolveInstallGroupInstallsInput struct {
	AppID    string           `json:"app_id"`
	GroupID  string           `json:"group_id"`
	Selector *labels.Selector `json:"selector"`
}

type ResolveInstallGroupInstallsOutput struct {
	InstallIDs []string `json:"install_ids"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ResolveInstallGroupInstalls(ctx context.Context, input *ResolveInstallGroupInstallsInput) (*ResolveInstallGroupInstallsOutput, error) {
	var installs []app.Install

	query := a.db.WithContext(ctx).
		Where(app.Install{AppID: input.AppID}).
		Scopes(labels.WithLabels("labels", input.Selector.MatchLabels)).
		Find(&installs)
	if query.Error != nil {
		return nil, query.Error
	}

	ids := make([]string, len(installs))
	for i, inst := range installs {
		ids[i] = inst.ID
	}

	a.l.Info("resolved install group via label selector",
		zap.String("group_id", input.GroupID),
		zap.Int("resolved_count", len(ids)),
	)

	return &ResolveInstallGroupInstallsOutput{InstallIDs: ids}, nil
}
