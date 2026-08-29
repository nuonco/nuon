package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type PreviewBaselineAppConfig struct {
	AppConfigID string
	BaseRunID   string
}

func (h *Helpers) ResolvePreviewBaselineAppConfig(ctx context.Context, runID, appBranchID string) (*PreviewBaselineAppConfig, error) {
	out := &PreviewBaselineAppConfig{}

	var comparison app.AppBranchRunComparison
	err := h.db.WithContext(ctx).
		Preload("BaseRun").
		Where(app.AppBranchRunComparison{HeadRunID: runID}).
		First(&comparison).Error
	if err == nil && comparison.BaseRunID != nil && *comparison.BaseRunID != "" {
		out.BaseRunID = *comparison.BaseRunID
		if comparison.BaseRun != nil && comparison.BaseRun.AppConfigID != "" {
			out.AppConfigID = comparison.BaseRun.AppConfigID
			return out, nil
		}

		var baseRun app.AppBranchRun
		if loadErr := h.db.WithContext(ctx).First(&baseRun, "id = ?", *comparison.BaseRunID).Error; loadErr == nil && baseRun.AppConfigID != "" {
			out.AppConfigID = baseRun.AppConfigID
			return out, nil
		}
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to load comparison: %w", err)
	}

	baseRun, findErr := h.FindBaseAppBranchRun(ctx, appBranchID)
	if findErr != nil {
		if findErr == gorm.ErrRecordNotFound {
			return out, nil
		}
		return nil, fmt.Errorf("unable to find base app branch run: %w", findErr)
	}

	out.BaseRunID = baseRun.ID
	out.AppConfigID = baseRun.AppConfigID
	return out, nil
}
