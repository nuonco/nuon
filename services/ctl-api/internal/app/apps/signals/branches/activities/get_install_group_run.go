package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetInstallGroupRunInput struct {
	AppBranchRunID string `json:"app_branch_run_id"`
	InstallGroupID string `json:"install_group_id"`
}

type GetInstallGroupRunOutput struct {
	// Found is false when the deploy step never recorded a group run: an empty
	// install group auto-skipped by the emptygroup check, or a deploy step a user
	// skipped. Callers must read this as "no deploy happened" rather than an error.
	Found bool `json:"found"`

	InstallGroupRunID string                       `json:"install_group_run_id"`
	InstallGroupName  string                       `json:"install_group_name"`
	Installs          []app.InstallGroupRunInstall `json:"installs,omitempty"`
}

// GetInstallGroupRun loads the group run recorded by the deploy step. The
// post-deploy runbook step runs as a separate flow step, so the deploy outcome
// per install has to come from this row rather than in-memory state.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetInstallGroupRun(ctx context.Context, input *GetInstallGroupRunInput) (*GetInstallGroupRunOutput, error) {
	var groupRun app.InstallGroupRun
	err := a.db.WithContext(ctx).
		Where(app.InstallGroupRun{
			AppBranchRunID: input.AppBranchRunID,
			InstallGroupID: input.InstallGroupID,
		}).
		Order("created_at DESC").
		First(&groupRun).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &GetInstallGroupRunOutput{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get install group run: %w", err)
	}

	return &GetInstallGroupRunOutput{
		Found:             true,
		InstallGroupRunID: groupRun.ID,
		InstallGroupName:  groupRun.InstallGroupName,
		Installs:          groupRun.Installs,
	}, nil
}
