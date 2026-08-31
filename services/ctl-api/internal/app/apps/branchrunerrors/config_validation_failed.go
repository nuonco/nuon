package branchrunerrors

import (
	"errors"

	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const (
	ConfigValidationFailedType         compositeerrors.Type = "app_branch_run.config_validation_failed"
	ConfigValidationFailedTemporalType                      = "app_branch_run_config_validation_failed"
)

type ConfigValidationFailedError struct {
	Detail string `json:"detail,omitempty"`
}

var _ compositeerrors.CompositeError = (*ConfigValidationFailedError)(nil)
var _ compositeerrors.HintsProvider = (*ConfigValidationFailedError)(nil)

func (e *ConfigValidationFailedError) Error() string {
	return "App configuration validation failed"
}

func (e *ConfigValidationFailedError) Type() compositeerrors.Type {
	return ConfigValidationFailedType
}

func (e *ConfigValidationFailedError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityFatal
}

func (e *ConfigValidationFailedError) Sections() []compositeerrors.Section {
	sections := []compositeerrors.Section{
		compositeerrors.MarkdownSection("Why", "The branch run stopped because the app configuration contains invalid template references."),
	}
	if e.Detail != "" {
		sections = append(sections, compositeerrors.CodeSection("Validation errors", e.Detail))
	}
	return append(sections, compositeerrors.MarkdownSection("How to fix", "Fix the invalid references in the app configuration, commit the changes, and run the branch again."))
}

func (e *ConfigValidationFailedError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithTerminal()
}

func ValidationDetail(err error) (string, bool) {
	var appErr *temporal.ApplicationError
	if !errors.As(err, &appErr) || appErr.Type() != ConfigValidationFailedTemporalType {
		return "", false
	}
	return appErr.Message(), true
}
