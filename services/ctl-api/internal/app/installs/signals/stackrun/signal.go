package stackrun

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/stategen"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/inputsupdated"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/rolechange"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	runnersignalsv2 "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/installstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "stack-run"

const installSignalsQueueName = "install-signals"

type Signal struct {
	InstallStackID        string `json:"install_stack_id"`
	InstallStackVersionID string `json:"install_stack_version_id,omitempty"`
	RunID                 string `json:"run_id"`
	RequestType           string `json:"request_type,omitempty"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }
func (s *Signal) AutoRetry() bool         { return true }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		Operation: "stack-run",
		OwnerID:   s.InstallStackID,
		OwnerType: "install_stacks",
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallStackID == "" {
		return fmt.Errorf("install stack id is required")
	}
	if s.RunID == "" {
		return fmt.Errorf("run id is required")
	}
	return nil
}

type roleSnapshot struct {
	Enabled     bool
	Provisioned bool
	RoleID      string
	RoleName    string
	RoleType    string
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	install, err := activities.AwaitGetInstallForStackByStackID(ctx, s.InstallStackID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	version, err := s.resolveVersion(ctx, install)
	if err != nil {
		return err
	}

	beforeRoles := snapshotRoles(ctx, install.ID, l)
	beforeInputs := snapshotInputs(ctx, version.InstallStackID)

	if err := s.processOutputs(ctx, install, version, l); err != nil {
		return err
	}

	afterRoles := snapshotRoles(ctx, install.ID, l)

	run, err := activities.AwaitGetInstallStackVersionRunByVersionID(ctx, version.ID)
	if err != nil {
		l.Warn("unable to get run for input diff", zap.Error(err))
	}

	var afterInputs map[string]string
	if run != nil {
		appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
		if err == nil {
			if so := decodeStackOutputs(appCfg, run); so != nil {
				afterInputs = extractInputValuesFromStackOutput(so)
			}
		}
	}

	isFirstRun, err := s.isFirstRun(ctx, version.ID)
	if err != nil {
		return errors.Wrap(err, "unable to check if first run")
	}

	if isFirstRun {
		s.handleProvisionComplete(ctx, install, version, l)
	}

	s.propagateToNewerVersion(ctx, install, version, l)

	diffAndEnqueueRoleChanges(ctx, install.ID, beforeRoles, afterRoles, l)
	diffAndEnqueueInputChanges(ctx, install.ID, beforeInputs, afterInputs, l)

	return nil
}

func (s *Signal) resolveVersion(ctx workflow.Context, install *app.Install) (*app.InstallStackVersion, error) {
	if s.InstallStackVersionID != "" {
		return activities.AwaitGetInstallStackVersionByID(ctx, activities.GetInstallStackVersionByIDRequest{
			VersionID: s.InstallStackVersionID,
		})
	}
	return activities.AwaitGetInstallStackVersionByInstallID(ctx, install.ID)
}

func (s *Signal) isFirstRun(ctx workflow.Context, versionID string) (bool, error) {
	resp, err := activities.AwaitCountInstallStackVersionRuns(ctx, activities.CountInstallStackVersionRunsRequest{
		VersionID: versionID,
	})
	if err != nil {
		return false, err
	}
	return resp.Count == 1, nil
}

func (s *Signal) handleProvisionComplete(ctx workflow.Context, install *app.Install, version *app.InstallStackVersion, l log.Logger) {
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   install.RunnerID,
		OwnerType: "runners",
		Signal: &runnersignalsv2.Signal{
			RunnerID:                 install.RunnerID,
			InstallStackVersionRunID: s.RunID,
		},
	})
	if err != nil {
		l.Warn("unable to enqueue signal to runner", zap.Error(err))
	}

	if err := statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
		ID:     version.ID,
		Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusActive),
	}); err != nil {
		l.Warn("unable to update stack version status", zap.Error(err))
	}

	cbResp, err := activities.AwaitGetInstallStackVersionCallback(ctx, activities.GetInstallStackVersionCallbackRequest{
		VersionID: version.ID,
	})
	if err != nil {
		l.Warn("unable to get callback ref", zap.Error(err))
		return
	}
	if cbResp.CallbackRef.IsSet() {
		callback.Send(ctx, nil, cbResp.CallbackRef, callback.Result{
			Status: "success",
		})
	}
}

func (s *Signal) propagateToNewerVersion(ctx workflow.Context, install *app.Install, phoneHomeVersion *app.InstallStackVersion, l log.Logger) {
	latestVersion, err := activities.AwaitGetInstallStackVersionByInstallID(ctx, install.ID)
	if err != nil {
		l.Warn("unable to get latest version for propagation", zap.Error(err))
		return
	}

	if latestVersion.ID == phoneHomeVersion.ID {
		return
	}

	run, err := activities.AwaitGetInstallStackVersionRunByVersionID(ctx, phoneHomeVersion.ID)
	if err != nil {
		l.Warn("unable to get phone home run for propagation", zap.Error(err))
		return
	}

	newRun, err := activities.AwaitCreateInstallStackVersionRun(ctx, activities.CreateInstallStackVersionRunRequest{
		InstallStackVersionID: latestVersion.ID,
		Data:                  run.Data,
	})
	if err != nil {
		l.Warn("unable to create run for newer version", zap.Error(err))
		return
	}

	if err := statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: latestVersion.ID,
		Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusActive, map[string]any{
			"applied_from_version_id": phoneHomeVersion.ID,
			"applied_from_run_id":     run.ID,
			"propagated_run_id":       newRun.ID,
		}),
	}); err != nil {
		l.Warn("unable to update newer version status", zap.Error(err))
	}

	cbResp, err := activities.AwaitGetInstallStackVersionCallback(ctx, activities.GetInstallStackVersionCallbackRequest{
		VersionID: latestVersion.ID,
	})
	if err != nil {
		l.Warn("unable to get newer version callback", zap.Error(err))
		return
	}
	if cbResp.CallbackRef.IsSet() {
		callback.Send(ctx, nil, cbResp.CallbackRef, callback.Result{
			Status: "success",
		})
	}
}

func (s *Signal) processOutputs(ctx workflow.Context, install *app.Install, version *app.InstallStackVersion, l log.Logger) error {
	run, err := activities.AwaitGetInstallStackVersionRunByVersionID(ctx, version.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get run outputs")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config by id")
	}

	installStackOutputs := app.InstallStackOutputs{}
	var stackOutputs app.StackOutput
	switch appCfg.RunnerConfig.Type {
	case app.AppRunnerTypeAWS:
		installStackOutputs.AWSStackOutputs = &app.AWSStackOutputs{}
		stackOutputs = installStackOutputs.AWSStackOutputs
	case app.AppRunnerTypeAzure:
		installStackOutputs.AzureStackOutputs = &app.AzureStackOutputs{}
		stackOutputs = installStackOutputs.AzureStackOutputs
	case app.AppRunnerTypeGCP:
		installStackOutputs.GCPStackOutputs = &app.GCPStackOutputs{}
		stackOutputs = installStackOutputs.GCPStackOutputs
	default:
		return nil
	}

	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
			pkggenerics.StringToMapDecodeHook(),
		),
		WeaklyTypedInput: true,
		Result:           stackOutputs,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return errors.Wrapf(err, "unable to create %s decoder", appCfg.RunnerConfig.Type)
	}
	if err := decoder.Decode(run.Data); err != nil {
		return errors.Wrapf(err, "unable to parse %s install outputs", appCfg.RunnerConfig.Type)
	}

	if err := activities.AwaitUpdateInstallStackOutputs(ctx, activities.UpdateInstallStackOutputs{
		InstallStackID:           version.InstallStackID,
		InstallStackVersionRunID: run.ID,
		Data:                     generics.ToStringMap(run.Data),
	}); err != nil {
		return errors.Wrap(err, "unable to update install stack outputs")
	}

	if err := activities.AwaitUpdateInstallRolesFromStackOutputs(ctx, activities.UpdateInstallRolesFromStackOutputs{
		InstallID: install.ID,
	}); err != nil {
		l.Warn("unable to update install roles from stack outputs", zap.Error(err))
	}

	runnerIAMRoleARN := ""
	if installStackOutputs.AWSStackOutputs != nil {
		runnerIAMRoleARN = installStackOutputs.AWSStackOutputs.RunnerIAMRoleARN
	}
	if err := activities.AwaitUpdateRunnerGroupSettings(ctx, &activities.UpdateRunnerGroupSettings{
		RunnerID:           install.RunnerID,
		LocalAWSIAMRoleARN: runnerIAMRoleARN,
	}); err != nil {
		return errors.Wrap(err, "unable to update runner group settings")
	}

	if installStackOutputs.GCPStackOutputs != nil && installStackOutputs.GCPStackOutputs.Region != "" {
		if err := activities.AwaitUpdateGCPAccountRegion(ctx, &activities.UpdateGCPAccountRegion{
			InstallID: install.ID,
			Region:    installStackOutputs.GCPStackOutputs.Region,
			ProjectID: installStackOutputs.GCPStackOutputs.ProjectID,
		}); err != nil {
			return errors.Wrap(err, "unable to update gcp account from stack outputs")
		}
	}

	if err := validateRegion(*install, installStackOutputs); err != nil {
		return errors.Wrap(err, "unable to validate region")
	}

	installInputValues, err := stackOutputs.InstallInputValues()
	if err != nil {
		return errors.Wrap(err, "unable to fetch install input values from stack outputs")
	}
	if len(installInputValues) > 0 {
		if err := activities.AwaitUpdateInstallInputsFromStack(ctx, &activities.UpdateInstallInputsFromStackRequest{
			InstallID:             install.ID,
			InputConfigID:         appCfg.InputConfig.ID,
			InputValues:           installInputValues,
			InstallStackVersionID: version.ID,
		}); err != nil {
			return errors.Wrap(err, "unable to update install inputs from stack outputs")
		}
	}

	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	if err := stategen.HintOrGenerate(ctx, stategen.Request{
		StateGenV2:      statemanager.UseStateGenV2(orgEnabled, install.Metadata),
		InstallID:       install.ID,
		Targets:         statemanager.TargetsForHint(statemanager.HintStackOutputsUpdated, ""),
		ForceAll:        true,
		TriggeredByID:   run.ID,
		TriggeredByType: "install_stack_version_runs",
	}); err != nil {
		return err
	}

	return nil
}

func snapshotRoles(ctx workflow.Context, installID string, l log.Logger) map[string]roleSnapshot {
	resp, err := activities.AwaitGetInstallRolesForInstall(ctx, activities.GetInstallRolesForInstallRequest{
		InstallID: installID,
	})
	if err != nil {
		l.Warn("unable to snapshot roles", zap.Error(err))
		return nil
	}

	snapshot := make(map[string]roleSnapshot, len(resp.Roles))
	for _, r := range resp.Roles {
		snapshot[r.ID] = roleSnapshot{
			Enabled:     r.Enabled,
			Provisioned: r.Provisioned,
			RoleID:      r.RoleID,
			RoleName:    r.AppRoleConfig.Name,
			RoleType:    string(r.AppRoleConfig.Type),
		}
	}
	return snapshot
}

func snapshotInputs(ctx workflow.Context, installStackID string) map[string]string {
	prevOutputs, err := activities.AwaitGetInstallStackOutputsByStackID(ctx, activities.GetInstallStackOutputsByStackIDRequest{
		InstallStackID: installStackID,
	})
	if err != nil || prevOutputs == nil {
		return nil
	}
	return extractInputValues(prevOutputs)
}

func diffAndEnqueueRoleChanges(ctx workflow.Context, installID string, before, after map[string]roleSnapshot, l log.Logger) {
	if before == nil || after == nil {
		return
	}

	for id, afterRole := range after {
		beforeRole, existed := before[id]
		if !existed {
			if afterRole.Provisioned {
				enqueueRoleChange(ctx, installID, afterRole, "enabled", l)
			}
			continue
		}
		if !beforeRole.Provisioned && afterRole.Provisioned {
			enqueueRoleChange(ctx, installID, afterRole, "enabled", l)
		} else if beforeRole.Provisioned && !afterRole.Provisioned {
			enqueueRoleChange(ctx, installID, afterRole, "disabled", l)
		}
	}

	for id, beforeRole := range before {
		if _, exists := after[id]; !exists && beforeRole.Provisioned {
			enqueueRoleChange(ctx, installID, beforeRole, "disabled", l)
		}
	}
}

func enqueueRoleChange(ctx workflow.Context, installID string, role roleSnapshot, changeType string, l log.Logger) {
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   installID,
		OwnerType: "installs",
		QueueName: installSignalsQueueName,
		Signal: &rolechange.Signal{
			InstallID:  installID,
			RoleName:   role.RoleName,
			RoleType:   role.RoleType,
			ChangeType: changeType,
			RoleID:     role.RoleID,
		},
	})
	if err != nil {
		l.Warn("unable to enqueue role-change signal", zap.Error(err))
	}
}

func diffAndEnqueueInputChanges(ctx workflow.Context, installID string, before, after map[string]string, l log.Logger) {
	if before == nil && after == nil {
		return
	}
	if before == nil {
		before = map[string]string{}
	}
	if after == nil {
		after = map[string]string{}
	}

	var added, removed, changed []string
	for k, v := range after {
		if oldV, ok := before[k]; !ok {
			added = append(added, k)
		} else if oldV != v {
			changed = append(changed, k)
		}
	}
	for k := range before {
		if _, ok := after[k]; !ok {
			removed = append(removed, k)
		}
	}

	if len(added) == 0 && len(removed) == 0 && len(changed) == 0 {
		return
	}

	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   installID,
		OwnerType: "installs",
		QueueName: installSignalsQueueName,
		Signal: &inputsupdated.Signal{
			InstallID:   installID,
			ChangedKeys: changed,
			AddedKeys:   added,
			RemovedKeys: removed,
		},
	})
	if err != nil {
		l.Warn("unable to enqueue inputs-updated signal", zap.Error(err))
	}
}

func extractInputValues(outputs *app.InstallStackOutputs) map[string]string {
	if outputs == nil {
		return nil
	}
	var so app.StackOutput
	switch {
	case outputs.AWSStackOutputs != nil:
		so = outputs.AWSStackOutputs
	case outputs.GCPStackOutputs != nil:
		so = outputs.GCPStackOutputs
	case outputs.AzureStackOutputs != nil:
		so = outputs.AzureStackOutputs
	default:
		return nil
	}
	vals, _ := so.InstallInputValues()
	return vals
}

func extractInputValuesFromStackOutput(so app.StackOutput) map[string]string {
	if so == nil {
		return nil
	}
	vals, _ := so.InstallInputValues()
	return vals
}

func decodeStackOutputs(appCfg *app.AppConfig, run *app.InstallStackVersionRun) app.StackOutput {
	var stackOutputs app.StackOutput
	switch appCfg.RunnerConfig.Type {
	case app.AppRunnerTypeAWS:
		stackOutputs = &app.AWSStackOutputs{}
	case app.AppRunnerTypeAzure:
		stackOutputs = &app.AzureStackOutputs{}
	case app.AppRunnerTypeGCP:
		stackOutputs = &app.GCPStackOutputs{}
	default:
		return nil
	}

	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
			pkggenerics.StringToMapDecodeHook(),
		),
		WeaklyTypedInput: true,
		Result:           stackOutputs,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil
	}
	if err := decoder.Decode(run.Data); err != nil {
		return nil
	}
	return stackOutputs
}

func validateRegion(install app.Install, outputs app.InstallStackOutputs) error {
	switch {
	case install.AWSAccount != nil:
		if install.AWSAccount.Region != "" && outputs.AWSStackOutputs != nil && install.AWSAccount.Region != outputs.AWSStackOutputs.Region {
			return errors.New("install stack was run for a different region than the install was configured for")
		}
	case install.AzureAccount != nil:
		if install.AzureAccount.Location != "" && outputs.AzureStackOutputs != nil && install.AzureAccount.Location != outputs.AzureStackOutputs.ResourceGroupLocation {
			return errors.New("install stack was run for a different region than the install was configured for")
		}
	}
	return nil
}
