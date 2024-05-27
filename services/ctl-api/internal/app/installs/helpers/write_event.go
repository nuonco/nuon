package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

func (h *Helpers) getInstallRun(ctx context.Context, runID string) (*app.InstallSandboxRun, error) {
	var run app.InstallSandboxRun
	res := h.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		First(&run, "id = ?", runID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install sandbox run: %w", res.Error)
	}

	return &run, nil
}

func (h *Helpers) getInstallDeploy(ctx context.Context, deployID string) (*app.InstallDeploy, error) {
	var deploy app.InstallDeploy
	res := h.db.WithContext(ctx).
		Preload("InstallComponent").
		Preload("InstallComponent.Component").
		Preload("CreatedBy").
		First(&deploy, "id = ?", deployID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	return &deploy, nil
}

func (h *Helpers) getInstall(ctx context.Context, installID string) (*app.Install, error) {
	var install app.Install
	res := h.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("App").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find install: %w", res.Error)
	}

	return &install, nil
}

func (h *Helpers) WriteDeployEvent(ctx context.Context,
	deployID string,
	op eventloop.SignalType,
	status app.OperationStatus,
) error {
	deploy, err := h.getInstallDeploy(ctx, deployID)
	if err != nil {
		return fmt.Errorf("unable to get install deploy: %w", err)
	}

	byts, err := json.Marshal(deploy)
	if err != nil {
		return fmt.Errorf("unable to marshal payload to json: %w", err)
	}

	ev := &app.InstallEvent{
		OrgID:           deploy.OrgID,
		CreatedByID:     deploy.CreatedByID,
		InstallID:       deploy.InstallID,
		Operation:       string(op),
		OperationStatus: status,
		Payload:         byts,
	}

	res := h.db.WithContext(ctx).Create(&ev)
	if res.Error != nil {
		return fmt.Errorf("unable to create install event: %w", res.Error)
	}

	return nil
}

func (h *Helpers) WriteInstallEvent(ctx context.Context,
	installID string,
	op eventloop.SignalType,
	status app.OperationStatus,
) error {
	install, err := h.getInstall(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get installs: %w", err)
	}

	byts, err := json.Marshal(install)
	if err != nil {
		return fmt.Errorf("unable to marshal payload to json: %w", err)
	}

	ev := &app.InstallEvent{
		OrgID:           install.OrgID,
		CreatedByID:     install.CreatedByID,
		InstallID:       installID,
		Operation:       string(op),
		OperationStatus: status,
		Payload:         byts,
	}

	res := h.db.WithContext(ctx).Create(&ev)
	if res.Error != nil {
		return fmt.Errorf("unable to create install event: %w", res.Error)
	}

	return nil
}

func (h *Helpers) WriteRunEvent(ctx context.Context,
	runID string,
	op eventloop.SignalType,
	status app.OperationStatus,
) error {
	run, err := h.getInstallRun(ctx, runID)
	if err != nil {
		return fmt.Errorf("unable to get installs: %w", err)
	}

	byts, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("unable to marshal payload to json: %w", err)
	}

	ev := &app.InstallEvent{
		OrgID:           run.OrgID,
		CreatedByID:     run.CreatedByID,
		Operation:       string(op),
		InstallID:       run.InstallID,
		OperationStatus: status,
		Payload:         byts,
	}

	res := h.db.WithContext(ctx).Create(&ev)
	if res.Error != nil {
		return fmt.Errorf("unable to create install event: %w", res.Error)
	}

	return nil
}
