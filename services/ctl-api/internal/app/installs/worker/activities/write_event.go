package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type WriteEventRequest struct {
	DeployID     string `validate:"required"`
	InstallID    string `validate:"required"`
	SandboxRunID string `validate:"required"`

	Operation       eventloop.SignalType `validate:"required"`
	OperationStatus app.OperationStatus  `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) WriteEvent(ctx context.Context, req WriteEventRequest) error {
	if req.DeployID != "" {
		return a.helpers.WriteDeployEvent(ctx, req.DeployID, req.Operation, req.OperationStatus)
	}

	if req.InstallID != "" {
		return a.helpers.WriteInstallEvent(ctx, req.InstallID, req.Operation, req.OperationStatus)
	}

	if req.SandboxRunID != "" {
		return a.helpers.WriteRunEvent(ctx, req.SandboxRunID, req.Operation, req.OperationStatus)
	}

	return nil
}
