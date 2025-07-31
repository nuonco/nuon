package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateInstallConfigParams struct {
	ApprovalOption app.InstallApprovalOption `json:"approval_option"`
}

func (h *Helpers) CreateInstallConfig(ctx context.Context, installID string, req *CreateInstallConfigParams) (*app.InstallConfig, error) {
	installConfig := &app.InstallConfig{
		InstallID:      installID,
		ApprovalOption: req.ApprovalOption,
	}

	if err := h.db.WithContext(ctx).Create(installConfig).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config: %w", err)
	}
	return installConfig, nil
}
