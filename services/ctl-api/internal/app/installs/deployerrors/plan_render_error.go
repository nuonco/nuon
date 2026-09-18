package deployerrors

import (
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const DeployPlanRenderErrorType compositeerrors.Type = "deploy.plan_render_failed"

type DeployPlanRenderError struct {
	ComponentName string `json:"component_name,omitempty"`
	Stage         string `json:"stage,omitempty"`
	Detail        string `json:"detail,omitempty"`
}

var _ compositeerrors.CompositeError = (*DeployPlanRenderError)(nil)
var _ compositeerrors.HintsProvider = (*DeployPlanRenderError)(nil)

func (e *DeployPlanRenderError) Error() string {
	if e.ComponentName != "" {
		return fmt.Sprintf("Unable to render the deploy config for %s", e.ComponentName)
	}
	return "Unable to render the deploy config"
}

func (e *DeployPlanRenderError) Type() compositeerrors.Type { return DeployPlanRenderErrorType }

func (e *DeployPlanRenderError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityFatal
}

func (e *DeployPlanRenderError) Sections() []compositeerrors.Section {
	name := e.ComponentName
	if name == "" {
		name = "this component"
	}

	why := fmt.Sprintf("The deploy config for %s references a value that could not be resolved, so no plan could be built. Nothing was applied to your cloud account.", name)
	if e.Stage != "" {
		why = fmt.Sprintf("The %s for %s referenced a value that could not be resolved, so no plan could be built. Nothing was applied to your cloud account.", e.Stage, name)
	}

	sections := []compositeerrors.Section{
		compositeerrors.MarkdownSection("Why", why),
	}
	if e.Detail != "" {
		sections = append(sections, compositeerrors.CodeSection("Error detail", e.Detail))
	}
	sections = append(sections, compositeerrors.MarkdownSection("How to fix", "Check the referenced value exists in your app config — a missing sandbox output, action workflow output, or input is the usual cause. Fix the reference, re-run `nuon apps sync`, then retry the deploy."))
	return sections
}

func (e *DeployPlanRenderError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithTerminal()
}

const DeployPlanRenderFailedTemporalType = "deploy_plan_render_failed"

func NewDeployPlanRenderFailed(err error, msg string) error {
	return temporal.NewNonRetryableApplicationError(msg, DeployPlanRenderFailedTemporalType, errors.Wrap(err, msg))
}

func findPlanRenderAppError(err error) *temporal.ApplicationError {
	for err != nil {
		var appErr *temporal.ApplicationError
		if !stderrors.As(err, &appErr) {
			return nil
		}
		if appErr.Type() == DeployPlanRenderFailedTemporalType {
			return appErr
		}
		err = stderrors.Unwrap(appErr)
	}
	return nil
}

func IsDeployPlanRenderFailed(err error) bool {
	return findPlanRenderAppError(err) != nil
}

func PlanRenderDetail(err error) string {
	var deepest *temporal.ApplicationError
	for err != nil {
		var appErr *temporal.ApplicationError
		if !stderrors.As(err, &appErr) {
			break
		}
		deepest = appErr
		err = stderrors.Unwrap(appErr)
	}
	if deepest == nil {
		return ""
	}
	return deepest.Message()
}

func PlanRenderStage(err error) string {
	appErr := findPlanRenderAppError(err)
	if appErr == nil {
		return ""
	}
	return strings.TrimPrefix(appErr.Message(), "unable to render ")
}
