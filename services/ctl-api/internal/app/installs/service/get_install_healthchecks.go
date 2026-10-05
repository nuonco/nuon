package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type InstallHealthcheck struct {
	ActionID   string     `json:"action_id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	LastRunAt  *time.Time `json:"last_run_at,omitempty"`
	WorkflowID string     `json:"workflow_id,omitempty"`
}

func (s *service) listInstallHealthchecks(ctx context.Context, orgID, installID, appConfigID string) ([]InstallHealthcheck, error) {
	checks := []InstallHealthcheck{}
	if appConfigID == "" {
		return checks, nil
	}

	var configs []app.ActionWorkflowConfig
	if err := s.db.WithContext(ctx).
		Preload("ActionWorkflow").
		Where(app.ActionWorkflowConfig{
			OrgID:         orgID,
			AppConfigID:   appConfigID,
			IsHealthcheck: true,
		}).
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("unable to list healthcheck actions: %w", err)
	}
	if len(configs) == 0 {
		return checks, nil
	}

	actionIDs := make([]string, 0, len(configs))
	for i := range configs {
		actionIDs = append(actionIDs, configs[i].ActionWorkflowID)
	}

	var installActions []app.InstallActionWorkflow
	if err := s.db.WithContext(ctx).
		Where(app.InstallActionWorkflow{OrgID: orgID, InstallID: installID}).
		Where("action_workflow_id IN ?", actionIDs).
		Find(&installActions).Error; err != nil {
		return nil, fmt.Errorf("unable to list install healthcheck actions: %w", err)
	}

	installActionByAction := make(map[string]string, len(installActions))
	installActionIDs := make([]string, 0, len(installActions))
	for i := range installActions {
		installActionByAction[installActions[i].ActionWorkflowID] = installActions[i].ID
		installActionIDs = append(installActionIDs, installActions[i].ID)
	}

	latestByInstallAction := map[string]app.InstallActionWorkflowRun{}
	if len(installActionIDs) > 0 {
		var runs []app.InstallActionWorkflowRun
		if err := s.db.WithContext(ctx).
			Where(app.InstallActionWorkflowRun{OrgID: orgID, InstallID: installID}).
			Where("install_action_workflow_id IN ?", installActionIDs).
			Order("created_at DESC").
			Find(&runs).Error; err != nil {
			return nil, fmt.Errorf("unable to list healthcheck action runs: %w", err)
		}
		for i := range runs {
			id := runs[i].InstallActionWorkflowID.String
			if id == "" || latestByInstallAction[id].ID != "" {
				continue
			}
			latestByInstallAction[id] = runs[i]
		}
	}

	for i := range configs {
		cfg := &configs[i]
		check := InstallHealthcheck{
			ActionID: cfg.ActionWorkflowID,
			Name:     cfg.ActionWorkflow.Name,
			Status:   string(app.InstallActionRunStatusUnknown),
		}
		run, ok := latestByInstallAction[installActionByAction[cfg.ActionWorkflowID]]
		if ok {
			check.Status = string(run.Status)
			ranAt := run.CreatedAt
			check.LastRunAt = &ranAt
			if run.InstallWorkflowID != nil {
				check.WorkflowID = *run.InstallWorkflowID
			}
		}
		checks = append(checks, check)
	}
	return checks, nil
}

func failingHealthchecks(checks []InstallHealthcheck) int {
	failed := 0
	for i := range checks {
		switch app.InstallActionWorkflowRunStatus(checks[i].Status) {
		case app.InstallActionRunStatusError, app.InstallActionRunStatusTimedOut:
			failed++
		}
	}
	return failed
}
