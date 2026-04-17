package sandboxctl

import "time"

func AllJobTypes() []string {
	return []string{
		"terraform-deploy",
		"helm-chart-deploy",
		"kubernetes-manifest-deploy",
		"job-deploy",
		"noop-deploy",

		"docker-build",
		"container-image-build",
		"terraform-module-build",
		"helm-chart-build",
		"kubernetes-manifest-build",
		"noop-build",

		"oci-sync",
		"noop-sync",
		"fetch-image-metadata",

		"actions-workflow",

		"sandbox-terraform",
		"sandbox-terraform-plan",
		"sandbox-sync-secrets",

		"runner-helm",
		"runner-terraform",
	}
}

type JobCategory string

const (
	CategoryDeploy  JobCategory = "deploy"
	CategoryBuild   JobCategory = "build"
	CategoryActions JobCategory = "actions"
	CategorySandbox JobCategory = "sandbox"
	CategorySync    JobCategory = "sync"
	CategoryRunner  JobCategory = "runner"
)

var jobTypeCategories = map[string]JobCategory{
	"terraform-deploy":           CategoryDeploy,
	"helm-chart-deploy":          CategoryDeploy,
	"kubernetes-manifest-deploy": CategoryDeploy,
	"job-deploy":                 CategoryDeploy,
	"noop-deploy":                CategoryDeploy,

	"docker-build":              CategoryBuild,
	"container-image-build":     CategoryBuild,
	"terraform-module-build":    CategoryBuild,
	"helm-chart-build":          CategoryBuild,
	"kubernetes-manifest-build": CategoryBuild,
	"noop-build":                CategoryBuild,

	"oci-sync":             CategorySync,
	"noop-sync":            CategorySync,
	"fetch-image-metadata": CategorySync,

	"actions-workflow": CategoryActions,

	"sandbox-terraform":      CategorySandbox,
	"sandbox-terraform-plan": CategorySandbox,
	"sandbox-sync-secrets":   CategorySandbox,

	"runner-helm":      CategoryRunner,
	"runner-terraform": CategoryRunner,
}

func CategoryForJobType(jobType string) JobCategory {
	if cat, ok := jobTypeCategories[jobType]; ok {
		return cat
	}
	return CategoryDeploy
}

type PresetInfo struct {
	Name        ResponsePreset `json:"name"`
	Description string         `json:"description"`
}

func PresetsForCategory(cat JobCategory) []PresetInfo {
	switch cat {
	case CategoryDeploy:
		return []PresetInfo{
			{PresetDefault, "Default behavior (configured duration, no faults)"},
			{PresetSuccessSlow, "Success after 30s"},
			{PresetSuccessFast, "Success after 1s"},
			{PresetFailure, "Fail immediately with error"},
			{PresetFailureTimeout, "Fail after 60s (simulate timeout)"},
		}
	case CategoryBuild:
		return []PresetInfo{
			{PresetDefault, "Default behavior"},
			{PresetSuccessSlow, "Success after 30s"},
			{PresetSuccessFast, "Success after 1s"},
			{PresetFailure, "Fail immediately with error"},
		}
	case CategoryActions:
		return []PresetInfo{
			{PresetDefault, "Default behavior"},
			{PresetFailure, "Fail all steps"},
			{PresetPartialFailure, "Fail on last step"},
		}
	case CategorySandbox:
		return []PresetInfo{
			{PresetDefault, "Default behavior"},
			{PresetFailure, "Fail with error"},
		}
	case CategorySync:
		return []PresetInfo{
			{PresetDefault, "Default behavior"},
			{PresetSuccessFast, "Success after 1s"},
			{PresetFailure, "Fail with error"},
		}
	case CategoryRunner:
		return []PresetInfo{
			{PresetDefault, "Default behavior"},
			{PresetFailure, "Fail with error"},
		}
	default:
		return []PresetInfo{
			{PresetDefault, "Default behavior"},
			{PresetFailure, "Fail with error"},
		}
	}
}

func PresetsForJobType(jobType string) []PresetInfo {
	return PresetsForCategory(CategoryForJobType(jobType))
}

func ApplyPreset(preset ResponsePreset, defaultDuration time.Duration) *JobTypeConfig {
	switch preset {
	case PresetSuccessSlow:
		return &JobTypeConfig{
			Preset:   preset,
			Duration: 30 * time.Second,
		}
	case PresetSuccessFast:
		return &JobTypeConfig{
			Preset:   preset,
			Duration: 1 * time.Second,
		}
	case PresetFailure:
		return &JobTypeConfig{
			Preset:       preset,
			Duration:     1 * time.Second,
			ErrorMessage: "sandbox: injected failure",
		}
	case PresetFailureTimeout:
		return &JobTypeConfig{
			Preset:       preset,
			Duration:     60 * time.Second,
			ErrorMessage: "sandbox: simulated timeout",
		}
	case PresetFailurePanic:
		return &JobTypeConfig{
			Preset:       preset,
			Duration:     0,
			ErrorMessage: "sandbox: panic requested via preset",
		}
	case PresetPartialFailure:
		return &JobTypeConfig{
			Preset:       preset,
			Duration:     defaultDuration,
			ErrorMessage: "sandbox: partial failure on last step",
		}
	default:
		return &JobTypeConfig{
			Preset:   PresetDefault,
			Duration: defaultDuration,
		}
	}
}
