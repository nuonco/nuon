// Package stackerrors holds the typed CompositeError implementations produced
// during install-stack-version generation and sandbox plan rendering. These
// errors surface config mistakes that won't resolve by retrying, so they carry
// WithTerminal hints that prevent the conductor from burning auto-retries.
package stackerrors

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// StackTemplateRenderErrorType is the discriminator for a stack version that
// could not be rendered due to a template or config problem.
const StackTemplateRenderErrorType compositeerrors.Type = "stack.template_render_failed"

// StackTemplateRenderError is persisted on an InstallStackVersion row when the
// template rendering step (CloudFormation, ARM, or GCP tfvars) fails. It
// signals that the failure is caused by the app config, not transient
// infrastructure, and that retrying without a config change is pointless.
type StackTemplateRenderError struct {
	// Platform identifies the cloud target ("aws", "azure", "gcp").
	Platform string `json:"platform,omitempty"`
	// Detail is the sanitised error message from the renderer (untrusted; rendered
	// as a code block).
	Detail string `json:"detail,omitempty"`
}

var _ compositeerrors.CompositeError = (*StackTemplateRenderError)(nil)
var _ compositeerrors.HintsProvider = (*StackTemplateRenderError)(nil)

func (e *StackTemplateRenderError) Error() string {
	if e.Platform != "" {
		return fmt.Sprintf("stack template rendering failed for %s", e.Platform)
	}
	return "stack template rendering failed"
}

func (e *StackTemplateRenderError) Type() compositeerrors.Type { return StackTemplateRenderErrorType }

func (e *StackTemplateRenderError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityFatal
}

func (e *StackTemplateRenderError) Sections() []compositeerrors.Section {
	sections := []compositeerrors.Section{
		compositeerrors.MarkdownSection("Why", "The install stack template could not be rendered. This is usually caused by a misconfigured app stack config (invalid template URL, missing nested stack, or a variable that couldn't be resolved)."),
	}
	if e.Detail != "" {
		sections = append(sections, compositeerrors.CodeSection("Error detail", e.Detail))
	}
	sections = append(sections, compositeerrors.MarkdownSection("How to fix", "Review your app stack config and fix the configuration error, then re-run `nuon apps sync` to regenerate the template."))
	return sections
}

func (e *StackTemplateRenderError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithTerminal()
}

// SandboxPlanRenderErrorType is the discriminator for a sandbox run whose plan
// failed to render before any infrastructure was touched.
const SandboxPlanRenderErrorType compositeerrors.Type = "sandbox.plan_render_failed"

// SandboxPlanRenderError is persisted on an InstallSandboxRun row when the
// plan-creation step fails in a way that indicates a config or template
// problem (i.e. before or during the runner job, not as a result of live
// infrastructure). It carries a terminal hint so the conductor does not
// auto-retry a failure that requires a config change to fix.
type SandboxPlanRenderError struct {
	// Detail is the sanitised error message (untrusted; rendered as a code block).
	Detail string `json:"detail,omitempty"`
}

var _ compositeerrors.CompositeError = (*SandboxPlanRenderError)(nil)
var _ compositeerrors.HintsProvider = (*SandboxPlanRenderError)(nil)

func (e *SandboxPlanRenderError) Error() string {
	return "sandbox plan rendering failed"
}

func (e *SandboxPlanRenderError) Type() compositeerrors.Type { return SandboxPlanRenderErrorType }

func (e *SandboxPlanRenderError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityFatal
}

func (e *SandboxPlanRenderError) Sections() []compositeerrors.Section {
	sections := []compositeerrors.Section{
		compositeerrors.MarkdownSection("Why", "The sandbox plan could not be rendered. This is typically caused by a misconfigured sandbox or app config that prevents the plan from being prepared."),
	}
	if e.Detail != "" {
		sections = append(sections, compositeerrors.CodeSection("Error detail", e.Detail))
	}
	sections = append(sections, compositeerrors.MarkdownSection("How to fix", "Review your sandbox and app configuration. Fix the configuration error and retry the operation."))
	return sections
}

func (e *SandboxPlanRenderError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithTerminal()
}
