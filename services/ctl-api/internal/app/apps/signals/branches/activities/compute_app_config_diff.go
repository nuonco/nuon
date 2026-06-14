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
	summary := d.Summary()

	output := &ComputeAppConfigDiffOutput{
		ConfigFile: "nuon.toml",
		Additions:  summary.Added,
		Removals:   summary.Removed,
		Changed:    summary.Changed,
	}

	// Walk the diff tree to extract sections we care about
	if d.Children != nil {
		for _, child := range d.Children {
			section := diffNodeToSection(child)
			if section != nil && len(section.Entries) > 0 {
				output.Sections = append(output.Sections, *section)
			}
		}
	}

	return output, nil
}

// diffNodeToSection converts a top-level diff node (like "components", "actions", etc.)
// into a flat ConfigDiffSection for the UI.
func diffNodeToSection(node *diff.Diff) *ConfigDiffSection {
	if node == nil {
		return nil
	}

	sectionName := sectionDisplayName(node.Key)
	if sectionName == "" {
		return nil
	}

	section := &ConfigDiffSection{
		Name: sectionName,
	}

	for _, child := range node.Children {
		collectEntries(child, "", section)
	}

	return section
}

// collectEntries recursively walks a diff subtree and collects leaf entries.
// itemName is the top-level item name (e.g., "ctl-api" for a component) that
// persists through the entire recursion so leaf entries are attributed to the
// right parent regardless of nesting depth.
func collectEntries(node *diff.Diff, itemName string, section *ConfigDiffSection) {
	if node == nil {
		return
	}

	if node.Diff != nil && node.Diff.Op != diff.OpNoop {
		entry := ConfigDiffEntry{
			Op:   string(node.Diff.Op),
			Name: node.Key,
		}
		if itemName != "" {
			entry.Name = itemName
			entry.Description = node.Key + ": " + node.Diff.Diff
		} else {
			entry.Description = node.Diff.Diff
		}

		switch node.Diff.Op {
		case diff.OpAdd:
			section.Additions++
		case diff.OpRemove:
			section.Removals++
		case diff.OpChange:
			section.Changed++
		}

		section.Entries = append(section.Entries, entry)
		return
	}

	if len(node.Children) == 0 {
		return
	}

	// Collect all leaf diffs from the subtree, preserving itemName through
	// the recursion so deeply nested changes (e.g., step env vars) are still
	// attributed to the top-level item (e.g., the action name).
	resolvedName := itemName
	if resolvedName == "" {
		resolvedName = node.Key
	}

	for _, c := range node.Children {
		collectEntries(c, resolvedName, section)
	}
}

// sectionDisplayName maps diff tree keys to UI section names.
func sectionDisplayName(key string) string {
	switch key {
	case "components":
		return "Components"
	case "actions":
		return "Actions"
	case "inputs":
		return "Install inputs"
	case "secrets":
		return "Secrets"
	case "sandbox":
		return "Sandbox"
	case "runner":
		return "Runner"
	case "permissions":
		return "Permissions"
	case "stack":
		return "Stack"
	default:
		return ""
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
