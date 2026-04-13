package stack

import (
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// applyInstallStackOverrides applies per-install stack template overrides
// from the InstallConfig onto the app-level AppStackConfig.
// Install-level overrides take precedence over app-level defaults.
// For CustomNestedStacks, entries with the same Name replace the app-level
// entry; entries with new Names are appended. Original ordering is preserved.
func ApplyInstallStackOverrides(install *app.Install, stackCfg *app.AppStackConfig) {
	if install.InstallConfig == nil {
		return
	}
	ic := install.InstallConfig

	if ic.VPCNestedTemplateURL != nil && *ic.VPCNestedTemplateURL != "" {
		stackCfg.VPCNestedTemplateURL = *ic.VPCNestedTemplateURL
	}
	if ic.RunnerNestedTemplateURL != nil && *ic.RunnerNestedTemplateURL != "" {
		stackCfg.RunnerNestedTemplateURL = *ic.RunnerNestedTemplateURL
	}

	if len(ic.CustomNestedStacks) == 0 {
		return
	}

	// Build a lookup of install-level overrides by name.
	overrides := make(map[string]config.CustomNestedStack, len(ic.CustomNestedStacks))
	for _, s := range ic.CustomNestedStacks {
		overrides[s.Name] = s
	}

	// Walk the app-level stacks in order, replacing any that have an install-level override.
	seen := make(map[string]bool, len(stackCfg.CustomNestedStacks)+len(ic.CustomNestedStacks))
	result := make([]config.CustomNestedStack, 0, len(stackCfg.CustomNestedStacks)+len(ic.CustomNestedStacks))
	for _, s := range stackCfg.CustomNestedStacks {
		if override, ok := overrides[s.Name]; ok {
			result = append(result, override)
		} else {
			result = append(result, s)
		}
		seen[s.Name] = true
	}

	// Append any install-level stacks that are new (not overriding an existing one),
	// preserving their original order from the install config.
	for _, s := range ic.CustomNestedStacks {
		if !seen[s.Name] {
			result = append(result, s)
		}
	}

	stackCfg.CustomNestedStacks = result
}
