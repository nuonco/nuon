package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updated"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateInstallInputsRequest struct {
	Inputs           map[string]*string `json:"inputs" validate:"required,gte=1"`
	Role             string             `json:"role"`
	DeployDependents *bool              `json:"deploy_dependents,omitempty" swaggertype:"boolean" extensions:"x-nullable"`
}

func (c *UpdateInstallInputsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateInstallInputs
// @Summary				Updates install input config for app
// @Description.markdown	update_install_inputs.md
// @Tags					installs
// @Accept					json
// @Param					req	body	UpdateInstallInputsRequest	true	"Input"
// @Produce				json
// @Param					install_id	path	string	true	"install ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallInputs
// @Router					/v1/installs/{install_id}/inputs [patch]
func (s *service) UpdateInstallInputs(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req UpdateInstallInputsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if len(install.App.AppInputConfigs) < 1 {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("no app input configs defined on app"),
			Description: "no app input configs defined",
		})
		return
	}

	// Default to deploying dependents when the field is omitted, preserving the
	// historical always-deploy behavior; an explicit false is now respected.
	deployDependents := req.DeployDependents == nil || *req.DeployDependents

	inputs, err := s.applyInstallInputsUpdate(ctx, install, req.Inputs, req.Role, deployDependents, false, app.WorkflowTypeInputUpdate)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, inputs)
}

// applyInstallInputsUpdate merges patch over the install's current inputs,
// persists a new install-inputs revision, and starts (plus enqueues) the
// input-update workflow that reconciles the change. It is shared by the
// inputs PATCH endpoint and any flow that drives install inputs (e.g. the
// component enable/disable toggle, which writes the synthetic enabled input).
func (s *service) applyInstallInputsUpdate(ctx context.Context, install *app.Install, patch map[string]*string, role string, deployDependents bool, planOnly bool, workflowType app.WorkflowType) (*app.InstallInputs, error) {
	latestLatestInstallInputs, err := s.getLatestInstallInputs(ctx, install.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest install inputs: %w", err)
	}

	pinnedAppInputConfig, err := s.helpers.GetPinnedAppInputConfig(ctx, install.AppID, install.AppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest app input config: %w", err)
	}
	if pinnedAppInputConfig == nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("invalid install inputs provided"),
			Description: "inputs provided on install, that are not defined on the app",
		}
	}

	// Reject any install_stack (customer) sourced inputs in the provided subset.
	// This intentionally operates ONLY on the subset the caller sent — existing
	// customer-sourced values carried over by the merge are preserved, not re-validated.
	if err := s.validateVendorSourceInputs(pinnedAppInputConfig, patch); err != nil {
		return nil, err
	}

	// Merge the provided subset over the install's current inputs, then validate the
	// full resulting set so required inputs remain satisfied after a partial update.
	merged := mergeInstallInputs(latestLatestInstallInputs.Values, patch, pinnedAppInputConfig)
	if err := s.helpers.ValidateInstallInputs(ctx, pinnedAppInputConfig, merged); err != nil {
		return nil, err
	}

	// Reject enabling a component whose dependency is disabled. A component may
	// only be turned on when every component it depends on (declared deps or
	// output refs) is effectively enabled in the resulting desired state.
	if err := s.validateComponentEnableDependencies(ctx, install, patch, merged); err != nil {
		return nil, err
	}

	inputs, changedInputs, changedInputValues, err := s.newInstallInputs(
		ctx,
		latestLatestInstallInputs,
		pinnedAppInputConfig,
		merged,
		patch,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create install inputs: %w", err)
	}

	workflow, err := s.helpers.CreateAndStartInputUpdateWorkflow(
		ctx,
		install.ID,
		*changedInputs,
		changedInputValues,
		role,
		deployDependents,
		planOnly,
		workflowType,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create install inputs: %w", err)
	}

	// Enqueue queue signals so the input-update workflow runs.
	signalsQueueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
	if err != nil {
		return nil, err
	}
	workflowsQueueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
	if err != nil {
		return nil, err
	}
	if err := s.enqueueInstallSignal(ctx, signalsQueueID, &updated.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		return nil, fmt.Errorf("enqueue signal: %w", err)
	}
	if err := s.enqueueInstallSignal(ctx, workflowsQueueID, &executeflow.Signal{
		WorkflowID: workflow.ID,
	}, workflow.ID, "install_workflows"); err != nil {
		return nil, fmt.Errorf("enqueue signal: %w", err)
	}

	inputs.WorkflowID = &workflow.ID
	return inputs, nil
}

func (s *service) getLatestInstallInputs(ctx context.Context, installID string) (*app.InstallInputs, error) {
	installInputs := app.InstallInputs{}
	res := s.db.WithContext(ctx).
		Where("install_id = ?", installID).
		Order("created_at DESC").
		First(&installInputs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install inputs: %w", res.Error)
	}

	return &installInputs, nil
}

func (s *service) getLatestAppInputConfig(ctx context.Context, appID string) (*app.AppInputConfig, error) {
	appInputConfig := app.AppInputConfig{}
	res := s.db.WithContext(ctx).
		Preload("AppInputs").
		Where("app_id = ?", appID).
		Order("created_at DESC").
		First(&appInputConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app input config: %w", res.Error)
	}

	return &appInputConfig, nil
}

