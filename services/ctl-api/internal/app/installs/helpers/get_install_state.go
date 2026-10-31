package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/inputstate"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// GetInstallState reads the current state of the install from the DB, and returns it in a structure that can be used for variable interpolation.
func (h *Helpers) GetInstallState(ctx context.Context, installID string, redacted bool, skipVersionCheck bool) (result *state.State, err error) {
	ctx = routing.WithoutReplica(ctx)
	started := time.Now()
	defer func() { h.stateMetrics.Record(ctx, "get", started, err) }()
	cctx.GetLogger(ctx, h.l).Info("getting install state", zap.String("install_id", installID))

	resp, err := h.RegenerateInstallState(ctx, &pkgstate.ExecuteRegenerationRequest{
		InstallID: installID, TriggeredByID: installID, TriggeredByType: "installs",
	}, skipVersionCheck)
	if err != nil {
		return nil, err
	}
	snapshot := resp.State
	if redacted {
		if err := h.RedactStateInputs(ctx, installID, snapshot); err != nil {
			return nil, err
		}
	}
	if err := h.hydrateInstallMetadata(ctx, installID, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (h *Helpers) GetInstallInputsState(ctx context.Context, installID string, redacted, skipVersionCheck bool) (*state.InputsState, error) {
	ctx = routing.WithoutReplica(ctx)
	var inputs app.InstallInputs
	if err := h.db.WithContext(ctx).Where(app.InstallInputs{InstallID: installID}).Order("created_at DESC").First(&inputs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &state.InputsState{}, nil
		}
		return nil, errors.Wrap(err, "unable to get install inputs")
	}
	var install app.Install
	if err := h.db.WithContext(ctx).Select("id", "app_config_id").Where(app.Install{ID: installID}).First(&install).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}
	cfg, err := h.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, skipVersionCheck)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}
	return ToInputState(&inputs, cfg, redacted), nil
}

func (h *Helpers) getStateInstall(ctx context.Context, installID string) (*app.Install, error) {
	var install app.Install
	res := h.db.WithContext(ctx).
		Preload("App").
		Preload("App.AppSecrets").
		Preload("Org").
		Preload("CreatedBy").
		Preload("AppRunnerConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("AzureAccount").
		Preload("GCPAccount").
		Preload("AWSAccount").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.InstallInputs{}, ".created_at DESC")).Limit(1)
		}).
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find install: %w", res.Error)
	}

	return &install, nil
}

func (h *Helpers) ToInstallStackState(stack *app.InstallStack) *state.InstallStackState {
	return ToInstallStackState(stack)
}

func (h *Helpers) ToInputState(inputs *app.InstallInputs, cfg *app.AppConfig, redacted bool) *state.InputsState {
	return ToInputState(inputs, cfg, redacted)
}

func (h *Helpers) ToCloudAccount(install *app.Install) *state.CloudAccount {
	return ToCloudAccount(install)
}

func (h *Helpers) ToSandboxesState(sandboxRuns []app.InstallSandboxRun) *state.SandboxState {
	if len(sandboxRuns) < 1 {
		return state.NewSandboxState()
	}
	return ToSandboxState(&sandboxRuns[0])
}

func (h *Helpers) ToAppState(currentApp app.App) *state.AppState {
	return ToAppState(currentApp)
}

func (h *Helpers) ToOrgState(org app.Org) *state.OrgState {
	return ToOrgState(org)
}

func (h *Helpers) ToRunnerState(runner app.Runner) *state.RunnerState {
	return ToRunnerState(runner)
}

func (h *Helpers) ToDomainState(run *app.InstallSandboxRun) *state.DomainState {
	return ToDomainState(run)
}

func (h *Helpers) ToComponentState(installComp app.InstallComponent) *state.ComponentState {
	return ToComponentState(installComp)
}

func (h *Helpers) ToComponents(installComps []app.InstallComponent) *state.ComponentsState {
	st := state.NewComponentsState()
	st.Populated = true

	for _, instCmp := range installComps {
		st.Components[instCmp.Component.Name] = h.ToComponentState(instCmp)
	}
	return st
}

func (h *Helpers) ToActions(installActions []app.InstallActionWorkflow) *state.ActionsState {
	st := state.NewActionsState()
	st.Populated = true

	for _, instAct := range installActions {
		st.Workflows[instAct.ActionWorkflow.Name] = h.ToActionWorkflowState(instAct)
	}
	return st
}

func (h *Helpers) ToActionWorkflowState(act app.InstallActionWorkflow) *state.ActionWorkflowState {
	return ToActionWorkflowState(act)
}

