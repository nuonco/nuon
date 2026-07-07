package activities

import (
	"context"
	"encoding/json"
	"fmt"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
)

type DiffIntermediateConfigsInput struct {
	AppID                     string `json:"app_id" validate:"required"`
	NewIntermediateConfigJSON string `json:"new_intermediate_config_json" validate:"required"`
	OldAppConfigID            string `json:"old_app_config_id"`
}

type DiffIntermediateConfigsOutput struct {
	Changed bool                        `json:"changed"`
	Diff    *ComputeAppConfigDiffOutput `json:"diff,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) DiffIntermediateConfigs(ctx context.Context, input *DiffIntermediateConfigsInput) (*DiffIntermediateConfigsOutput, error) {
	var newCfg pkgconfig.AppConfig
	if err := json.Unmarshal([]byte(input.NewIntermediateConfigJSON), &newCfg); err != nil {
		return nil, fmt.Errorf("unable to parse new intermediate config: %w", err)
	}

	var oldCfg *pkgconfig.AppConfig
	if input.OldAppConfigID != "" {
		loaded, err := a.loadIntermediateConfig(ctx, input.AppID, input.OldAppConfigID)
		if err == nil {
			oldCfg = loaded
		}
	}

	d := newCfg.Diff(oldCfg)
	summary := d.Summary()

	if !summary.HasChanged {
		return &DiffIntermediateConfigsOutput{Changed: false}, nil
	}

	output := &DiffIntermediateConfigsOutput{
		Changed: true,
		Diff: &ComputeAppConfigDiffOutput{
			ConfigFile: "nuon.toml",
			Additions:  summary.Added,
			Removals:   summary.Removed,
			Changed:    summary.Changed,
		},
	}

	if d.Children != nil {
		for _, child := range d.Children {
			section := diffNodeToSection(child)
			if section != nil && len(section.Entries) > 0 {
				output.Diff.Sections = append(output.Diff.Sections, *section)
			}
		}
	}

	return output, nil
}
