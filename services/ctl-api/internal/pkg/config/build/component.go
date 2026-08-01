package build

import (
	"fmt"

	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
)

// ComponentConnectionInput is everything that lands on
// ComponentConfigConnection rather than on the per-type config row.
type ComponentConnectionInput struct {
	ComponentID   string
	AppConfigID   string
	References    []string
	Checksum      string
	DependencyIDs []string

	BuildTimeout  string
	DeployTimeout string

	MaxAutoRetries               *int
	SkipNoops                    *bool
	Toggleable                   *bool
	DefaultEnabled               *bool
	AutoApproveOnPoliciesPassing *bool
	DriftSchedule                *string

	HealthEnabled             *bool
	HealthStabilizationWindow string
	HealthBlockDeploy         *bool
	HealthProbes              []config.ComponentHealthProbeConfig
	HealthRequiredChecks      []string

	OperationRoles        []config.EntityOperationRole
	KubernetesContextName string
}

// ComponentConnection validates and builds the shared connection; the caller
// attaches the per-type config before persisting.
func ComponentConnection(in ComponentConnectionInput) (*app.ComponentConfigConnection, error) {
	if in.BuildTimeout != "" {
		if err := validation.ValidateBuildTimeout(in.BuildTimeout); err != nil {
			return nil, err
		}
	}
	if in.DeployTimeout != "" {
		if err := validation.ValidateDeployTimeout(in.DeployTimeout); err != nil {
			return nil, err
		}
	}
	if in.MaxAutoRetries != nil {
		if err := validation.ValidateMaxAutoRetries(*in.MaxAutoRetries); err != nil {
			return nil, err
		}
	}
	if in.DriftSchedule != nil {
		if err := validation.ValidateCronSchedule(*in.DriftSchedule); err != nil {
			return nil, err
		}
	}
	if in.HealthStabilizationWindow != "" {
		if err := validation.ValidateHealthStabilizationWindow(in.HealthStabilizationWindow); err != nil {
			return nil, err
		}
	}
	if err := validation.ValidateHealthProbeList(in.HealthProbes); err != nil {
		return nil, err
	}
	if err := validation.ValidateRequiredChecks(in.HealthRequiredChecks); err != nil {
		return nil, err
	}
	if err := ValidateOperationRoles(in.OperationRoles); err != nil {
		return nil, err
	}

	ccc := &app.ComponentConfigConnection{
		ComponentID:                  in.ComponentID,
		AppConfigID:                  in.AppConfigID,
		References:                   pq.StringArray(in.References),
		Checksum:                     in.Checksum,
		ComponentDependencyIDs:       pq.StringArray(in.DependencyIDs),
		BuildTimeout:                 in.BuildTimeout,
		DeployTimeout:                in.DeployTimeout,
		MaxAutoRetries:               in.MaxAutoRetries,
		SkipNoops:                    in.SkipNoops,
		Toggleable:                   in.Toggleable,
		DefaultEnabled:               in.DefaultEnabled,
		AutoApproveOnPoliciesPassing: in.AutoApproveOnPoliciesPassing,
		HealthEnabled:                in.HealthEnabled,
		HealthStabilizationWindow:    in.HealthStabilizationWindow,
		HealthBlockDeploy:            in.HealthBlockDeploy,
		HealthProbes:                 validation.ToAppHealthProbesFromList(in.HealthProbes),
		HealthRequiredChecks:         validation.ToAppRequiredChecks(in.HealthRequiredChecks),
		OperationRoles:               OperationRoles(in.OperationRoles),
		KubernetesContextName:        in.KubernetesContextName,
	}

	if in.DriftSchedule != nil {
		ccc.DriftSchedule = *in.DriftSchedule
	}

	return ccc, nil
}

