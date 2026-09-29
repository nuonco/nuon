package app

import (
	"github.com/nuonco/nuon/pkg/config"
)

func ApplyInstallStackOverrides(install *Install, stackCfg *AppStackConfig) {
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

	overrides := make(map[string]config.CustomNestedStack, len(ic.CustomNestedStacks))
	for _, s := range ic.CustomNestedStacks {
		overrides[s.Name] = s
	}

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

	for _, s := range ic.CustomNestedStacks {
		if !seen[s.Name] {
			result = append(result, s)
		}
	}

	stackCfg.CustomNestedStacks = result
}
