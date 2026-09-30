package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/installgrouperrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type SetAppBranchRunInstallUpdateFailedErrorRequest struct {
	RunID       string `json:"run_id" validate:"required"`
	InstallID   string `json:"install_id" validate:"required"`
	InstallName string `json:"install_name"`
	WorkflowID  string `json:"workflow_id"`
	Detail      string `json:"detail"`
}

// @temporal-gen-v2 activity
// @max-retries 3
func (a *Activities) SetAppBranchRunInstallUpdateFailedError(ctx context.Context, req SetAppBranchRunInstallUpdateFailedErrorRequest) (*compositeerrors.CompositeErrorData, error) {
	data, err := compositeerrors.New(
		&installgrouperrors.InstallUpdateFailedError{
			InstallID:   req.InstallID,
			InstallName: req.InstallName,
			WorkflowID:  req.WorkflowID,
			Detail:      req.Detail,
		},
		compositeerrors.WithSource("installs", req.InstallID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build install update composite error: %w", err)
	}

	if err := a.setAppBranchRunCompositeError(ctx, req.RunID, data); err != nil {
		return nil, err
	}
	return data, nil
}