func (h *Helpers) hydrateInstallMetadata(ctx context.Context, installID string, is *state.State) error {
	var install app.Install
	if err := h.db.WithContext(ctx).
		Select("id", "labels").
		Preload("RunnerGroup.Settings").
		Where(app.Install{ID: installID}).
		First(&install).Error; err != nil {
		return errors.Wrap(err, "unable to get install metadata")
	}

	h.applyInstallMetadata(&install, is)
	return nil
}

func (h *Helpers) applyInstallMetadata(install *app.Install, is *state.State) {
	is.Labels = map[string]string(install.Labels)
	if is.Runner == nil {
		is.Runner = state.NewRunnerState()
	}
	is.Runner.APIURL = install.RunnerGroup.Settings.RunnerAPIURL
	if is.Runner.APIURL == "" {
		is.Runner.APIURL = h.cfg.RunnerAPIURL
	}
}

func (h *Helpers) RegenerateInstallState(ctx context.Context, req *pkgstate.ExecuteRegenerationRequest, skipVersionCheck bool) (*pkgstate.ExecuteRegenerationResponse, error) {
	if blobstore.GetBlobService(ctx) == nil {
		ctx = blobstore.WithBlobService(ctx, h.blobSvc)
	}
	ctx = blobstore.WithBlobReadEnabled(ctx, h.cfg.BlobReadEnabled)
	ctx = routing.WithoutReplica(ctx)
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(25 * time.Millisecond << (attempt - 1)):
			}
		}
		baseline, err := h.GetInstallStateBaseline(ctx, req.InstallID)
		if err != nil {
			return nil, err
		}
		is := baseline.State
		selected := make(map[pkgstate.PartialName]bool)
		for _, target := range req.Targets {
			selected[target.Name] = true
		}
		for _, partial := range baseline.Head.StalePartials {
			selected[partial] = true
		}
		if is.ID == "" || req.ForceAll {
			for _, partial := range pkgstate.AllPartials {
				selected[partial] = true
			}
		}
		resp := &pkgstate.ExecuteRegenerationResponse{State: is, LastModifiedAt: make(map[pkgstate.PartialName]time.Time)}
		for partial, modified := range req.LastModifiedAt {
			resp.LastModifiedAt[partial] = modified
		}
		if len(selected) == 0 {
			resp.GeneratedAt = time.Now().UTC()
			return resp, nil
		}
		if err := h.waitForInstallStateSources(ctx, baseline.Head.OrgID); err != nil {
			return nil, errors.Wrap(err, "unable to wait for install state source transactions")
		}
		install, err := h.getStateInstall(ctx, req.InstallID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install")
		}
		is.ID = install.ID
		is.Name = install.Name
		resp.AppID, resp.AppName, resp.AppConfigID = install.AppID, install.App.Name, install.AppConfigID
		for _, partial := range pkgstate.AllPartials {
			if !selected[partial] {
				continue
			}
			started := time.Now()
			if err := h.regenerateStalePartial(ctx, install, partial, is, skipVersionCheck); err != nil {
				return nil, errors.Wrapf(err, "unable to regenerate partial %s", partial)
			}
			h.stateMetrics.Record(ctx, "partial."+string(partial), started, nil)
			resp.LastModifiedAt[partial] = time.Now().UTC()
			resp.UpdatedPartials = append(resp.UpdatedPartials, partial)
		}
		MapLegacyFields(is)
		published, err := h.PublishInstallState(ctx, &app.InstallState{
			InstallID: req.InstallID, State: is, GeneratedBy: app.InstallStateGenerateSourceStateManager,
			TriggeredByID: req.TriggeredByID, TriggeredByType: req.TriggeredByType,
		}, baseline.Head, resp.UpdatedPartials)
		if IsInstallStateConflict(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		resp.State = published.State
		resp.GeneratedAt = time.Now().UTC()
		return resp, nil
	}
	return nil, temporal.NewApplicationError("install state kept changing during refresh", InstallStateConflict)
}

// RedactStateInputs masks sensitive inputs at read time; persisted state always holds the real values.
func (h *Helpers) RedactStateInputs(ctx context.Context, installID string, states ...*state.State) error {
	var inputs app.InstallInputs
	res := h.db.WithContext(ctx).
		Where(app.InstallInputs{InstallID: installID}).
		Order(views.TableOrViewName(h.db, &app.InstallInputs{}, ".created_at DESC")).
		Limit(1).
		Find(&inputs)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to get install inputs")
	}

	for name, masked := range inputs.ValuesRedacted {
		if pkggenerics.FromPtrStr(masked) != redactedInputValue {
			continue
		}
		for _, es := range states {
			redactStateInput(es, name)
		}
	}
	return nil
}

