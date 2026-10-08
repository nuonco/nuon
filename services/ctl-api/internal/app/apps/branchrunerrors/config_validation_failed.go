package branchrunerrors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
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
		compositeerrors.MarkdownSection("Why", "The branch run stopped because the app configuration could not be parsed or validated."),
	}
	if e.Detail != "" {
		sections = append(sections, compositeerrors.CodeSection("Validation errors", e.Detail))
	}
	return append(sections, compositeerrors.MarkdownSection("How to fix", "Fix the invalid app configuration, commit the changes, and run the branch again."))
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

func UserDetail(err error) string {
	if err == nil {
		return ""
	}
	return stripConfigFailurePrefixes(userDetail(err))
}

func userDetail(err error) string {
	var parseErr parse.ParseErr
	if errors.As(err, &parseErr) {
		return parseErr.Error()
	}
	var configErr config.ErrConfig
	if errors.As(err, &configErr) {
		if configErr.Description != "" {
			return configErr.Description
		}
		return configErr.Error()
	}
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return formatValidatorErrors(validationErrs)
	}
	return err.Error()
}

func formatValidatorErrors(errs validator.ValidationErrors) string {
	lines := make([]string, 0, len(errs))
	for _, fieldErr := range errs {
		name := fieldErr.Namespace()
		if name == "" {
			name = fieldErr.Field()
		}
		lines = append(lines, fmt.Sprintf("%s: %s", name, validatorTagMessage(fieldErr)))
	}
	return strings.Join(lines, "\n")
}

func validatorTagMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "is required"
	default:
		return fmt.Sprintf("failed %s validation", fieldErr.Tag())
	}
}

func stripConfigFailurePrefixes(msg string) string {
	prefixes := []string{
		"unable to parse directory: ",
		"unable to convert to app config: ",
		"unable to parse config from repo: ",
	}
	trimmed := true
	for trimmed {
		trimmed = false
		for _, prefix := range prefixes {
			if strings.HasPrefix(msg, prefix) {
				msg = strings.TrimPrefix(msg, prefix)
				trimmed = true
			}
		}
	}
	return strings.TrimSpace(msg)
}
