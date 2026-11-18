package helpers

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateworkflowsteps"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (s *Helpers) CreateWorkflowWithRole(
	ctx context.Context,
	installID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
	role string,
) (*app.Workflow, error) {
	return s.createWorkflow(
		ctx,
		installID,
		workflowType,
		metadata,
		planOnly,
		role,
		nil,
	)
}

func (s *Helpers) CreateWorkflow(
	ctx context.Context,
	installID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
	request *app.WorkflowRequest,
) (*app.Workflow, error) {
	return s.createWorkflow(
		ctx,
		installID,
		workflowType,
		metadata,
		planOnly,
		"",
		request,
	)
}

func (s *Helpers) createWorkflow(ctx context.Context,
	installID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
	role string,
	request *app.WorkflowRequest,
) (*app.Workflow, error) {
	installWorkflow, installName, err := s.prepareInstallWorkflow(ctx, installID, workflowType, metadata, planOnly, role, request)
	if err != nil {
		return nil, err
	}
	return s.insertInstallWorkflow(ctx, installWorkflow, installName)
}

func (s *Helpers) prepareInstallWorkflow(ctx context.Context,
	installID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
	role string,
	request *app.WorkflowRequest,
) (*app.Workflow, string, error) {
	if err := s.CheckInstallRunnerForWorkflow(ctx, installID, workflowType, metadata); err != nil {
		return nil, "", err
	}

	approvalOption := app.InstallApprovalOptionPrompt
	installConfig, err := s.GetLatestInstallConfig(ctx, installID)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to set approval option")
	}

	if installConfig != nil {
		approvalOption = installConfig.ApprovalOption
	}

	metadata["install_id"] = installID
	var install app.Install
	if res := s.db.WithContext(ctx).Where("id = ?", installID).First(&install); res.Error == nil && install.Name != "" {
		metadata[app.WorkflowMetadataKeyOwnerName] = install.Name
	}

	approvalType := ""
	if approvalOption == app.InstallApprovalOptionApproveAll {
		approvalType = "install-config"
	}
	installOption := approvalOption
	groupOption, err := appshelpers.GroupApprovalOptionWithDB(ctx, s.db, install.AppBranchID.ValueString(), install.AppBranchGroup)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to set approval option")
	}
	if groupOption == app.InstallApprovalOptionApproveAll {
		approvalOption = groupOption
		approvalType = "install-group"
	}
	if approvalType != "" {
		metadata["approval_type"] = approvalType
	}

	// #region agent log
	if f, ferr := os.OpenFile("/Users/robertbruce/nuonco/nuon/.cursor/debug-81e497.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); ferr == nil {
		payload, _ := json.Marshal(map[string]any{
			"sessionId":    "81e497",
			"hypothesisId": "group-overrides-install",
			"location":     "create_install_workflow.go:prepareInstallWorkflow",
			"message":      "resolved workflow approval option",
			"timestamp":    time.Now().UnixMilli(),
			"data": map[string]any{
				"install_id":      installID,
				"branch_id":       install.AppBranchID.ValueString(),
				"group":           install.AppBranchGroup,
				"install_option":  string(installOption),
				"group_option":    string(groupOption),
				"workflow_option": string(approvalOption),
				"approval_type":   approvalType,
				"workflow_type":   string(workflowType),
			},
		})
		_, _ = f.Write(append(payload, '\n'))
		_ = f.Close()
	}
	// #endregion

	status := app.NewCompositeStatus(ctx, app.StatusPending)
	if approvalType != "" {
		status.Metadata["approval_type"] = approvalType
	}

	installWorkflow := app.Workflow{
		Type:      workflowType,
		OwnerID:   installID,
		OwnerType: "installs",
		Metadata:  generics.ToHstore(metadata),
		Status:    status,
		// DEPRECATED: for now we always abort on step errors
		StepErrorBehavior: app.StepErrorBehaviorAbort,
		ApprovalOption:    approvalOption,
		PlanOnly:          planOnly,
		Role:              role,
		Request:           request,
		GenerateStepsSignal: &signaldb.SignalData{
			Signal: &generateworkflowsteps.Signal{},
		},
	}

	return &installWorkflow, install.Name, nil
}

func (s *Helpers) insertInstallWorkflow(ctx context.Context, installWorkflow *app.Workflow, installName string) (*app.Workflow, error) {
	res := s.db.WithContext(ctx).Create(installWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create install workflow")
	}

	cctx.GetLogger(ctx, s.l).Info("flow telemetry",
		zap.String("flow_event", "workflow.created"),
		zap.String("install_id", installWorkflow.OwnerID),
		zap.String("install_name", installName),
		zap.String("workflow_id", installWorkflow.ID),
		zap.String("workflow_type", string(installWorkflow.Type)),
		zap.Bool("plan_only", installWorkflow.PlanOnly),
	)

	return installWorkflow, nil
}

// CheckInstallRunnerForWorkflow returns the conflict that creating this
// workflow would hit, so callers can reject before persisting anything.
func (s *Helpers) CheckInstallRunnerForWorkflow(ctx context.Context, installID string, workflowType app.WorkflowType, metadata map[string]string) error {
	if inputsOnlyUpdate(workflowType, metadata) {
		return nil
	}

	if workflowType.RequiresInstallRunner() && !stackChangeDefersRunnerGate(workflowType, metadata) {
		disabled, err := s.IsRunnerDisabled(ctx, installID)
		if err != nil {
			return err
		}
		if disabled {
			return NewRunnerDisabledConflict()
		}
	}

	if requiresLiveInstallRunner(workflowType, metadata) {
		active, err := s.HasActiveRunner(ctx, installID)
		if err != nil {
			return err
		}
		if !active {
			return NewNoActiveRunnerConflict()
		}
	}
	return nil
}

func inputsOnlyUpdate(workflowType app.WorkflowType, metadata map[string]string) bool {
	return workflowType == app.WorkflowTypeInputUpdate && metadata[app.WorkflowMetadataKeyInputsOnly] == "true"
}

func requiresLiveInstallRunner(workflowType app.WorkflowType, metadata map[string]string) bool {
	if inputsOnlyUpdate(workflowType, metadata) {
		return false
	}
	if stackChangeDefersRunnerGate(workflowType, metadata) {
		return false
	}
	return workflowType.RequiresLiveInstallRunner()
}

// stackChangeDefersRunnerGate reports whether this workflow is an app-branch
// config update whose stack apply can bring the install runner back. Creation
// must not reject it for an offline or disabled runner.
func stackChangeDefersRunnerGate(workflowType app.WorkflowType, metadata map[string]string) bool {
	return workflowType == app.WorkflowTypeAppBranchConfigUpdate &&
		metadata[app.WorkflowMetadataKeyStackChanged] == "true"
}
