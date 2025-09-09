package statusactivities

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

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
// @temporal-gen activity
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
// @temporal-gen activity
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

// @temporal-gen activity
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
		return errors.Wrap(err, "unable to get existing status")
	}
	history := existingStatus.History
	existingStatus.History = nil
	status.History = append(history, existingStatus)

	if status.Metadata == nil {
		status.Metadata = make(map[string]any, 0)
	}
	for k, v := range existingStatus.Metadata {
		if _, ok := status.Metadata[k]; ok {
			continue
		}

		status.Metadata[k] = v
	}

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
	return nil
}

// @temporal-gen activity
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

// @temporal-gen activity
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

type UpdateBuildStatusV2 struct {
	BuildID           string                   `validate:"required"`
	Status            app.ComponentBuildStatus `validate:"required"`
	StatusDescription string                   `validate:"required"`
}

// @temporal-gen activity
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

// @temporal-gen activity
// @by-id RunID
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

// @temporal-gen activity
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

type UpdateDeployStatusV2Request struct {
	DeployID          string     `validate:"required"`
	Status            app.Status `validate:"required"`
	StatusDescription string     `validate:"required"`
	SkipStatusSync    bool
}

// @temporal-gen activity
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
