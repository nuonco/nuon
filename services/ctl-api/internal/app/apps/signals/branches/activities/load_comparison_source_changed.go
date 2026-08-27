package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type LoadComparisonSourceChangedInput struct {
	RunID string `json:"run_id" validate:"required"`
}

type LoadComparisonSourceChangedOutput struct {
	ByComponentName map[string]bool `json:"by_component_name"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) LoadComparisonSourceChanged(ctx context.Context, input *LoadComparisonSourceChangedInput) (*LoadComparisonSourceChangedOutput, error) {
	if err := a.v.Struct(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	out := &LoadComparisonSourceChangedOutput{
		ByComponentName: map[string]bool{},
	}

	var comparison app.AppBranchRunComparison
	res := a.db.WithContext(ctx).
		Where(app.AppBranchRunComparison{HeadRunID: input.RunID}).
		First(&comparison)
	if res.Error != nil {
		return out, nil
	}
	if comparison.ConfigDiff == nil || !comparison.ConfigDiff.IsSet() {
		return out, nil
	}

	blobCtx := blobstore.WithBlobService(ctx, a.blobSvc)
	raw, err := comparison.ConfigDiff.Get(blobCtx)
	if err != nil || raw == "" {
		return out, nil
	}

	var configDiff ConfigDiffWithSourceOutput
	if err := json.Unmarshal([]byte(raw), &configDiff); err != nil {
		return out, nil
	}

	for _, sec := range configDiff.Sections {
		if sec.Name != "Components" {
			continue
		}
		for _, entry := range sec.Entries {
			if entry.Name == "" {
				continue
			}
			out.ByComponentName[entry.Name] = entry.SourceChanged
		}
	}

	return out, nil
}