func redactStateInput(es *state.State, name string) {
	if es == nil {
		return
	}
	if es.Inputs != nil {
		if _, ok := es.Inputs.Inputs[name]; ok {
			es.Inputs.Inputs[name] = redactedInputValue
		}
	}
	if es.Install != nil {
		if _, ok := es.Install.Inputs[name]; ok {
			es.Install.Inputs[name] = redactedInputValue
		}
	}
}

// redactedInputValue matches the mask install_inputs_view_v1 writes for sensitive inputs.
const redactedInputValue = "********"

// regenerateStalePartial refreshes a single partial into state, reusing the same data fetch path as state gen signal.
func (h *Helpers) regenerateStalePartial(ctx context.Context, install *app.Install, partial pkgstate.PartialName, is *state.State, skipVersionCheck bool) error {
	switch partial {
	case pkgstate.PartialOrg:
		is.Org = h.ToOrgState(install.Org)
	case pkgstate.PartialApp:
		is.App = h.ToAppState(install.App)
	case pkgstate.PartialRunner:
		is.Runner = nil
		if len(install.RunnerGroup.Runners) > 0 {
			is.Runner = h.ToRunnerState(install.RunnerGroup.Runners[0])
		}
	case pkgstate.PartialCloud:
		is.Cloud = h.ToCloudAccount(install)
	case pkgstate.PartialInputs:
		inputs, err := h.GetInstallInputsState(ctx, install.ID, false, skipVersionCheck)
		if err != nil {
			return err
		}
		is.Inputs = inputs
	case pkgstate.PartialDomain:
		sandboxRuns, err := h.getInstallSandboxRuns(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get install sandbox runs")
		}
		if len(sandboxRuns) > 0 {
			is.Domain = h.ToDomainState(&sandboxRuns[0])
		} else {
			is.Domain = h.ToDomainState(nil)
		}
	case pkgstate.PartialSandbox:
		sandboxRuns, err := h.getInstallSandboxRuns(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get install sandbox runs")
		}
		is.Sandbox = h.ToSandboxesState(sandboxRuns)
	case pkgstate.PartialStack:
		stack, err := h.getInstallStack(ctx, install.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err, "unable to get install stack")
		}
		is.InstallStack = h.ToInstallStackState(stack)
	case pkgstate.PartialSecrets:
		secrets, err := h.getSecrets(ctx, install.ID, install.RunnerID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err, "unable to get secrets")
		}
		is.Secrets = secrets
	case pkgstate.PartialComponents:
		installComps, err := h.getInstallComponentsState(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get install components")
		}
		comps := h.ToComponents(installComps)
		is.Components = make(map[string]any, len(comps.Components))
		for name, c := range comps.Components {
			cMap, err := state.AsMap(c)
			if err != nil {
				return errors.Wrap(err, "unable to create component map")
			}
			is.Components[name] = cMap
		}
	case pkgstate.PartialActions:
		actions, err := h.getInstallActionWorkflows(ctx, install.ID, true)
		if err != nil {
			return errors.Wrap(err, "unable to get actions")
		}
		is.Actions = h.ToActions(actions)
	default:
		return errors.Errorf("unknown partial: %s", partial)
	}

	return nil
}

func (h *Helpers) MapLegacyFields(is *state.State) {
	MapLegacyFields(is)
}

func ToInstallStackState(stack *app.InstallStack) *state.InstallStackState {
	if stack == nil || len(stack.InstallStackVersions) < 1 {
		return nil
	}
	is := state.NewInstallStackState()
	is.Populated = true
	version := stack.InstallStackVersions[0]
	is.QuickLinkURL = version.QuickLinkURL
	is.TemplateURL = version.TemplateURL
	is.TemplateJSON = string(version.Contents)
	is.Checksum = version.Checksum
	is.Status = string(version.Status.Status)
	is.Outputs = stack.InstallStackOutputs.DataContents
	return is
}

func ToInputState(inputs *app.InstallInputs, cfg *app.AppConfig, redacted bool) *state.InputsState {
	return inputstate.Resolve(inputs, cfg, redacted)
}

func ToAppState(currentApp app.App) *state.AppState {
	st := state.NewAppState()
	st.Populated = true
	st.ID = currentApp.ID
	st.Name = currentApp.Name
	st.Status = string(currentApp.Status)
	for _, secr := range currentApp.AppSecrets {
		st.Variables[secr.Name] = secr.Value
	}
	return st
}

