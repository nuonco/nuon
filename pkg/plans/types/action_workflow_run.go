package plantypes

import (
	"time"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

type ActionWorkflowRunPlan struct {
	ID        string `json:"id"`
	InstallID string `json:"install_id"`

	Attrs map[string]string `json:"attrs"`

	Steps           []*ActionWorkflowRunStepPlan `json:"steps"`
	BuiltinEnvVars  map[string]string            `json:"builtin_env_vars"`
	OverrideEnvVars map[string]string            `json:"override_env_vars"`
	Timeout         time.Duration                `json:"timeout,omitempty" swaggertype:"primitive,integer"`

	// Optional fields based on the configuration.
	//
	// NOTE: keep this comment detached from the field below (blank line above
	// the field) — a doc comment on a $ref field makes go-swagger generate the
	// SDK model field as an inline struct VALUE, so a null cluster_info
	// round-trips to {} in the runner and becomes a non-nil empty ClusterInfo.

	ClusterInfo *kube.ClusterInfo        `json:"cluster_info,block"`
	AWSAuth     *awscredentials.Config   `json:"aws_auth,omitempty"`
	AzureAuth   *azurecredentials.Config `json:"azure_auth,omitempty"`
	GCPAuth     *gcpcredentials.Config   `json:"gcp_auth,omitempty"`

	// Image-backed actions: SourceImage is the rendered app-authored ref
	// (e.g. ghcr.io/acme/tools:v1); ImageRegistry/ImageTag point at the
	// install-registry mirror the runner pulls from.
	SourceImage   string                         `json:"source_image,omitempty"`
	ImageRegistry *configs.OCIRegistryRepository `json:"image_registry,omitempty"`
	ImageTag      string                         `json:"image_tag,omitempty"`
	// ImageDigestRef is the digest-pinned pull reference resolved by the mirror
	// job (<login_server>/<repository>@sha256:...). When set, the runner pulls
	// this exact manifest instead of the mutable tag, binding execution to the
	// content that was mirrored.
	ImageDigestRef string `json:"image_digest_ref,omitempty"`

	MinSandboxMode
}

type ActionWorkflowRunStepPlan struct {
	ID string `json:"run_id"`

	Attrs                      map[string]string `json:"attrs"`
	InterpolatedEnvVars        map[string]string `json:"interpolated_env_vars"`
	GitSource                  *GitSource        `json:"git_source"`
	InterpolatedInlineContents string            `json:"interpolated_inline_contents"`
	InterpolatedCommand        string            `json:"interpolated_command"`
}
