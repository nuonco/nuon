package stackerrors

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const StackTemplateRenderErrorType compositeerrors.Type = "stack.template_render_failed"

type StackTemplateRenderError struct {
	Platform string `json:"platform,omitempty"`
	Detail   string `json:"detail,omitempty"`
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

const SandboxPlanRenderErrorType compositeerrors.Type = "sandbox.plan_render_failed"

type SandboxPlanRenderError struct {
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

const PlanRenderFailedTemporalType = "plan_render_failed"

func IsPlanRenderFailed(err error) bool {
	var appErr *temporal.ApplicationError
	return errors.As(err, &appErr) && appErr.Type() == PlanRenderFailedTemporalType
}
