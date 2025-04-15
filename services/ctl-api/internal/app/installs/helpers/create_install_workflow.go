package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

func (s *Helpers) CreateInstallWorkflow(ctx context.Context, installID string, workflowType app.InstallWorkflowType, metadata map[string]string, errBehavior app.StepErrorBehavior) (*app.InstallWorkflow, error) {
	installWorkflow := app.InstallWorkflow{
		Type:              workflowType,
		InstallID:         installID,
		Metadata:          generics.ToHstore(metadata),
		Status:            app.NewCompositeStatus(ctx, app.StatusPending),
		StepErrorBehavior: errBehavior,
	}

	res := s.db.WithContext(ctx).Create(&installWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create install workflow")
	}

	return &installWorkflow, nil
}
