package operationrun

import (
	"time"

	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
)

// BundleInfoFromManifest flattens a bundle's logical manifest into the
// portal-facing inventory the controller publishes at activation.
func BundleInfoFromManifest(deploymentID, digest string, manifest bundle.LogicalManifest, activatedAt time.Time) *operation.BundleInfo {
	info := &operation.BundleInfo{
		SchemaVersion: operation.SchemaVersion,
		DeploymentID:  deploymentID,
		BundleDigest:  digest,
		ActivatedAt:   activatedAt.UTC(),
		Target:        &operation.BundleTarget{OS: manifest.Target.OS, Architecture: manifest.Target.Architecture},
		Verification:  operation.BundleVerification{BlobsVerified: true, EnvelopeParsed: true},
	}
	if manifest.Release.ID != "" || manifest.Release.Digest != "" {
		info.Release = &operation.BundleReleaseIdentity{ID: manifest.Release.ID, Digest: manifest.Release.Digest}
	}
	if manifest.Package.ID != "" || manifest.Package.Digest != "" {
		info.Package = &operation.BundlePackageIdentity{
			ID: manifest.Package.ID, Digest: manifest.Package.Digest, Format: manifest.Package.Format, Target: manifest.Package.Target,
		}
	}
	add := func(kind, name, detail, digest, configDigest string, size int64) {
		info.Contents = append(info.Contents, operation.BundleContent{Kind: kind, Name: name, Detail: detail, Digest: digest, ConfigDigest: configDigest, Size: size})
		info.TotalSize += size
	}
	for _, c := range manifest.Components {
		info.Contents = append(info.Contents, operation.BundleContent{
			Kind: operation.BundleContentKindComponent, Name: c.Name, Detail: c.Type,
			Digest: c.Artifact.Digest, ConfigDigest: c.ConfigDigest, Size: c.Artifact.Size,
			ComponentDefinition: c.Definition,
		})
		info.TotalSize += c.Artifact.Size
	}
	if sandbox := manifest.Sandbox; sandbox != nil {
		add(operation.BundleContentKindSandbox, sandbox.Type, sandbox.Source.Repository, sandbox.Artifact.Digest, sandbox.ConfigDigest, sandbox.Artifact.Size)
	}
	for _, img := range manifest.Images {
		add(operation.BundleContentKindImage, img.Name, img.Repository, img.Artifact.Digest, "", img.Artifact.Size)
	}
	for _, action := range manifest.Actions {
		var size int64
		definition := &operation.BundleActionDefinition{}
		if action.Definition != nil {
			definition.TimeoutNanos = action.Definition.TimeoutNanos
			definition.Role = action.Definition.Role
			definition.BreakGlassRoleARN = action.Definition.BreakGlassRoleARN
			definition.EnableKubeConfig = action.Definition.EnableKubeConfig
			definition.KubernetesContextName = action.Definition.KubernetesContextName
			definition.ComponentDependencies = action.Definition.ComponentDependencies
			definition.References = action.Definition.References
			for _, trigger := range action.Definition.Triggers {
				definition.Triggers = append(definition.Triggers, operation.BundleActionTrigger{
					Type: trigger.Type, Index: trigger.Index, CronSchedule: trigger.CronSchedule, ComponentName: trigger.ComponentName,
				})
			}
			for _, step := range action.Definition.Steps {
				definition.Steps = append(definition.Steps, operation.BundleActionStep{
					Name: step.Name, Index: step.Index, Command: step.Command, InlineContentsDigest: step.InlineContentsDigest, Environment: step.Environment,
				})
			}
		}
		stepsByName := make(map[string]int, len(definition.Steps))
		for i := range definition.Steps {
			stepsByName[definition.Steps[i].Name] = i
		}
		for _, step := range action.Steps {
			stepIndex, found := stepsByName[step.Name]
			if !found {
				definition.Steps = append(definition.Steps, operation.BundleActionStep{
					Name: step.Name, Command: step.Command, InlineContentsDigest: step.InlineContentsDigest,
				})
				stepIndex = len(definition.Steps) - 1
			}
			bundleStep := &definition.Steps[stepIndex]
			if step.Source != nil {
				bundleStep.Source = &operation.BundleSource{
					Repository:   step.Source.Repository,
					RequestedRef: step.Source.RequestedRef,
					Commit:       step.Source.Commit,
					Directory:    step.Source.Directory,
					Version:      step.Source.Version,
					Digest:       step.Source.Digest,
				}
			}
			if step.Artifact != nil {
				size += step.Artifact.Size
				bundleStep.ArtifactDigest = step.Artifact.Digest
			}
		}
		info.Contents = append(info.Contents, operation.BundleContent{
			Kind:             operation.BundleContentKindAction,
			Name:             action.Name,
			Digest:           action.ConfigDigest,
			ConfigDigest:     action.ConfigDigest,
			Size:             size,
			ActionDefinition: definition,
		})
		info.TotalSize += size
	}
	for _, runbook := range manifest.Runbooks {
		definition := &operation.BundleRunbookDefinition{ReadmeDigest: runbook.Definition.ReadmeDigest, Steps: make([]operation.BundleRunbookStep, 0, len(runbook.Definition.Steps))}
		for _, input := range runbook.Definition.Inputs {
			definition.Inputs = append(definition.Inputs, operation.BundleRunbookInput{
				Name: input.Name, DisplayName: input.DisplayName, Description: input.Description,
				Default: input.Default, Type: input.Type, Index: input.Index, Required: input.Required, Sensitive: input.Sensitive,
			})
		}
		for _, step := range runbook.Definition.Steps {
			definition.Steps = append(definition.Steps, operation.BundleRunbookStep{
				Kind: step.Kind, Name: step.Name, Index: step.Index, Reference: step.Reference, Component: step.Component,
				Role: step.Role, PlanOnly: step.PlanOnly, DeployDependents: step.DeployDependents,
				TearDownDependents: step.TearDownDependents, SkipComponentDeploys: step.SkipComponentDeploys,
				Command: step.Command, InlineContentsDigest: step.InlineContentsDigest, Environment: step.Environment,
				TimeoutNanos: step.TimeoutNanos, TriggerName: step.TriggerName, EventTypes: step.EventTypes, FiltersDigest: step.FiltersDigest,
			})
		}
		info.Contents = append(info.Contents, operation.BundleContent{
			Kind: operation.BundleContentKindRunbook, Name: runbook.Name,
			Digest: runbook.ConfigDigest, ConfigDigest: runbook.ConfigDigest, RunbookDefinition: definition,
		})
	}
	for _, asset := range manifest.StackAssets {
		add(operation.BundleContentKindStackAsset, asset.Role, asset.SourceURL, asset.Digest, "", asset.Size)
	}
	if runner := manifest.Runner; runner != nil {
		if runner.Binary != nil {
			add(operation.BundleContentKindRunnerBinary, "runner", runner.Version, runner.Binary.Digest, "", runner.Binary.Size)
		}
		if runner.Image != nil {
			add(operation.BundleContentKindRunnerImage, runner.Image.Name, runner.Image.Repository, runner.Image.Artifact.Digest, "", runner.Image.Artifact.Size)
		}
	}
	return info
}
