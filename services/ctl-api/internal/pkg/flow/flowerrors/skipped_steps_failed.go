package flowerrors

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const SkippedStepsFailedType compositeerrors.Type = "workflow.skipped_steps_failed"

type FailedStep struct {
	StepID string `json:"step_id"`
	Name   string `json:"name"`
	Detail string `json:"detail,omitempty"`
}

type SkippedStepsFailedError struct {
	Steps []FailedStep `json:"steps"`
}

var _ compositeerrors.CompositeError = (*SkippedStepsFailedError)(nil)

func (e *SkippedStepsFailedError) Error() string {
	switch len(e.Steps) {
	case 0:
		return "A step failed during the workflow"
	case 1:
		return fmt.Sprintf("%s failed", e.Steps[0].Name)
	default:
		return fmt.Sprintf("%d steps failed", len(e.Steps))
	}
}

func (e *SkippedStepsFailedError) Type() compositeerrors.Type {
	return SkippedStepsFailedType
}

func (e *SkippedStepsFailedError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

func (e *SkippedStepsFailedError) Sections() []compositeerrors.Section {
	if len(e.Steps) == 0 {
		return nil
	}
	lines := make([]string, 0, len(e.Steps))
	for _, step := range e.Steps {
		if step.Detail == "" {
			lines = append(lines, step.Name)
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s", step.Name, step.Detail))
	}
	return []compositeerrors.Section{
		compositeerrors.MarkdownSection("What happened", "The workflow finished its other steps, but these steps failed and were allowed to continue."),
		compositeerrors.TextSection("Failed steps", strings.Join(lines, "\n")),
	}
}