func (s *service) newInstallInputs(
	ctx context.Context,
	installInputs *app.InstallInputs,
	appInputConfig *app.AppInputConfig,
	merged map[string]*string,
	patch map[string]*string,
) (*app.InstallInputs, *[]string, string, error) {
	changed, err := helpers.ComputeChangedInputs(
		installInputs.Values,
		patch,
		appInputConfig.AppInputs,
	)
	if err != nil {
		return nil, nil, "", fmt.Errorf("unable to compute changed inputs: %w", err)
	}

	// this update will be tied to the latest AppInputConfigID for the app
	obj := &app.InstallInputs{
		AppInputConfigID: appInputConfig.ID,
		InstallID:        installInputs.InstallID,
		Values:           pgtype.Hstore(merged),
	}
	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return nil, nil, "", fmt.Errorf("unable to create install inputs: %w", res.Error)
	}

	latestInstallInputs, err := s.getLatestInstallInputs(ctx, installInputs.InstallID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("unable to get latest install inputs: %w", err)
	}

	latestInstallInputs.Values = nil

	return latestInstallInputs, &changed.Names, changed.ChangedValuesJSON, nil
}

// mergeInstallInputs overlays the provided subset onto the install's existing input
// values and drops any inputs no longer defined in the pinned app input config.
func mergeInstallInputs(existing map[string]*string, patch map[string]*string, appInputConfig *app.AppInputConfig) map[string]*string {
	merged := map[string]*string{}
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range patch {
		merged[k] = v
	}

	appInputNames := map[string]struct{}{}
	for _, input := range appInputConfig.AppInputs {
		appInputNames[input.Name] = struct{}{}
	}
	for k := range merged {
		if _, ok := appInputNames[k]; !ok {
			delete(merged, k)
		}
	}

	return merged
}

func (s *service) validateVendorSourceInputs(appInputConfig *app.AppInputConfig, inputs map[string]*string) error {
	appInputSources := map[string]app.AppInputSource{}
	for _, input := range appInputConfig.AppInputs {
		appInputSources[input.Name] = input.Source
	}

	for name := range inputs {
		source, ok := appInputSources[name]
		if !ok {
			return stderr.ErrUser{
				Err:         fmt.Errorf("input %s is not defined in app input config", name),
				Description: "input " + name + " does not exist in the app inputs",
			}
		}

		// Reject customer sourced inputs
		if source == app.AppInputSourceCustomer {
			return stderr.ErrUser{
				Err:         fmt.Errorf("%s has source install_stack, cannot be updated via api", name),
				Description: name + " has source install_stack and cannot be updated via the api",
			}
		}
	}

	return nil
}

// validateComponentEnableDependencies rejects an inputs update that turns a
// component on while one of its dependencies is disabled. The check is scoped
// to components the patch is enabling (synthetic enabled input set to true), so
// it never blocks disabling a component (which cascades to its dependents) — it
// only stops a user from authoring the inconsistent "enabled component depends
// on a disabled component" state. merged is the full resulting desired input
// state used to resolve effective-enabled.
func (s *service) validateComponentEnableDependencies(ctx context.Context, install *app.Install, patch, merged map[string]*string) error {
	overriddenNames := make(map[string]struct{})
	for name := range patch {
		kind, compName, ok := config.ParseComponentOverrideInputName(name)
		if !ok || kind != config.ComponentOverrideKindEnabled {
			continue
		}
		overriddenNames[compName] = struct{}{}
	}
	if len(overriddenNames) == 0 {
		return nil
	}

	appCfg, err := s.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, true)
	if err != nil {
		return fmt.Errorf("unable to get app config for enable validation: %w", err)
	}

	cccByID := make(map[string]*app.ComponentConfigConnection, len(appCfg.ComponentConfigConnections))
	cccByName := make(map[string]*app.ComponentConfigConnection, len(appCfg.ComponentConfigConnections))
	for i := range appCfg.ComponentConfigConnections {
		ccc := &appCfg.ComponentConfigConnections[i]
		cccByID[ccc.ComponentID] = ccc
		cccByName[ccc.Component.Name] = ccc
	}

	resolver := app.NewComponentEnablementResolver(cccByID, merged)

	for compName := range overriddenNames {
		ccc, ok := cccByName[compName]
		if !ok || !ccc.IsToggleable() {
			continue
		}
		// Resolve the desired own-enabled state exactly as the runtime does
		// (null or unparseable input falls back to default_enabled), so a
		// component left at a true default is still validated.
		if !app.ComponentEnabledFromInputs(merged, ccc) {
			continue
		}
		disabledDeps := resolver.DisabledDependencies(ccc.ComponentID)
		if len(disabledDeps) == 0 {
			continue
		}
		depNames := make([]string, 0, len(disabledDeps))
		for _, depID := range disabledDeps {
			if dep, ok := cccByID[depID]; ok {
				depNames = append(depNames, dep.Component.Name)
			} else {
				depNames = append(depNames, depID)
			}
		}
		joined := strings.Join(depNames, ", ")
		return stderr.ErrUser{
			Err:         fmt.Errorf("cannot enable component %s: depends on disabled component(s) %s", compName, joined),
			Description: fmt.Sprintf("Cannot enable %q because it depends on disabled component(s): %s. Enable %s first.", compName, joined, joined),
		}
	}

	return nil
}
