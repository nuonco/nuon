package helpers

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ResolveStepDispatch derives the parser-dispatch ParseContext and the
// InvocationContext from a workflow step.
//
// The returned ParseContext is built as `<producer>/<phase>/<cloud>` with
// each segment best-effort: when the producer is unknown (e.g. an opaque
// step target type) the parse context is left empty and the pipeline falls
// back to the unknown_error builtin via the catalog's ancestor walk.
//
// The returned InvocationContext is populated with whatever the step chain
// makes available — the caller's own InvocationContext wins on overlap.
//
// This method is safe to call from an activity. It executes one to two
// short DB lookups depending on the step target type.
func (h *Helpers) ResolveStepDispatch(
	ctx context.Context,
	stepID string,
) (composite_error.ParseContext, composite_error.InvocationContext, error) {
	if stepID == "" {
		return "", composite_error.InvocationContext{}, errors.New("composite_errors: step id required")
	}

	var step app.WorkflowStep
	if err := h.db.WithContext(ctx).
		Select("id", "org_id", "name", "step_target_id", "step_target_type").
		Where("id = ?", stepID).
		First(&step).Error; err != nil {
		return "", composite_error.InvocationContext{}, errors.Wrap(err, "load step")
	}

	inv := composite_error.InvocationContext{
		OrgID:     step.OrgID,
		OwnerID:   step.ID,
		OwnerType: step.TableName(),
		StepID:    step.ID,
	}

	producer, install, component := h.resolveStepTarget(ctx, &step)
	if install != nil {
		inv.InstallID = install.ID
		inv.CloudPlatform = string(install.CloudPlatform)
	}
	if component != nil {
		inv.ComponentID = component.ID
		inv.ComponentType = string(component.Type)
		if producer == "" {
			producer = producerFromComponentType(component.Type)
		}
	}

	phase := phaseFromStepName(step.Name, producer)

	return buildParseContext(producer, phase, inv.CloudPlatform), inv, nil
}

// resolveStepTarget loads the entity that the step's StepTarget points at and
// returns producer plus install / component pointers — any of which may be
// empty/nil when the lookup fails or the target type is unknown.
//
// Errors are degraded to empty results so the parser pipeline still falls
// back to the unknown_error builtin instead of failing the whole record
// path. Not-found is logged at debug; unexpected DB errors (connection,
// permission, etc.) are logged at warn so they show up in operational
// dashboards.
func (h *Helpers) resolveStepTarget(
	ctx context.Context,
	step *app.WorkflowStep,
) (producer string, install *app.Install, component *app.Component) {
	if step.StepTargetID == "" {
		return "", nil, nil
	}

	// AppRunnerConfig is preloaded with only the columns that feed
	// install.AfterFind's CloudPlatform / RunnerType derivation —
	// dispatch never needs the heavier yaml / config blob.
	runnerConfigSelect := func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "install_id", "cloud_platform", "type")
	}

	switch app.WorkflowStepTargetType(step.StepTargetType) {
	case app.WorkflowStepTargetTypeInstallDeploy,
		app.WorkflowStepTargetTypeInstallDeploys:

		var deploy app.InstallDeploy
		if err := h.db.WithContext(ctx).
			Preload("InstallComponent").
			Preload("InstallComponent.Component").
			Preload("InstallComponent.Install").
			Preload("InstallComponent.Install.AppRunnerConfig", runnerConfigSelect).
			First(&deploy, "id = ?", step.StepTargetID).Error; err != nil {
			h.logTargetLookupError("install_deploy", step, err)
			return "", nil, nil
		}
		comp := deploy.InstallComponent.Component
		ins := deploy.InstallComponent.Install
		return producerFromComponentType(comp.Type), &ins, &comp

	case app.WorkflowStepTargetTypeInstallSandboxRun,
		app.WorkflowStepTargetTypeInstallSandboxRuns:

		var run app.InstallSandboxRun
		if err := h.db.WithContext(ctx).
			Preload("Install").
			Preload("Install.AppRunnerConfig", runnerConfigSelect).
			First(&run, "id = ?", step.StepTargetID).Error; err != nil {
			h.logTargetLookupError("install_sandbox_run", step, err)
			return "", nil, nil
		}
		// Sandbox runs are always terraform-driven.
		return "terraform", &run.Install, nil
	}

	return "", nil, nil
}

// logTargetLookupError writes the lookup failure at the right level: debug
// for not-found (expected when the target is gone), warn for everything
// else (worth investigating).
func (h *Helpers) logTargetLookupError(targetKind string, step *app.WorkflowStep, err error) {
	fields := []zap.Field{
		zap.String("target_kind", targetKind),
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.Error(err),
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.l.Debug("composite_errors: step target not found, degrading dispatch context", fields...)
		return
	}
	h.l.Warn("composite_errors: step target lookup failed, degrading dispatch context", fields...)
}

// producerFromComponentType maps a Component.Type to the parse-context
// producer segment. Returns "" when no producer is known.
func producerFromComponentType(t app.ComponentType) string {
	switch t {
	case app.ComponentTypeTerraformModule:
		return "terraform"
	case app.ComponentTypeHelmChart:
		return "helm"
	case app.ComponentTypeKubernetesManifest:
		return "kubernetes"
	case app.ComponentTypePulumi:
		return "pulumi"
	}
	return ""
}

// phaseFromStepName picks a phase segment for the parse context from the
// step's name. The conventions used by the workflow generators are:
//
//   - "<base>"            → apply (default)
//   - "<base> (plan)"     → plan
//   - "<base> (destroy)"  → destroy
//
// Helm and kubernetes steps don't split into plan/apply; they get a fixed
// producer-appropriate phase.
func phaseFromStepName(name, producer string) string {
	switch producer {
	case "helm":
		return "install"
	case "kubernetes":
		return "apply"
	case "pulumi":
		// pulumi mirrors terraform's plan/apply split today
		return phaseFromTerraformStepName(name)
	case "terraform":
		return phaseFromTerraformStepName(name)
	}
	return ""
}

func phaseFromTerraformStepName(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.HasSuffix(n, " (plan)"):
		return "plan"
	case strings.HasSuffix(n, " (destroy)"):
		return "destroy"
	default:
		return "apply"
	}
}

// buildParseContext joins producer / phase / cloud into a "/"-delimited path,
// stopping at the first empty segment. Cloud is normalized; CloudPlatformUnknown
// and "" are both dropped. The pipeline's ancestor walk handles the "less
// specific" cases (e.g. terraform/apply when cloud is unknown).
func buildParseContext(producer, phase, cloud string) composite_error.ParseContext {
	if producer == "" {
		return ""
	}
	parts := []string{producer}
	if phase != "" {
		parts = append(parts, phase)
		if cloud != "" && cloud != string(app.CloudPlatformUnknown) {
			parts = append(parts, cloud)
		}
	}
	return composite_error.ParseContext(strings.Join(parts, "/"))
}
