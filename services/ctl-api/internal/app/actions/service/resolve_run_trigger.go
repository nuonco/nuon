package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) resolveRunTriggers(ctx context.Context, runs []*app.InstallActionWorkflowRun) error {
	workflowIDs := make([]string, 0, len(runs))
	configIDs := make([]string, 0, len(runs))
	installIDs := make([]string, 0, len(runs))
	for _, run := range runs {
		installIDs = append(installIDs, run.InstallID)
		if run.InstallWorkflowID != nil && *run.InstallWorkflowID != "" {
			workflowIDs = append(workflowIDs, *run.InstallWorkflowID)
		}
		if !run.ActionWorkflowConfigID.Empty() {
			configIDs = append(configIDs, run.ActionWorkflowConfigID.String)
		}
	}

	workflows := map[string]app.Workflow{}
	runbookRuns := map[string]app.InstallRunbookRun{}
	if len(workflowIDs) > 0 {
		var wfs []app.Workflow
		if err := s.db.WithContext(ctx).
			Preload("CreatedBy").
			Where("id IN ?", workflowIDs).
			Find(&wfs).Error; err != nil {
			return fmt.Errorf("unable to get run workflows: %w", err)
		}
		for _, wf := range wfs {
			workflows[wf.ID] = wf
		}

		var rbRuns []app.InstallRunbookRun
		if err := s.db.WithContext(ctx).
			Preload("CreatedBy").
			Preload("InstallRunbook.Runbook").
			Where("install_workflow_id IN ?", workflowIDs).
			Find(&rbRuns).Error; err != nil {
			return fmt.Errorf("unable to get run runbook runs: %w", err)
		}
		for _, rbRun := range rbRuns {
			if rbRun.InstallWorkflowID != nil {
				runbookRuns[*rbRun.InstallWorkflowID] = rbRun
			}
		}
	}

	configs := map[string]app.ActionWorkflowConfig{}
	if len(configIDs) > 0 {
		var cfgs []app.ActionWorkflowConfig
		if err := s.db.WithContext(ctx).
			Select("id", "timeout").
			Where("id IN ?", configIDs).
			Find(&cfgs).Error; err != nil {
			return fmt.Errorf("unable to get run action configs: %w", err)
		}
		for _, cfg := range cfgs {
			configs[cfg.ID] = cfg
		}
	}

	installConfigs := map[string]*app.InstallConfig{}
	if len(installIDs) > 0 {
		var cfgs []app.InstallConfig
		if err := s.db.WithContext(ctx).
			Where("install_id IN ?", installIDs).
			Find(&cfgs).Error; err != nil {
			return fmt.Errorf("unable to get run install configs: %w", err)
		}
		for i := range cfgs {
			installConfigs[cfgs[i].InstallID] = &cfgs[i]
		}
	}

	for _, run := range runs {
		var wf *app.Workflow
		var rbRun *app.InstallRunbookRun
		if run.InstallWorkflowID != nil {
			if w, ok := workflows[*run.InstallWorkflowID]; ok {
				wf = &w
			}
			if r, ok := runbookRuns[*run.InstallWorkflowID]; ok {
				rbRun = &r
			}
		}
		run.Trigger = app.BuildActionRunTrigger(run, wf, rbRun)

		run.EffectiveTimeout = run.Timeout
		if !run.ActionWorkflowConfigID.Empty() {
			run.EffectiveTimeout = configs[run.ActionWorkflowConfigID.String].Timeout
		}
		run.EffectiveTimeout = installConfigs[run.InstallID].ResolveActionTimeout(run.EffectiveTimeout)
	}

	return nil
}
