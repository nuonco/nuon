package activities

import (
	"context"
	"encoding/json"
	"fmt"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ComputeAppConfigDiffInput struct {
	AppID       string `json:"app_id" validate:"required"`
	NewConfigID string `json:"new_config_id" validate:"required"`
	OldConfigID string `json:"old_config_id"`
}

type ConfigDiffSection struct {
	Name      string            `json:"name"`
	Additions int               `json:"additions"`
	Removals  int               `json:"removals"`
	Changed   int               `json:"changed"`
	Entries   []ConfigDiffEntry `json:"entries"`
}

type ConfigDiffEntry struct {
	Op          string `json:"op"`          // "add", "remove", "change"
	Name        string `json:"name"`        // primary identifier (component name, env var key, etc.)
	Description string `json:"description"` // secondary info (type, value, path, etc.)
}

type ComputeAppConfigDiffOutput struct {
	ConfigFile string              `json:"config_file"`
	Additions  int                 `json:"additions"`
	Removals   int                 `json:"removals"`
	Changed    int                 `json:"changed"`
	Sections   []ConfigDiffSection `json:"sections"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ComputeAppConfigDiff(ctx context.Context, input *ComputeAppConfigDiffInput) (*ComputeAppConfigDiffOutput, error) {
	newCfg, err := a.loadIntermediateConfig(ctx, input.AppID, input.NewConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to load new config: %w", err)
	}

	var oldCfg *pkgconfig.AppConfig
	if input.OldConfigID != "" {
		oldCfg, err = a.loadIntermediateConfig(ctx, input.AppID, input.OldConfigID)
		if err != nil {
			// Non-fatal: treat as first config (everything is "added")
			oldCfg = nil
		}
	}

	d := newCfg.Diff(oldCfg)

	output := &ComputeAppConfigDiffOutput{
		ConfigFile: "nuon.toml",
	}

	if d.Children != nil {
		for _, child := range d.Children {
			section := diffNodeToSection(child)
			if section != nil && (section.Additions > 0 || section.Removals > 0 || section.Changed > 0) {
				output.Sections = append(output.Sections, *section)
				output.Additions += section.Additions
				output.Removals += section.Removals
				output.Changed += section.Changed
			}
		}
	}

	return output, nil
}

// diffNodeToSection converts a top-level diff node (like "components", "actions", etc.)
// into a flat ConfigDiffSection for the UI. Counts are at the entity level:
// grouped sections count per-entity, ungrouped sections count as a single entity.
func diffNodeToSection(node *diff.Diff) *ConfigDiffSection {
	if node == nil {
		return nil
	}

	sectionName, grouped := sectionDisplayNameAndGrouped(node.Key)
	if sectionName == "" {
		return nil
	}

	section := &ConfigDiffSection{
		Name: sectionName,
	}

	if grouped {
		for _, entityNode := range node.Children {
			op := entityAggregateOp(entityNode)
			if op == "" {
				continue
			}

			entry := ConfigDiffEntry{
				Op:   string(op),
				Name: entityNode.Key,
			}
			section.Entries = append(section.Entries, entry)

			switch op {
			case diff.OpAdd:
				section.Additions++
			case diff.OpRemove:
				section.Removals++
			case diff.OpChange:
				section.Changed++
			}
		}
	} else {
		op := entityAggregateOp(node)
		if op != "" {
			section.Entries = append(section.Entries, ConfigDiffEntry{
				Op:   string(op),
				Name: node.Key,
			})
			switch op {
			case diff.OpAdd:
				section.Additions = 1
			case diff.OpRemove:
				section.Removals = 1
			case diff.OpChange:
				section.Changed = 1
			}
		}
	}

	return section
}

// entityAggregateOp determines the overall operation for a diff subtree.
// If there are adds but no removes → add (changes from zero-value defaults
// like false→true are treated as part of the add).
// If there are removes but no adds → remove.
// Otherwise → change.
// Returns "" if no changes exist.
func entityAggregateOp(node *diff.Diff) diff.Op {
	if node == nil {
		return ""
	}

	hasAdd := false
	hasRemove := false
	hasChange := false

	var walk func(n *diff.Diff)
	walk = func(n *diff.Diff) {
		if n.Diff != nil && n.Diff.Op != diff.OpNoop && n.Diff.Op != diff.OpUnknown {
			switch n.Diff.Op {
			case diff.OpAdd:
				hasAdd = true
			case diff.OpRemove:
				hasRemove = true
			case diff.OpChange:
				hasChange = true
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(node)

	if !hasAdd && !hasRemove && !hasChange {
		return ""
	}
	if hasAdd && !hasRemove {
		return diff.OpAdd
	}
	if hasRemove && !hasAdd {
		return diff.OpRemove
	}
	return diff.OpChange
}

// sectionDisplayNameAndGrouped maps diff tree keys to UI section names and
// whether the section contains multiple named entities (grouped) or is a
// single logical entity (ungrouped).
func sectionDisplayNameAndGrouped(key string) (string, bool) {
	switch key {
	case "components":
		return "Components", true
	case "actions":
		return "Actions", true
	case "inputs":
		return "Install inputs", true
	case "secrets":
		return "Secrets", true
	case "policies":
		return "Policies", true
	case "sandbox":
		return "Sandbox", false
	case "runner":
		return "Runner", false
	case "permissions":
		return "Permissions", false
	case "stack":
		return "Stack", false
	case "break_glass":
		return "Break glass", false
	case "operation_roles":
		return "Operation roles", false
	default:
		return "", false
	}
}

func (a *Activities) loadIntermediateConfig(ctx context.Context, appID, configID string) (*pkgconfig.AppConfig, error) {
	var appCfg app.AppConfig
	res := a.db.WithContext(ctx).
		Where(app.AppConfig{AppID: appID}).
		First(&appCfg, "id = ?", configID)
	if res.Error != nil {
		return nil, fmt.Errorf("config not found: %w", res.Error)
	}

	if appCfg.IntermediateConfig == nil {
		return nil, fmt.Errorf("config %s has no intermediate config", configID)
	}

	intermediateJSON, err := appCfg.IntermediateConfig.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load intermediate config: %w", err)
	}

	var cfg pkgconfig.AppConfig
	if err := json.Unmarshal([]byte(intermediateJSON), &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse intermediate config: %w", err)
	}

	return &cfg, nil
}
