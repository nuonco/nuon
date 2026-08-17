package actionerrors

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"

const PreparationFailedErrorType compositeerrors.Type = "action.preparation_failed"

type PreparationFailedError struct {
	Detail string `json:"detail"`
}

var _ compositeerrors.CompositeError = (*PreparationFailedError)(nil)

func (*PreparationFailedError) Error() string {
	return "Unable to prepare action run"
}

func (*PreparationFailedError) Type() compositeerrors.Type {
	return PreparationFailedErrorType
}

func (*PreparationFailedError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

func (e *PreparationFailedError) Sections() []compositeerrors.Section {
	sections := []compositeerrors.Section{
		compositeerrors.MarkdownSection("What happened", "The action failed before it could be sent to the runner."),
	}
	if e.Detail != "" {
		sections = append(sections, compositeerrors.CodeSection("Error details", e.Detail))
	}
	return append(sections, compositeerrors.MarkdownSection("How to fix", "Resolve the reported install or action configuration issue, then retry the action."))
}
