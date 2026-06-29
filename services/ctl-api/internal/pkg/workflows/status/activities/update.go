package statusactivities

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func isTerminalStatus(s app.Status) bool {
	switch s {
	case app.StatusSuccess, app.StatusError, app.StatusCancelled:
		return true
	}
	return false
}

// NOTE
//
// This package is the beginning of consolidating all status logic into a single package.
//
// Right now, it's a bit verbose with getters for statuses when updating, however long term we can either generate this
// or make the status selectable in isolation by selecting the field using reflection or something else.
//
// However, for now, this interface provides a few things:
// 1. ability to manage history of a status
// 2. ability to start doing things such as sending a signal to a channel if needed. This enables the ability to start
// blocking for a "status" change or a specific status.
type UpdateStatusRequest struct {
	ID     string              `validate:"required"`
	Status app.CompositeStatus `json:"status" validate:"required"`
}

// TODO(sdboyer) remove after workflow refactor
// @temporal-gen-v2 activity
func (a *Activities) PkgStatusUpdateInstallWorkflowStatus(ctx context.Context, req UpdateStatusRequest) error {
	obj := app.Workflow{
		ID: req.ID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.Workflow
		if err := a.getStatus(ctx, &obj, req.ID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.Status, nil
	}

	return a.updateStatus(ctx, &obj, req.Status, getter)
}

// TODO(sdboyer) remove after workflow refactor
// @temporal-gen-v2 activity
func (a *Activities) PkgStatusUpdateInstallWorkflowStepStatus(ctx context.Context, req UpdateStatusRequest) error {
	obj := app.WorkflowStep{
		ID: req.ID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.WorkflowStep
		if err := a.getStatus(ctx, &obj, req.ID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.Status, nil
	}

	return a.updateStatus(ctx, &obj, req.Status, getter)
}

// @temporal-gen-v2 activity
func (a *Activities) PkgStatusUpdateInstallStackVersionStatus(ctx context.Context, req UpdateStatusRequest) error {
	obj := app.InstallStackVersion{
		ID: req.ID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.InstallStackVersion
		if err := a.getStatus(ctx, &obj, req.ID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.Status, nil
	}

	return a.updateStatus(ctx, &obj, req.Status, getter)
}

func (a *Activities) updateStatus(ctx context.Context, obj any, status app.CompositeStatus, statusGetter func(ctx context.Context) (app.CompositeStatus, error)) error {
	return a.updateStatusCommon(ctx, obj, status, statusGetter, "status")
}

func (a *Activities) updateStatusV2(ctx context.Context, obj any, status app.CompositeStatus, statusGetter func(ctx context.Context) (app.CompositeStatus, error)) error {
	return a.updateStatusCommon(ctx, obj, status, statusGetter, "status_v2")
}

func (a *Activities) updateStatusCommon(ctx context.Context, obj any, status app.CompositeStatus, statusGetter func(ctx context.Context) (app.CompositeStatus, error), statusField string) error {
	createdBy, err := cctx.AccountIDFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get created by")
	}

	status.CreatedByID = createdBy
	status.CreatedAtTS = time.Now().Unix()

	existingStatus, err := statusGetter(ctx)
	if err != nil {
		return err
	}
	history := existingStatus.History
	existingStatus.History = nil
	history = append(history, existingStatus)
	// Limit history to the most recent 25 entries to prevent unbounded growth.
	if len(history) > 25 {
		history = history[len(history)-25:]
	}
	status.History = history

	// Carry forward existing metadata into the new status so it's not lost.
	newMetadata := status.Metadata
	if newMetadata == nil {
		newMetadata = make(map[string]any, 0)
	}
	for k, v := range existingStatus.Metadata {
		if _, ok := newMetadata[k]; ok {
			continue
		}
		newMetadata[k] = v
	}
	status.Metadata = newMetadata

	res := a.db.WithContext(ctx).Model(obj).Updates(
		map[string]any{
			statusField: status,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update")
	}
	if res.RowsAffected < 1 {
		return errors.New("no object found to update")
	}

	// Also atomically merge new metadata keys via jsonb_set so that any
	// metadata written concurrently (e.g. by the background enqueuer)
	// between our read and this write is not lost.
	if len(status.Metadata) > 0 {
		id := reflectID(obj)
		if err := generics.MergeJSONBMetadata(a.db.WithContext(ctx), obj, id, statusField, status.Metadata); err != nil {
			return errors.Wrap(err, "unable to merge metadata")
		}
	}

	return nil
}

// reflectID extracts the ID field from a model pointer via reflection.
func reflectID(obj any) string {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.FieldByName("ID").String()
}

// @temporal-gen-v2 activity
func (a *Activities) PkgStatusUpdateFlowStatus(ctx context.Context, req UpdateStatusRequest) error {
	obj := app.Workflow{
		ID: req.ID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.Workflow
		if err := a.getStatus(ctx, &obj, req.ID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.Status, nil
	}

	return a.updateStatus(ctx, &obj, req.Status, getter)
}

// @temporal-gen-v2 activity
func (a *Activities) PkgStatusUpdateFlowStepStatus(ctx context.Context, req UpdateStatusRequest) error {
	obj := app.WorkflowStep{
		ID: req.ID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.WorkflowStep
		if err := a.getStatus(ctx, &obj, req.ID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.Status, nil
	}

	return a.updateStatus(ctx, &obj, req.Status, getter)
}

// @temporal-gen-v2 activity
func (a *Activities) PkgStatusUpdateFlowStepGroupStatus(ctx context.Context, req UpdateStatusRequest) error {
	obj := app.WorkflowStepGroup{
		ID: req.ID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.WorkflowStepGroup
		if err := a.getStatus(ctx, &obj, req.ID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.Status, nil
	}

	return a.updateStatus(ctx, &obj, req.Status, getter)
}

type UpdateBuildStatusV2 struct {
	BuildID           string                   `validate:"required"`
	Status            app.ComponentBuildStatus `validate:"required"`
	StatusDescription string                   `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateBuildStatusV2(ctx context.Context, req UpdateBuildStatusV2) error {
	obj := app.ComponentBuild{
		ID: req.BuildID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.ComponentBuild
		if err := a.getStatus(ctx, &obj, req.BuildID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateInstallWorkflowRunStatusV2Request struct {
	RunID             string                             `validate:"required"`
	Status            app.InstallActionWorkflowRunStatus `validate:"required"`
	StatusDescription string                             `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunID
func (a *Activities) UpdateInstallWorkflowRunStatusV2(ctx context.Context, req UpdateInstallWorkflowRunStatusV2Request) error {
	install := app.InstallActionWorkflowRun{
		ID: req.RunID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var install app.InstallActionWorkflowRun
		if err := a.getStatus(ctx, &install, req.RunID); err != nil {
			return app.CompositeStatus{}, err
		}
		return install.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	return a.updateStatusV2(ctx, &install, status, getter)
}

type UpdateRunStatusV2Request struct {
	RunID             string               `validate:"required"`
	Status            app.SandboxRunStatus `validate:"required"`
	StatusDescription string               `validate:"required"`
	SkipStatusSync    bool
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateRunStatusV2(ctx context.Context, req UpdateRunStatusV2Request) error {
	install := app.InstallSandboxRun{
		ID: req.RunID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var install app.InstallSandboxRun
		if err := a.getStatus(ctx, &install, req.RunID); err != nil {
			return app.CompositeStatus{}, err
		}
		return install.StatusV2, nil
	}

	compStatus := app.NewCompositeStatus(ctx, app.Status(req.Status))
	compStatus.StatusHumanDescription = req.StatusDescription

	err := a.updateStatusV2(ctx, &install, compStatus, getter)
	if err != nil {
		return fmt.Errorf("unable to update run status: %w", err)
	}

	run := app.InstallSandboxRun{}
	res := a.db.WithContext(ctx).
		Where("id = ?", req.RunID).
		First(&run)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("unable to get install sandbox run: %w", res.Error)
	}

	installObj := &app.Install{}
	res = a.db.WithContext(ctx).
		Where("id = ?", run.InstallID).
		First(installObj)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("unable to get install: %w", res.Error)
	}

	installSandbox := app.InstallSandbox{}
	res = a.db.WithContext(ctx).
		Where("install_id = ?", installObj.ID).
		First(&installSandbox)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("unable to get install sandbox: %w", res.Error)
	}

	if res.Error == gorm.ErrRecordNotFound {
		return nil
	}

	runUpdate := app.InstallSandboxRun{
		ID: req.RunID,
	}
	res = a.db.WithContext(ctx).Model(&runUpdate).Updates(app.InstallSandboxRun{
		InstallSandboxID: &installSandbox.ID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install sandbox run with sandbox ID: %w", res.Error)
	}

	if !req.SkipStatusSync {
		getter := func(ctx context.Context) (app.CompositeStatus, error) {
			var ninstallSandbox app.InstallSandbox
			if err := a.getStatus(ctx, &ninstallSandbox, installSandbox.ID); err != nil {
				return app.CompositeStatus{}, err
			}
			return ninstallSandbox.StatusV2, nil
		}

		err = a.updateStatusV2(ctx, &installSandbox, compStatus, getter)
		if err != nil {
			return fmt.Errorf("unable to update install sandbox: %w", err)
		}
	}
	return nil
}

type UpdateOrgStatusV2Request struct {
	OrgID             string        `validate:"required"`
	Status            app.OrgStatus `validate:"required"`
	StatusDescription string        `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateOrgStatusV2(ctx context.Context, req UpdateOrgStatusV2Request) error {
	obj := app.Org{ID: req.OrgID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.Org
		if err := a.getStatus(ctx, &obj, req.OrgID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateAppStatusV2Request struct {
	AppID             string        `validate:"required"`
	Status            app.AppStatus `validate:"required"`
	StatusDescription string        `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateAppStatusV2(ctx context.Context, req UpdateAppStatusV2Request) error {
	obj := app.App{ID: req.AppID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.App
		if err := a.getStatus(ctx, &obj, req.AppID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateComponentStatusV2Request struct {
	ComponentID       string              `validate:"required"`
	Status            app.ComponentStatus `validate:"required"`
	StatusDescription string              `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateComponentStatusV2(ctx context.Context, req UpdateComponentStatusV2Request) error {
	obj := app.Component{ID: req.ComponentID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.Component
		if err := a.getStatus(ctx, &obj, req.ComponentID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateRunnerStatusV2Request struct {
	RunnerID          string           `validate:"required"`
	Status            app.RunnerStatus `validate:"required"`
	StatusDescription string           `validate:"required"`
	Metadata          map[string]any
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateRunnerStatusV2(ctx context.Context, req UpdateRunnerStatusV2Request) error {
	obj := app.Runner{ID: req.RunnerID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.Runner
		if err := a.getStatus(ctx, &obj, req.RunnerID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	for k, v := range req.Metadata {
		status.Metadata[k] = v
	}
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateActionWorkflowStatusV2Request struct {
	ActionWorkflowID  string                   `validate:"required"`
	Status            app.ActionWorkflowStatus `validate:"required"`
	StatusDescription string                   `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateActionWorkflowStatusV2(ctx context.Context, req UpdateActionWorkflowStatusV2Request) error {
	obj := app.ActionWorkflow{ID: req.ActionWorkflowID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.ActionWorkflow
		if err := a.getStatus(ctx, &obj, req.ActionWorkflowID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateRunnerOperationStatusV2Request struct {
	RunnerOperationID string                    `validate:"required"`
	Status            app.RunnerOperationStatus `validate:"required"`
	StatusDescription string                    `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateRunnerOperationStatusV2(ctx context.Context, req UpdateRunnerOperationStatusV2Request) error {
	obj := app.RunnerOperation{ID: req.RunnerOperationID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.RunnerOperation
		if err := a.getStatus(ctx, &obj, req.RunnerOperationID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateRunnerJobStatusV2Request struct {
	RunnerJobID       string              `validate:"required"`
	Status            app.RunnerJobStatus `validate:"required"`
	StatusDescription string              `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateRunnerJobStatusV2(ctx context.Context, req UpdateRunnerJobStatusV2Request) error {
	obj := app.RunnerJob{ID: req.RunnerJobID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.RunnerJob
		if err := a.getStatus(ctx, &obj, req.RunnerJobID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateRunnerJobExecutionStatusV2Request struct {
	RunnerJobExecutionID string                       `validate:"required"`
	Status               app.RunnerJobExecutionStatus `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateRunnerJobExecutionStatusV2(ctx context.Context, req UpdateRunnerJobExecutionStatusV2Request) error {
	obj := app.RunnerJobExecution{ID: req.RunnerJobExecutionID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.RunnerJobExecution
		if err := a.getStatus(ctx, &obj, req.RunnerJobExecutionID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateInstallActionWorkflowRunStepStatusV2Request struct {
	StepID string                                 `validate:"required"`
	Status app.InstallActionWorkflowRunStepStatus `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateInstallActionWorkflowRunStepStatusV2(ctx context.Context, req UpdateInstallActionWorkflowRunStepStatusV2Request) error {
	obj := app.InstallActionWorkflowRunStep{ID: req.StepID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.InstallActionWorkflowRunStep
		if err := a.getStatus(ctx, &obj, req.StepID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateAppConfigStatusV2Request struct {
	AppConfigID       string              `validate:"required"`
	Status            app.AppConfigStatus `validate:"required"`
	StatusDescription string              `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateAppConfigStatusV2(ctx context.Context, req UpdateAppConfigStatusV2Request) error {
	obj := app.AppConfig{ID: req.AppConfigID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.AppConfig
		if err := a.getStatus(ctx, &obj, req.AppConfigID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	status.StatusHumanDescription = req.StatusDescription
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateOrgInviteStatusV2Request struct {
	OrgInviteID string              `validate:"required"`
	Status      app.OrgInviteStatus `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateOrgInviteStatusV2(ctx context.Context, req UpdateOrgInviteStatusV2Request) error {
	obj := app.OrgInvite{ID: req.OrgInviteID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.OrgInvite
		if err := a.getStatus(ctx, &obj, req.OrgInviteID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, app.Status(req.Status))
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateDeployStatusV2Request struct {
	DeployID          string     `validate:"required"`
	Status            app.Status `validate:"required"`
	StatusDescription string     `validate:"required"`
	SkipStatusSync    bool
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateDeployStatusV2(ctx context.Context, req UpdateDeployStatusV2Request) error {
	installDeploy := app.InstallDeploy{
		ID: req.DeployID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var installDeploy app.InstallDeploy
		if err := a.getStatus(ctx, &installDeploy, req.DeployID); err != nil {
			return app.CompositeStatus{}, err
		}
		return installDeploy.StatusV2, nil
	}

	compositeStatus := app.NewCompositeStatus(ctx, req.Status)
	compositeStatus.StatusHumanDescription = req.StatusDescription

	err := a.updateStatusV2(ctx, &installDeploy, compositeStatus, getter)
	if err != nil {
		return fmt.Errorf("unable to update install deploy: %w", err)
	}
	if req.SkipStatusSync {
		return nil
	}

	extantInstallDeploy := app.InstallDeploy{}
	res := a.db.WithContext(ctx).
		Preload("InstallComponent").
		Where("id = ?", req.DeployID).
		First(&extantInstallDeploy)
	if res.Error != nil {
		return fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	compositeStatus = app.NewCompositeStatus(ctx, req.Status)
	compositeStatus.StatusHumanDescription = req.StatusDescription
	installComponent := app.InstallComponent{
		ID: extantInstallDeploy.InstallComponent.ID,
	}

	getInstallComponentStatus := func(ctx context.Context) (app.CompositeStatus, error) {
		if err := a.getStatus(ctx, &installComponent, installComponent.ID); err != nil {
			return app.CompositeStatus{}, err
		}
		return installComponent.StatusV2, nil
	}

	if err := a.updateStatusV2(ctx, &installComponent, compositeStatus, getInstallComponentStatus); err != nil {
		return fmt.Errorf("unable to update install component: %w", err)
	}

	return nil
}

type UpdateQueueStatusV2Request struct {
	QueueID           string     `validate:"required"`
	Status            app.Status `validate:"required"`
	StatusDescription string
}

// @temporal-gen-v2 activity
// @by-field QueueID
func (a *Activities) UpdateQueueStatusV2(ctx context.Context, req UpdateQueueStatusV2Request) error {
	obj := app.Queue{ID: req.QueueID}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.Queue
		if err := a.getStatus(ctx, &obj, req.QueueID); err != nil {
			return app.CompositeStatus{}, err
		}
		return obj.StatusV2, nil
	}

	status := app.NewCompositeStatus(ctx, req.Status)
	if req.StatusDescription != "" {
		status.StatusHumanDescription = req.StatusDescription
	}
	return a.updateStatusV2(ctx, &obj, status, getter)
}

type UpdateInstallLifecycleStatusV2Request struct {
	InstallID         string                     `validate:"required"`
	Status            app.InstallLifecycleStatus `validate:"required"`
	StatusDescription string                     `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) UpdateInstallLifecycleStatusV2(ctx context.Context, req UpdateInstallLifecycleStatusV2Request) error {
	var install app.Install
	if err := a.getStatus(ctx, &install, req.InstallID); err != nil {
		return err
	}

	transitioned := lifecyclephase.Transition(install.LifecyclePhase, req.Status, req.StatusDescription)

	createdBy, err := cctx.AccountIDFromContext(ctx)
	if err == nil {
		transitioned.UpdatedBy = createdBy
	}

	res := a.db.WithContext(ctx).Model(&app.Install{ID: req.InstallID}).Updates(map[string]any{
		"lifecycle_phase": transitioned,
	})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update lifecycle phase")
	}
	if res.RowsAffected < 1 {
		return errors.New("no install found to update lifecycle phase")
	}
	return nil
}

type UpdateQueueSignalStatusV2Request struct {
	QueueSignalID     string     `validate:"required"`
	Status            app.Status `validate:"required"`
	StatusDescription string
	Metadata          map[string]any
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateQueueSignalStatusV2(ctx context.Context, req UpdateQueueSignalStatusV2Request) error {
	obj := app.QueueSignal{ID: req.QueueSignalID}

	var signalType string
	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.QueueSignal
		if err := a.getStatus(ctx, &obj, req.QueueSignalID); err != nil {
			return app.CompositeStatus{}, err
		}
		signalType = string(obj.Type)
		return obj.Status, nil
	}

	status := app.NewCompositeStatus(ctx, req.Status)
	if req.StatusDescription != "" {
		status.StatusHumanDescription = req.StatusDescription
	}
	for k, v := range req.Metadata {
		status.Metadata[k] = v
	}
	if err := a.updateStatus(ctx, &obj, status, getter); err != nil {
		return generics.TemporalGormError(err)
	}

	if isTerminalStatus(req.Status) {
		a.mw.Incr("queue.signal.processed", metrics.ToTags(map[string]string{
			"signal_type": signalType,
			"status":      string(req.Status),
		}))
	}
	return nil
}