func ToOrgState(org app.Org) *state.OrgState {
	st := state.NewOrgState()
	st.Populated = true
	st.ID = org.ID
	st.Name = org.Name
	st.Status = string(org.Status)
	return st
}

func ToRunnerState(runner app.Runner) *state.RunnerState {
	st := state.NewRunnerState()
	st.Populated = true
	st.ID = runner.ID
	st.RunnerGroupID = runner.RunnerGroupID
	st.Status = string(runner.Status)
	return st
}

func ToDomainState(run *app.InstallSandboxRun) *state.DomainState {
	st := state.NewDomainState()
	if run == nil {
		return st
	}
	if v, ok := run.Outputs["public_domain"].(string); ok {
		st.PublicDomain = v
	}
	if v, ok := run.Outputs["internal_domain"].(string); ok {
		st.InternalDomain = v
	}
	return st
}

func ToComponentState(installComp app.InstallComponent) *state.ComponentState {
	st := state.NewComponentState()
	st.Populated = true
	st.Name = installComp.Component.Name
	st.ComponentID = installComp.ComponentID
	st.InstallComponentID = installComp.ID
	if len(installComp.InstallDeploys) > 0 {
		st.Status = string(installComp.InstallDeploys[0].Status)
		st.BuildID = string(installComp.InstallDeploys[0].ComponentBuildID)
		st.Outputs = installComp.InstallDeploys[0].Outputs
	}
	return st
}

func ToActionWorkflowState(act app.InstallActionWorkflow) *state.ActionWorkflowState {
	st := state.NewActionWorkflowState()
	st.Populated = true
	st.Status = string(act.Status)
	st.ID = act.ActionWorkflow.ID
	for _, run := range act.Runs {
		if run.RunnerJob != nil {
			st.Outputs = run.RunnerJob.ParsedOutputs
			break
		}
	}
	return st
}

func ToSecretsState(parsedOutputs map[string]interface{}) *state.SecretsState {
	empty := state.NewSecretsState()
	if len(parsedOutputs) == 0 {
		return &empty
	}
	var secretsState state.SecretsState
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
			pkggenerics.StringToMapDecodeHook(),
		),
		WeaklyTypedInput: true,
		Result:           &secretsState,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return &empty
	}
	if err := decoder.Decode(parsedOutputs); err != nil {
		return &empty
	}
	if secretsState == nil {
		return &empty
	}
	return &secretsState
}

func ToSandboxState(run *app.InstallSandboxRun) *state.SandboxState {
	st := state.NewSandboxState()
	st.Populated = true
	st.BuildID = run.AppSandboxBuildID
	st.Status = string(run.Status)
	st.Outputs = run.Outputs
	if run.AppSandboxConfig.PublicGitVCSConfig != nil {
		st.Type = run.AppSandboxConfig.PublicGitVCSConfig.Directory
		st.Version = run.AppSandboxConfig.PublicGitVCSConfig.Branch
	}
	if run.AppSandboxConfig.ConnectedGithubVCSConfig != nil {
		st.Type = run.AppSandboxConfig.ConnectedGithubVCSConfig.Directory
		st.Version = run.AppSandboxConfig.ConnectedGithubVCSConfig.Branch
	}
	return st
}

func ToCloudAccount(install *app.Install) *state.CloudAccount {
	st := state.NewCloudAccount()
	if install.AWSAccount != nil {
		st.AWS = &state.AWSCloudAccount{Region: install.AWSAccount.Region}
	}
	if install.AzureAccount != nil {
		st.Azure = &state.AzureCloudAccount{Location: install.AzureAccount.Location}
	}
	if install.GCPAccount != nil {
		st.GCP = &state.GCPCloudAccount{
			ProjectID: install.GCPAccount.ProjectID,
			Region:    install.GCPAccount.Region,
		}
	}
	return st
}

func MapLegacyFields(is *state.State) {
	is.Install = &state.InstallState{
		Populated: true,
		ID:        is.ID,
		Name:      is.Name,
	}
	if is.Sandbox != nil {
		is.Install.Sandbox = *is.Sandbox
	}
	if is.Domain != nil {
		is.Install.PublicDomain = is.Domain.PublicDomain
		is.Install.InternalDomain = is.Domain.InternalDomain
	}
	if is.Inputs != nil {
		is.Install.Inputs = is.Inputs.Inputs
	}
}