// ComponentConnectionInputFromConfig is the single place that knows which
// shared connection fields each component type exposes.
func ComponentConnectionInputFromConfig(comp *config.Component, componentID, appConfigID string, dependencyIDs []string) (ComponentConnectionInput, error) {
	in := ComponentConnectionInput{
		ComponentID:           componentID,
		AppConfigID:           appConfigID,
		Checksum:              comp.Checksum,
		DependencyIDs:         dependencyIDs,
		References:            refStrings(comp.References),
		Toggleable:            comp.Toggleable,
		DefaultEnabled:        comp.DefaultEnabled,
		OperationRoles:        comp.OperationRoles,
		KubernetesContextName: comp.KubernetesContext,
	}

	switch {
	case comp.DockerBuild != nil:
		obj := comp.DockerBuild
		in.BuildTimeout = obj.BuildTimeout
		in.DeployTimeout = obj.DeployTimeout
		in.MaxAutoRetries = obj.MaxAutoRetries
		in.SkipNoops = obj.SkipNoops
		in.AutoApproveOnPoliciesPassing = obj.AutoApproveOnPoliciesPassing
	case comp.HelmChart != nil:
		obj := comp.HelmChart
		in.BuildTimeout = obj.BuildTimeout
		in.DeployTimeout = obj.DeployTimeout
		in.MaxAutoRetries = obj.MaxAutoRetries
		in.SkipNoops = obj.SkipNoops
		in.AutoApproveOnPoliciesPassing = obj.AutoApproveOnPoliciesPassing
		in.DriftSchedule = obj.DriftSchedule
		applyHealth(&in, obj.Health)
	case comp.TerraformModule != nil:
		obj := comp.TerraformModule
		in.BuildTimeout = obj.BuildTimeout
		in.DeployTimeout = obj.DeployTimeout
		in.MaxAutoRetries = obj.MaxAutoRetries
		in.SkipNoops = obj.SkipNoops
		in.AutoApproveOnPoliciesPassing = obj.AutoApproveOnPoliciesPassing
		in.DriftSchedule = obj.DriftSchedule
	case comp.KubernetesManifest != nil:
		obj := comp.KubernetesManifest
		in.BuildTimeout = obj.BuildTimeout
		in.DeployTimeout = obj.DeployTimeout
		in.MaxAutoRetries = obj.MaxAutoRetries
		in.SkipNoops = obj.SkipNoops
		in.AutoApproveOnPoliciesPassing = obj.AutoApproveOnPoliciesPassing
		in.DriftSchedule = obj.DriftSchedule
		applyHealth(&in, obj.Health)
	case comp.Pulumi != nil:
		obj := comp.Pulumi
		in.BuildTimeout = obj.BuildTimeout
		in.DeployTimeout = obj.DeployTimeout
		in.MaxAutoRetries = obj.MaxAutoRetries
		in.SkipNoops = obj.SkipNoops
		in.AutoApproveOnPoliciesPassing = obj.AutoApproveOnPoliciesPassing
		in.DriftSchedule = obj.DriftSchedule
	case comp.ExternalImage != nil:
		in.BuildTimeout = comp.ExternalImage.BuildTimeout
		in.DeployTimeout = comp.ExternalImage.DeployTimeout
	case comp.Job != nil:
		in.BuildTimeout = comp.Job.BuildTimeout
		in.DeployTimeout = comp.Job.DeployTimeout
	default:
		return ComponentConnectionInput{}, fmt.Errorf("component %q has no type configuration", comp.Name)
	}

	return in, nil
}

func applyHealth(in *ComponentConnectionInput, health *config.ComponentHealthConfig) {
	if health == nil {
		return
	}
	in.HealthEnabled = health.Enabled
	in.HealthStabilizationWindow = health.StabilizationWindow
	in.HealthBlockDeploy = health.BlockDeploy
	in.HealthProbes = health.Probes
	in.HealthRequiredChecks = health.RequiredChecks
}

func refStrings(in []refs.Ref) []string {
	out := make([]string, 0, len(in))
	for _, ref := range in {
		out = append(out, ref.String())
	}
	return out
}
