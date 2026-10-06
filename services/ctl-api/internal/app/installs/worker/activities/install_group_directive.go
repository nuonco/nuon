package activities

import (
	"context"
	"errors"
	"fmt"

	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installgrouprelease"
)

type ClassifyInstallGroupDirectiveRequest struct {
	InstallID      string `json:"install_id" validate:"required"`
	WorkflowID     string `json:"workflow_id" validate:"required"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
	AppBranchRunID string `json:"app_branch_run_id,omitempty"`
}

type ClassifyInstallGroupDirectiveResponse struct {
	Directive       string `json:"directive,omitempty"`
	Reason          string `json:"reason,omitempty"`
	WaitingOnRunID  string `json:"waiting_on_run_id,omitempty"`
	PriorWorkflowID string `json:"prior_workflow_id,omitempty"`
	WaitForPrior    bool   `json:"wait_for_prior,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ClassifyInstallGroupDirective(ctx context.Context, req ClassifyInstallGroupDirectiveRequest) (*ClassifyInstallGroupDirectiveResponse, error) {
	decision, err := a.helpers.ClassifyInstallGroupDirective(ctx, req.InstallID, req.WorkflowID, req.NewAppConfigID, req.AppBranchRunID)
	if err != nil {
		return nil, err
	}
	return &ClassifyInstallGroupDirectiveResponse{
		Directive:       decision.Directive,
		Reason:          decision.Reason,
		WaitingOnRunID:  decision.WaitingOnRunID,
		PriorWorkflowID: decision.PriorWorkflowID,
	}, nil
}

type SendInstallGroupDirectiveRequest struct {
	GroupWorkflowID string `json:"group_workflow_id"`
	Namespace       string `json:"namespace,omitempty"`
	InstallID       string `json:"install_id" validate:"required"`
	AppBranchRunID  string `json:"app_branch_run_id,omitempty"`
	Directive       string `json:"directive" validate:"required"`
	Reason          string `json:"reason,omitempty"`
	WaitingOnRunID  string `json:"waiting_on_run_id,omitempty"`
}

type SendInstallGroupDirectiveResponse struct{}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) SendInstallGroupDirective(ctx context.Context, req SendInstallGroupDirectiveRequest) (*SendInstallGroupDirectiveResponse, error) {
	if req.GroupWorkflowID == "" {
		groupID, namespace, runID, err := a.openGroupWorkflow(ctx, req.InstallID)
		if err != nil {
			return nil, err
		}
		req.GroupWorkflowID = groupID
		if req.Namespace == "" {
			req.Namespace = namespace
		}
		if req.AppBranchRunID == "" {
			req.AppBranchRunID = runID
		}
	}
	if req.GroupWorkflowID == "" {
		return &SendInstallGroupDirectiveResponse{}, a.helpers.RecordInstallGroupDirective(ctx, req.AppBranchRunID, req.InstallID, req.Directive, req.Reason, req.WaitingOnRunID)
	}

	var client tclient.Client = a.tClient
	if req.Namespace != "" {
		namespaced, err := a.tClient.GetNamespaceClient(req.Namespace)
		if err != nil {
			return nil, fmt.Errorf("unable to get temporal namespace: %w", err)
		}
		client = namespaced
	}

	handle, err := client.UpdateWorkflow(ctx, tclient.UpdateWorkflowOptions{
		WorkflowID:   req.GroupWorkflowID,
		UpdateName:   installgrouprelease.UpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{installgrouprelease.Directive{
			InstallID:      req.InstallID,
			Directive:      req.Directive,
			Reason:         req.Reason,
			WaitingOnRunID: req.WaitingOnRunID,
		}},
	})
	if err != nil {
		if recordErr := a.helpers.RecordInstallGroupDirective(ctx, req.AppBranchRunID, req.InstallID, req.Directive, req.Reason, req.WaitingOnRunID); recordErr != nil {
			return nil, errors.Join(err, recordErr)
		}
		return &SendInstallGroupDirectiveResponse{}, nil
	}
	var accepted struct{}
	if err := handle.Get(ctx, &accepted); err != nil {
		if recordErr := a.helpers.RecordInstallGroupDirective(ctx, req.AppBranchRunID, req.InstallID, req.Directive, req.Reason, req.WaitingOnRunID); recordErr != nil {
			return nil, errors.Join(err, recordErr)
		}
	}
	return &SendInstallGroupDirectiveResponse{}, nil
}

type InstallUpdateTerminalRequest struct {
	WorkflowID string `json:"workflow_id" validate:"required"`
}

type InstallUpdateTerminalResponse struct {
	Terminal bool `json:"terminal"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) InstallUpdateTerminal(ctx context.Context, req InstallUpdateTerminalRequest) (*InstallUpdateTerminalResponse, error) {
	terminal, err := a.helpers.InstallWorkflowTerminal(ctx, req.WorkflowID)
	if err != nil {
		return nil, err
	}
	return &InstallUpdateTerminalResponse{Terminal: terminal}, nil
}

type RecomputeInstallConfigDiffRequest struct {
	InstallID                 string `json:"install_id" validate:"required"`
	InstallAppConfigVersionID string `json:"install_config_update_id" validate:"required"`
	NewAppConfigID            string `json:"new_app_config_id" validate:"required"`
}

type RecomputeInstallConfigDiffResponse struct {
	Diff *app.InstallConfigDiff `json:"diff,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) RecomputeInstallConfigDiff(ctx context.Context, req RecomputeInstallConfigDiffRequest) (*RecomputeInstallConfigDiffResponse, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", req.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	diff, err := a.helpers.AppBranchConfigDiff(ctx, &install, req.NewAppConfigID)
	if err != nil {
		return nil, err
	}
	if err := a.helpers.SaveInstallConfigDiffBlob(ctx, req.InstallAppConfigVersionID, diff); err != nil {
		return nil, err
	}
	if diff != nil && diff.StackChanged {
		var version app.InstallAppConfigVersion
		if err := a.db.WithContext(ctx).First(&version, "id = ?", req.InstallAppConfigVersionID).Error; err != nil {
			return nil, fmt.Errorf("unable to get install config update: %w", err)
		}
		if version.WorkflowID != nil && *version.WorkflowID != "" {
			if err := a.db.WithContext(ctx).Exec(
				`UPDATE install_workflows SET metadata = COALESCE(metadata, ''::hstore) || hstore(?, 'true') WHERE id = ?`,
				string(app.WorkflowMetadataKeyStackChanged),
				*version.WorkflowID,
			).Error; err != nil {
				return nil, fmt.Errorf("unable to record stack change: %w", err)
			}
		}
	}
	return &RecomputeInstallConfigDiffResponse{Diff: diff}, nil
}

func (a *Activities) openGroupWorkflow(ctx context.Context, installID string) (workflowID, namespace, appBranchRunID string, err error) {
	var wf app.Workflow
	err = a.db.WithContext(ctx).
		Where(app.Workflow{OwnerID: installID, OwnerType: "installs", Type: app.WorkflowTypeAppBranchConfigUpdate}).
		Where("finished_at IS NULL").
		Order("created_at DESC").
		First(&wf).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", "", nil
		}
		return "", "", "", fmt.Errorf("unable to get install workflow: %w", err)
	}
	return genericsFrom(wf.Metadata["group_workflow_id"]), genericsFrom(wf.Metadata["group_workflow_namespace"]), genericsFrom(wf.Metadata["app_branch_run_id"]), nil
}

func genericsFrom(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
