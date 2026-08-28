package customerbundle

import (
	"encoding/json"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

const (
	CurrentSchemaVersion     = 3
	LogicalManifestMediaType = "application/vnd.nuon.customermanaged.manifest.v1+json"
	ProvenanceMediaType      = "application/vnd.nuon.customermanaged.provenance.v1+json"
	QualificationMediaType   = "application/vnd.nuon.customermanaged.qualification.v1+json"
	PlanEnvelopeMediaType    = "application/vnd.nuon.customermanaged.plan.v1+json"
	SourceArchiveMediaType   = "application/vnd.nuon.customermanaged.source-archive.v1+json"
	BundleArtifactType       = "application/vnd.nuon.customermanaged.bundle.v1"
	RunnerBinaryArtifactType = "application/vnd.nuon.customermanaged.runner-binary.v1"
	RunnerBinaryMediaType    = "application/octet-stream"
)

type LogicalManifest struct {
	SchemaVersion int             `json:"schema_version"`
	Release       ReleaseIdentity `json:"release"`
	Package       PackageIdentity `json:"package"`
	Target        Target          `json:"target"`
	Components    []Component     `json:"components,omitempty"`
	Sandbox       *Sandbox        `json:"sandbox,omitempty"`
	Images        []Image         `json:"images,omitempty"`
	Actions       []Action        `json:"actions,omitempty"`
	Runbooks      []Runbook       `json:"runbooks,omitempty"`
	StackAssets   []StackAsset    `json:"stack_assets,omitempty"`
	Runner        *Runner         `json:"runner,omitempty"`
}

type ReleaseIdentity struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

type PackageIdentity struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
	Format string `json:"format"`
	Target string `json:"target"`
}

// Runner packages the runner itself so an offline install never phones home:
// a standalone binary matching the bundle target, and optionally the runner
// container image for hosts that run it under a container runtime.
type Runner struct {
	Version   string    `json:"version,omitempty"`
	SourceURL string    `json:"source_url,omitempty"`
	Binary    *Artifact `json:"binary,omitempty"`
	Image     *Image    `json:"image,omitempty"`
}

type Target struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

type Component struct {
	Name         string              `json:"name"`
	Type         string              `json:"type"`
	ConfigDigest string              `json:"config_digest"`
	Definition   ComponentDefinition `json:"definition,omitempty"`
	Source       Source              `json:"source"`
	Artifact     Artifact            `json:"artifact"`
}

type ComponentDefinition map[string]any

type Sandbox struct {
	Type         string   `json:"type"`
	ConfigDigest string   `json:"config_digest"`
	Source       Source   `json:"source"`
	Artifact     Artifact `json:"artifact"`
}

type Image struct {
	Name       string   `json:"name"`
	Repository string   `json:"repository"`
	Artifact   Artifact `json:"artifact"`
}

type Source struct {
	Repository   string `json:"repository,omitempty"`
	RequestedRef string `json:"requested_ref,omitempty"`
	Commit       string `json:"commit,omitempty"`
	Directory    string `json:"directory,omitempty"`
	Version      string `json:"version,omitempty"`
	Digest       string `json:"digest,omitempty"`
}

type Artifact struct {
	MediaType            string `json:"media_type"`
	Digest               string `json:"digest"`
	Size                 int64  `json:"size"`
	PlatformOS           string `json:"platform_os,omitempty"`
	PlatformArchitecture string `json:"platform_architecture,omitempty"`
}

type Action struct {
	Name         string            `json:"name"`
	ConfigDigest string            `json:"config_digest"`
	Definition   *ActionDefinition `json:"definition,omitempty"`
	Steps        []Step            `json:"steps,omitempty"`
}

type ActionDefinition struct {
	TimeoutNanos          int64                     `json:"timeout_nanos,omitempty"`
	Role                  string                    `json:"role,omitempty"`
	BreakGlassRoleARN     string                    `json:"break_glass_role_arn,omitempty"`
	EnableKubeConfig      bool                      `json:"enable_kube_config,omitempty"`
	KubernetesContextName string                    `json:"kubernetes_context_name,omitempty"`
	ComponentDependencies []string                  `json:"component_dependencies,omitempty"`
	References            []string                  `json:"references,omitempty"`
	Triggers              []ActionTriggerDefinition `json:"triggers,omitempty"`
	Steps                 []ActionStepDefinition    `json:"steps,omitempty"`
}

type ActionTriggerDefinition struct {
	Type          string `json:"type"`
	Index         int    `json:"index,omitempty"`
	CronSchedule  string `json:"cron_schedule,omitempty"`
	ComponentName string `json:"component_name,omitempty"`
}

type ActionStepDefinition struct {
	Name                 string            `json:"name"`
	Index                int               `json:"index,omitempty"`
	Command              string            `json:"command,omitempty"`
	InlineContentsDigest string            `json:"inline_contents_digest,omitempty"`
	Environment          map[string]string `json:"environment,omitempty"`
}

type Runbook struct {
	Name         string            `json:"name"`
	ConfigDigest string            `json:"config_digest"`
	Definition   RunbookDefinition `json:"definition"`
}

type RunbookDefinition struct {
	ReadmeDigest string                   `json:"readme_digest,omitempty"`
	Inputs       []RunbookInputDefinition `json:"inputs,omitempty"`
	Steps        []RunbookStepDefinition  `json:"steps"`
}

type RunbookStepDefinition struct {
	Kind                 string            `json:"kind"`
	Name                 string            `json:"name,omitempty"`
	Index                int               `json:"index,omitempty"`
	Reference            string            `json:"reference,omitempty"`
	Component            string            `json:"component,omitempty"`
	Role                 string            `json:"role,omitempty"`
	PlanOnly             bool              `json:"plan_only,omitempty"`
	DeployDependents     bool              `json:"deploy_dependents,omitempty"`
	TearDownDependents   bool              `json:"tear_down_dependents,omitempty"`
	SkipComponentDeploys bool              `json:"skip_component_deploys,omitempty"`
	Command              string            `json:"command,omitempty"`
	InlineContentsDigest string            `json:"inline_contents_digest,omitempty"`
	Environment          map[string]string `json:"environment,omitempty"`
	TimeoutNanos         int64             `json:"timeout_nanos,omitempty"`
	TriggerName          string            `json:"trigger_name,omitempty"`
	EventTypes           []string          `json:"event_types,omitempty"`
	FiltersDigest        string            `json:"filters_digest,omitempty"`
}

type RunbookInputDefinition struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
	Type        string `json:"type,omitempty"`
	Index       int    `json:"index,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
}

type Step struct {
	Name                 string    `json:"name"`
	Command              string    `json:"command,omitempty"`
	InlineContentsDigest string    `json:"inline_contents_digest"`
	Source               *Source   `json:"source,omitempty"`
	Artifact             *Artifact `json:"artifact,omitempty"`
}

type StackAsset struct {
	Role      string `json:"role"`
	SourceURL string `json:"source_url"`
	Digest    string `json:"digest"`
	MediaType string `json:"media_type"`
	Size      int64  `json:"size"`
}

type Documents struct {
	Provenance          json.RawMessage
	QualificationReport json.RawMessage
	PlanEnvelope        json.RawMessage
	SourceArchive       json.RawMessage
}

type Root struct {
	Descriptor ocispec.Descriptor
	Source     oras.ReadOnlyTarget
}
