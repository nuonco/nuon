package alignment

import (
	"context"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/appconfiggraph"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type InstallAlignmentResourceKind string // @name service.InstallAlignmentResourceKind

const (
	InstallAlignmentResourceKindComponent InstallAlignmentResourceKind = "component"
	InstallAlignmentResourceKindImage     InstallAlignmentResourceKind = "image"
	InstallAlignmentResourceKindSandbox   InstallAlignmentResourceKind = "sandbox"
	InstallAlignmentResourceKindStack     InstallAlignmentResourceKind = "stack"
)

type InstallAlignmentCategory string // @name service.InstallAlignmentCategory

const (
	InstallAlignmentCategoryChangesNotApplied  InstallAlignmentCategory = "changes-not-applied"
	InstallAlignmentCategoryStateDriftDetected InstallAlignmentCategory = "state-drift-detected"
	InstallAlignmentCategoryInSync             InstallAlignmentCategory = "in-sync"
	InstallAlignmentCategoryUnknown            InstallAlignmentCategory = "unknown"
	InstallAlignmentCategoryDisabled           InstallAlignmentCategory = "disabled"
	InstallAlignmentCategoryTornDown           InstallAlignmentCategory = "torn-down"
)

type InstallAlignmentOrigin struct {
	WorkflowID   string                 `json:"workflow_id,omitempty"`
	WorkflowType app.WorkflowType       `json:"workflow_type,omitempty"`
	Run          *InstallOverviewCommit `json:"run,omitempty"`
} // @name service.InstallAlignmentOrigin

type InstallAlignmentOpenDeployment struct {
	ID        string           `json:"id"`
	Status    string           `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	Type      app.WorkflowType `json:"type,omitempty"`
} // @name service.InstallAlignmentOpenDeployment

type InstallAlignmentPlan struct {
	WorkflowID     string                       `json:"workflow_id"`
	StepID         string                       `json:"step_id"`
	ApprovalID     string                       `json:"approval_id"`
	OwnerID        string                       `json:"owner_id"`
	OwnerType      string                       `json:"owner_type"`
	Type           app.WorkflowStepApprovalType `json:"type"`
	PlanOnly       bool                         `json:"plan_only"`
	CreatedAt      time.Time                    `json:"created_at"`
	ChangesCreate  int                          `json:"changes_create"`
	ChangesUpdate  int                          `json:"changes_update"`
	ChangesDelete  int                          `json:"changes_delete"`
	ChangesReplace int                          `json:"changes_replace"`
	ChangesNoop    int                          `json:"changes_noop"`
} // @name service.InstallAlignmentPlan

type InstallAlignmentResource struct {
	ID               string                          `json:"id"`
	Kind             InstallAlignmentResourceKind    `json:"kind"`
	Name             string                          `json:"name"`
	ComponentID      string                          `json:"component_id,omitempty"`
	ComponentType    app.ComponentType               `json:"component_type,omitempty" swaggertype:"string"`
	Status           string                          `json:"status"`
	Category         InstallAlignmentCategory        `json:"category"`
	Reason           string                          `json:"reason,omitempty"`
	ExpectedBuildID  string                          `json:"expected_build_id,omitempty"`
	AppliedBuildID   string                          `json:"applied_build_id,omitempty"`
	AppliedDeployID  string                          `json:"applied_deploy_id,omitempty"`
	AppliedDeployAt  *time.Time                      `json:"applied_deploy_at,omitempty"`
	ExpectedConfigID string                          `json:"expected_config_id,omitempty"`
	AppliedConfigID  string                          `json:"applied_config_id,omitempty"`
	AppliedBy        *InstallAlignmentOrigin         `json:"applied_by,omitempty"`
	OpenDeployment   *InstallAlignmentOpenDeployment `json:"open_deployment,omitempty"`
	Plan             *InstallAlignmentPlan           `json:"plan,omitempty"`

	appliedAt         *time.Time
	appliedWorkflowID string
	appliedWorkflow   app.WorkflowType
	appliedRunID      string
	inputsDiffer      bool
	inputsUnknown     bool
} // @name service.InstallAlignmentResource

type InstallAlignment struct {
	Error     string                     `json:"error,omitempty"`
	Resources []InstallAlignmentResource `json:"resources"`
} // @name service.InstallAlignment

type installAlignmentState struct {
	install          *app.Install
	components       []app.InstallComponent
	targets          map[string]*installhelpers.ComponentTarget
	enablement       *appconfiggraph.ComponentEnablementResolver
	appliedDeploys   map[string]*app.InstallDeploy
	sync             map[string]*componentSync
	sandbox          *app.InstallSandbox
	sandboxRow       *sandboxSync
	appliedSandbox   *app.InstallSandboxRun
	expectedSandbox  *app.AppSandboxConfig
	stack            *app.InstallStack
	stackHasVersions bool
	inputs           *app.InstallInputs
	appConfigs       map[string]*app.AppConfig

	appliedBuilds         map[string]*app.ComponentBuild
	consumedBuilds        map[string]*app.ComponentBuild
	consumedSandboxBuilds map[string]*app.AppSandboxBuild
	appliedSandboxBuild   *app.AppSandboxBuild
	appliedSandboxConfig  *app.AppSandboxConfig
	sandboxChanged        bool
	stackChanged          bool
	stackErr              error
}

func (s *Helpers) loadInstallAlignmentState(ctx context.Context, install *app.Install) (*installAlignmentState, error) {
	st := &installAlignmentState{install: install}

	if err := s.db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{InstallID: install.ID, OrgID: install.OrgID}).
		Order("created_at ASC").
		Find(&st.components).Error; err != nil {
		return nil, fmt.Errorf("unable to list install components: %w", err)
	}

	targets, err := s.installs.EffectiveComponentTargetConfigs(ctx, install.AppConfigID)
	if err != nil {
		return nil, err
	}
	st.targets = targets

	var enabledInputs map[string]*string
	inputs, err := s.installs.LatestInstallInputs(ctx, install.ID)
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	if inputs != nil {
		enabledInputs = inputs.Values
	}
	st.inputs = inputs

	var rows []componentSync
	if err := s.db.WithContext(ctx).
		Table(views.DefaultViewName(s.db, &app.DriftedObject{}, 5)).
		Where("install_id = ? AND org_id = ? AND resource_type = ?", install.ID, install.OrgID, "install_component").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to get component sync state: %w", err)
	}
	st.sync = make(map[string]*componentSync, len(rows))
	for i := range rows {
		st.sync[rows[i].InstallComponentID] = &rows[i]
	}

	buildIDs := make([]string, 0, len(st.components))
	liveIDs := make([]string, 0, len(st.components))
	for i := range st.components {
		row := st.sync[st.components[i].ID]
		if row == nil {
			continue
		}
		if st.targets[st.components[i].ComponentID] != nil && row.ExpectedBuildID != "" {
			buildIDs = append(buildIDs, row.ExpectedBuildID)
		}
		if row.LiveID != "" {
			liveIDs = append(liveIDs, row.LiveID)
		}
	}
	builds, err := s.installs.ComponentBuildsWithConfigs(ctx, buildIDs)
	if err != nil {
		return nil, err
	}
	for i := range st.components {
		row := st.sync[st.components[i].ID]
		if row == nil || row.ExpectedBuildID == "" {
			continue
		}
		if target := st.targets[st.components[i].ComponentID]; target != nil {
			target.Build = builds[row.ExpectedBuildID]
		}
	}

	configIDs := []string{install.AppConfigID}
	for _, target := range targets {
		if target.Build != nil {
			configIDs = append(configIDs, target.Build.ComponentConfigConnection.AppConfigID)
		}
	}
	var appConfigs []app.AppConfig
	if err := s.db.WithContext(ctx).Scopes(scopes.WithDisableViews).Select("id").
		Preload("InputConfig.AppInputs").
		Where("id IN ?", configIDs).Find(&appConfigs).Error; err != nil {
		return nil, fmt.Errorf("unable to get alignment input defaults: %w", err)
	}
	st.appConfigs = make(map[string]*app.AppConfig, len(appConfigs))
	for i := range appConfigs {
		st.appConfigs[appConfigs[i].ID] = &appConfigs[i]
	}
	configs := make(map[string]*app.ComponentConfigConnection, len(targets))
	for id, target := range targets {
		configs[id] = target.Config
	}
	st.enablement = appconfiggraph.NewComponentEnablementResolver(configs, enabledInputs)

	st.appliedDeploys, err = s.installs.AppliedInstallDeploysByIDs(ctx, liveIDs)
	if err != nil {
		return nil, err
	}

	var sandbox app.InstallSandbox
	err = s.db.WithContext(ctx).
		Where(app.InstallSandbox{InstallID: install.ID, OrgID: install.OrgID}).
		First(&sandbox).Error
	if err != nil && !isNotFound(err) {
		return nil, fmt.Errorf("unable to get install sandbox: %w", err)
	}
	if err == nil {
		st.sandbox = &sandbox
		st.appliedSandbox, err = s.installs.AppliedInstallSandboxRun(ctx, &sandbox)
		if err != nil {
			return nil, err
		}
		var sandboxRow sandboxSync
		if err := s.db.WithContext(ctx).
			Table(views.DefaultViewName(s.db, &app.DriftedObject{}, 5)).
			Where("install_id = ? AND org_id = ? AND resource_type = ? AND install_sandbox_id = ?", install.ID, install.OrgID, "install_sandbox", sandbox.ID).
			Scan(&sandboxRow).Error; err != nil {
			return nil, fmt.Errorf("unable to get sandbox sync state: %w", err)
		}
		if sandboxRow.InstallSandboxID != "" {
			st.sandboxRow = &sandboxRow
		}
		if install.AppConfigID != "" {
			var cfg app.AppSandboxConfig
			err = s.db.WithContext(ctx).
				Where(app.AppSandboxConfig{AppConfigID: install.AppConfigID}).
				Order("created_at DESC").
				First(&cfg).Error
			if err != nil && !isNotFound(err) {
				return nil, fmt.Errorf("unable to get app sandbox config: %w", err)
			}
			if err == nil {
				st.expectedSandbox = &cfg
			}
		}
	}

	var stack app.InstallStack
	err = s.db.WithContext(ctx).
		Where(app.InstallStack{InstallID: install.ID, OrgID: install.OrgID}).
		First(&stack).Error
	if err != nil && !isNotFound(err) {
		return nil, fmt.Errorf("unable to get install stack: %w", err)
	}
	if err == nil {
		st.stack = &stack
		var count int64
		if err := s.db.WithContext(ctx).
			Model(&app.InstallStackVersion{}).
			Where(app.InstallStackVersion{InstallStackID: stack.ID}).
			Count(&count).Error; err != nil {
			return nil, fmt.Errorf("unable to count install stack versions: %w", err)
		}
		st.stackHasVersions = count > 0
	}

	if err := s.loadDependencyState(ctx, st); err != nil {
		return nil, err
	}

	return st, nil
}

func (st *installAlignmentState) componentEnabled(target *installhelpers.ComponentTarget) bool {
	if target == nil || target.Config == nil {
		return true
	}
	return st.enablement.EffectiveEnabled(target.Config.ComponentID)
}

func (st *installAlignmentState) desiredInputs(buildConfigID string) *state.InputsState {
	current := st.appConfigs[st.install.AppConfigID]
	if current == nil {
		return nil
	}
	inputs := &state.InputsState{}
	if st.inputs != nil {
		inputs = installhelpers.ToInputState(st.inputs, current, false)
	}
	resolved := &state.State{Inputs: inputs}
	if buildConfigID != "" && buildConfigID != current.ID {
		installhelpers.FillInputDefaults(resolved, st.appConfigs[buildConfigID])
	}
	return resolved.Inputs
}

func (res *InstallAlignmentResource) compareInputs(names []string, desired *state.InputsState, appliedDigest string) {
	if res.Category != InstallAlignmentCategoryInSync {
		return
	}
	if appliedDigest == "" {
		if len(names) > 0 {
			res.inputsUnknown = true
			res.Reason = "Inputs not verified. Applied input values weren’t recorded for this deployment."
		}
		return
	}
	if appliedDigest != installhelpers.InputsDigest(names, desired) {
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.inputsDiffer = true
		res.Reason = "Saved input changes haven’t taken effect."
	}
}

func (st *installAlignmentState) componentResource(ic *app.InstallComponent) InstallAlignmentResource {
	kind := InstallAlignmentResourceKindComponent
	if ic.Component.Type.IsImage() {
		kind = InstallAlignmentResourceKindImage
	}
	res := InstallAlignmentResource{
		ID:            ic.ID,
		Kind:          kind,
		Name:          ic.Component.Name,
		ComponentID:   ic.ComponentID,
		ComponentType: ic.Component.Type,
		Status:        string(ic.Status),
	}
	if res.Status == "" {
		res.Status = string(app.InstallComponentStatusUnknown)
	}

	target := st.targets[ic.ComponentID]
	if target != nil {
		res.ExpectedConfigID = st.install.AppConfigID
		if target.Build != nil {
			res.ExpectedBuildID = target.Build.ID
		}
	}
	applied := st.appliedDeploys[ic.ID]
	if applied != nil {
		res.AppliedBuildID = applied.ComponentBuildID
		res.AppliedDeployID = applied.ID
		res.AppliedDeployAt = applied.AppliedAt
		if res.AppliedDeployAt == nil {
			res.AppliedDeployAt = &applied.UpdatedAt
		}
		res.AppliedConfigID = applied.AppConfigID
		if res.AppliedConfigID == "" {
			res.AppliedConfigID = ic.AppConfigRef.AppliedConfigID
		}
		if res.AppliedConfigID == "" && applied.ComponentBuild != nil {
			res.AppliedConfigID = applied.ComponentBuild.ComponentConfigConnection.AppConfigID
		}
		res.appliedAt = applied.AppliedAt
		res.appliedWorkflowID = generics.FromPtrStr(applied.InstallWorkflowID)
	}
	if row := st.sync[ic.ID]; row != nil && row.LiveWorkflowType != "" {
		res.appliedWorkflow = row.LiveWorkflowType
		res.appliedRunID = row.LiveAppBranchRunID
	}
	inactive := ic.Status == app.InstallComponentStatusInactive || ic.Status == app.InstallComponentStatusDeleted || ic.Status == app.InstallComponentStatusDisabled
	last := st.sync[ic.ID]
	tornDown := last != nil && last.LastType == app.InstallDeployTypeTeardown && last.LastStatus == app.InstallDeployStatusInactive &&
		last.LastWorkflowType != app.WorkflowTypeInputUpdate && last.LastWorkflowType != app.WorkflowTypeComponentDisabled

	switch {
	case tornDown && (target == nil || st.componentEnabled(target)):
		res.Category = InstallAlignmentCategoryTornDown
		res.Reason = "component was explicitly torn down on this install"
	case target == nil:
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "component is not part of the install's selected app config"
	case !st.componentEnabled(target) && !inactive && (applied != nil || ic.EverDeployed()):
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "This component is disabled, but is still deployed."
	case !st.componentEnabled(target):
		res.Category = InstallAlignmentCategoryDisabled
		res.Reason = "component is disabled on this install"
	case target.Build == nil:
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "no build is associated with the selected component config"
	case target.Build.Status != app.ComponentBuildStatusActive:
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = fmt.Sprintf("the selected build is %s and cannot be deployed", target.Build.Status)
	case inactive:
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "This component is enabled, but hasn’t been deployed."
	case applied == nil && ic.EverDeployed():
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "the applied deploy for this component could not be determined"
	case applied == nil:
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "This component hasn’t been deployed yet."
	case target.Config.Checksum == "" || applied.ComponentBuild == nil || applied.ComponentBuild.ComponentConfigConnection.Checksum != target.Config.Checksum ||
		!configdiff.ComponentSourcesEqual(applied.ComponentBuild, target.Build):
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "The deployed version doesn’t match this install’s configuration."
	default:
		res.Category = InstallAlignmentCategoryInSync
	}
	if res.Category == InstallAlignmentCategoryInSync {
		switch changed, verified := st.dependenciesAppliedSince(ic.ComponentID, target.Config, applied); {
		case changed:
			res.Category = InstallAlignmentCategoryChangesNotApplied
			res.Reason = "Component dependency changes haven’t been applied."
		case !verified:
			res.Reason = "Dependencies not verified. Upstream builds weren’t recorded for this deployment."
		}
	}
	if res.Kind == InstallAlignmentResourceKindComponent && target != nil && target.Build != nil && applied != nil {
		cfg := &target.Build.ComponentConfigConnection
		inputs := st.desiredInputs(cfg.AppConfigID)
		res.compareInputs(installhelpers.ComponentInputNames(cfg, inputs), inputs, applied.AppliedInputsDigest)
	}
	return res
}

func (st *installAlignmentState) sandboxResource() *InstallAlignmentResource {
	if st.sandbox == nil {
		return nil
	}
	res := InstallAlignmentResource{
		ID:     st.sandbox.ID,
		Kind:   InstallAlignmentResourceKindSandbox,
		Name:   "sandbox",
		Status: string(st.sandbox.Status),
	}
	if st.sandbox.StatusV2.Status != "" {
		res.Status = string(st.sandbox.StatusV2.Status)
	}
	if st.expectedSandbox != nil {
		res.ExpectedConfigID = st.install.AppConfigID
	}
	if st.appliedSandbox != nil {
		res.AppliedConfigID = st.appliedSandbox.AppSandboxConfig.AppConfigID
		res.appliedAt = st.appliedSandbox.AppliedAt
		res.appliedWorkflowID = generics.FromPtrStr(st.appliedSandbox.InstallWorkflowID)
		if row := st.sandboxRow; row != nil && row.LiveWorkflowType != "" {
			res.appliedWorkflow = row.LiveWorkflowType
			res.appliedRunID = row.LiveAppBranchRunID
		}
	}

	switch {
	case st.expectedSandbox == nil:
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "the selected app config has no sandbox config"
	case st.appliedSandbox == nil && st.sandbox.AppConfigRef.AppliedConfigByID != "":
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "The sandbox isn’t currently provisioned."
	case st.appliedSandbox == nil && st.sandbox.Status == app.InstallSandboxStatusActive:
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "the applied sandbox run could not be determined"
	case st.appliedSandbox == nil:
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "The sandbox hasn’t been set up yet."
	case st.sandboxChanged:
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "Sandbox configuration changes haven’t been applied."
	default:
		res.Category = InstallAlignmentCategoryInSync
	}
	if st.expectedSandbox != nil && st.appliedSandbox != nil {
		res.compareInputs(installhelpers.InputNames(st.expectedSandbox.Refs), st.desiredInputs(""), st.appliedSandbox.AppliedInputsDigest)
	}
	return &res
}

func (st *installAlignmentState) stackResource() *InstallAlignmentResource {
	if st.stack == nil {
		return nil
	}
	res := InstallAlignmentResource{
		ID:               st.stack.ID,
		Kind:             InstallAlignmentResourceKindStack,
		Name:             "stack",
		ExpectedConfigID: st.install.AppConfigID,
		AppliedConfigID:  st.stack.AppConfigRef.AppliedConfigID,
		appliedAt:        st.stack.AppConfigRef.AppliedConfigAt,
	}
	applied := st.stack.AppConfigRef.AppliedConfigID
	switch {
	case st.install.AppConfigID == "":
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "the install has no selected app config"
	case applied == "" && st.stackHasVersions:
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "the applied stack version could not be determined"
	case applied == "":
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "The stack hasn’t been set up yet."
	case applied == st.install.AppConfigID:
		res.Category = InstallAlignmentCategoryInSync
		res.Reason = "stack state drift is not scanned"
	case st.stackErr != nil:
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "unable to compare the applied stack config with the selected config"
	case st.stackChanged:
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "Stack configuration changes haven’t been applied."
	default:
		res.Category = InstallAlignmentCategoryInSync
		res.Reason = "stack state drift is not scanned"
	}
	return &res
}

func (s *Helpers) Read(ctx context.Context, install *app.Install) (*InstallAlignment, error) {
	st, err := s.loadInstallAlignmentState(ctx, install)
	if err != nil {
		return nil, err
	}

	alignment := &InstallAlignment{Resources: []InstallAlignmentResource{}}
	if r := st.stackResource(); r != nil {
		alignment.Resources = append(alignment.Resources, *r)
	}
	if r := st.sandboxResource(); r != nil {
		alignment.Resources = append(alignment.Resources, *r)
	}
	for i := range st.components {
		alignment.Resources = append(alignment.Resources, st.componentResource(&st.components[i]))
	}

	ev, err := s.loadAlignmentEvidence(ctx, st)
	if err != nil {
		return nil, err
	}
	open := st.openDeployments()

	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		res.OpenDeployment = open[res.ID]
		res.applyDriftEvidence(ev)
		res.Plan = ev.plans[res.ID]
		if scan := ev.scanPlans[res.ID]; scan != nil && res.Category == InstallAlignmentCategoryStateDriftDetected {
			res.Plan = scan
		}
		if res.Plan != nil && res.appliedAt != nil && res.Plan.CreatedAt.Before(*res.appliedAt) {
			res.Plan = nil
		}
	}
	if err := s.loadOrigins(ctx, alignment); err != nil {
		return nil, err
	}
	return alignment, nil
}

func (res *InstallAlignmentResource) applyDriftEvidence(ev *alignmentEvidence) {
	if (res.inputsDiffer || res.inputsUnknown) && ev.inputsAligned[res.ID] {
		res.Category = InstallAlignmentCategoryInSync
		res.Reason = ""
	}
	if res.Category != InstallAlignmentCategoryInSync && !res.inputsUnknown {
		return
	}
	if ev.drifted[res.ID] {
		res.Category = InstallAlignmentCategoryStateDriftDetected
		res.Reason = "Live resources differ from what was last applied."
	}
}

type alignmentEvidence struct {
	plans         map[string]*InstallAlignmentPlan
	scanPlans     map[string]*InstallAlignmentPlan
	drifted       map[string]bool
	inputsAligned map[string]bool
}

func (s *Helpers) loadAlignmentEvidence(ctx context.Context, st *installAlignmentState) (*alignmentEvidence, error) {
	ev := &alignmentEvidence{
		plans:         map[string]*InstallAlignmentPlan{},
		scanPlans:     map[string]*InstallAlignmentPlan{},
		drifted:       map[string]bool{},
		inputsAligned: map[string]bool{},
	}

	for i := range st.components {
		icID := st.components[i].ID
		row := st.sync[icID]
		if row == nil {
			continue
		}
		applied := st.appliedDeploys[icID]
		target := st.targets[st.components[i].ComponentID]
		if target != nil && target.Build != nil && target.Build.ID == row.LastBuildID &&
			row.LastStatus == app.InstallDeployStatusAutoSkipped && row.LastPlannedAt != nil && row.LastPolicyAction != "skip_apply" &&
			(applied == nil || applied.AppliedAt == nil || !row.LastPlannedAt.Before(*applied.AppliedAt)) {
			cfg := &target.Build.ComponentConfigConnection
			inputs := st.desiredInputs(cfg.AppConfigID)
			ev.inputsAligned[icID] = row.LastPlannedInputsDigest == installhelpers.InputsDigest(installhelpers.ComponentInputNames(cfg, inputs), inputs)
		}
		ev.drifted[icID] = row.ScanDrifted
	}
	var approvalIDs []string
	for _, row := range st.sync {
		for _, id := range []string{row.RealPlanApprovalID, row.ScanPlanApprovalID} {
			if id != "" {
				approvalIDs = append(approvalIDs, id)
			}
		}
	}
	if len(approvalIDs) > 0 {
		approvals, err := s.loadAlignmentApprovals(ctx, approvalIDs)
		if err != nil {
			return nil, err
		}
		for icID, row := range st.sync {
			if a := approvals[row.RealPlanApprovalID]; a != nil {
				ev.plans[icID] = alignmentPlanFromApproval(a)
			}
			if a := approvals[row.ScanPlanApprovalID]; a != nil {
				ev.scanPlans[icID] = alignmentPlanFromApproval(a)
			}
		}
	}

	if st.sandbox != nil && st.sandboxRow != nil {
		row := st.sandboxRow
		if st.expectedSandbox != nil && row.LastAppSandboxConfigID == st.expectedSandbox.ID &&
			row.LastStatus == app.SandboxRunAutoSkipped && row.LastPlannedAt != nil &&
			row.LastPolicyAction != "skip_apply" &&
			(st.appliedSandbox == nil || st.appliedSandbox.AppliedAt == nil || !row.LastPlannedAt.Before(*st.appliedSandbox.AppliedAt)) {
			ev.inputsAligned[st.sandbox.ID] = row.LastPlannedInputsDigest == installhelpers.InputsDigest(installhelpers.InputNames(st.expectedSandbox.Refs), st.desiredInputs(""))
		}
		if row.ScanCreatedAt != nil && st.appliedSandbox != nil && sandboxScanMatchesApplied(st, row.ScanAppSandboxConfigID) &&
			(st.appliedSandbox.AppliedAt == nil || !row.ScanCreatedAt.Before(*st.appliedSandbox.AppliedAt)) &&
			(row.LastCreatedAt == nil || !row.ScanCreatedAt.Before(*row.LastCreatedAt)) {
			ev.drifted[st.sandbox.ID] = row.ScanStatus == app.SandboxRunStatus(app.WorkflowStepDrifted)
		}

		var approvalIDs []string
		for _, id := range []string{row.RealPlanApprovalID, row.ScanPlanApprovalID} {
			if id != "" {
				approvalIDs = append(approvalIDs, id)
			}
		}
		if len(approvalIDs) > 0 {
			approvals, err := s.loadAlignmentApprovals(ctx, approvalIDs)
			if err != nil {
				return nil, err
			}
			if a := approvals[row.RealPlanApprovalID]; a != nil {
				ev.plans[st.sandbox.ID] = alignmentPlanFromApproval(a)
			}
			if a := approvals[row.ScanPlanApprovalID]; a != nil {
				ev.scanPlans[st.sandbox.ID] = alignmentPlanFromApproval(a)
			}
		}
	}

	return ev, nil
}

func sandboxScanMatchesApplied(st *installAlignmentState, scanAppSandboxConfigID string) bool {
	if scanAppSandboxConfigID == st.appliedSandbox.AppSandboxConfigID {
		return true
	}
	if st.expectedSandbox == nil || scanAppSandboxConfigID != st.expectedSandbox.ID {
		return false
	}
	if st.appliedSandbox.AppSandboxConfigID == st.expectedSandbox.ID {
		return true
	}
	return !st.sandboxChanged
}

type alignmentApproval struct {
	approval *app.WorkflowStepApproval
	workflow *app.Workflow
}

func (s *Helpers) loadAlignmentApprovals(ctx context.Context, ids []string) (map[string]*alignmentApproval, error) {
	var approvals []app.WorkflowStepApproval
	if err := s.db.WithContext(ctx).
		Select("id", "org_id", "install_workflow_step_id", "owner_id", "owner_type", "type", "created_at",
			"changes_create", "changes_update", "changes_delete", "changes_replace", "changes_noop").
		Preload("InstallWorkflowStep", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "install_workflow_id")
		}).
		Where("id IN ?", ids).
		Find(&approvals).Error; err != nil {
		return nil, fmt.Errorf("unable to load approvals: %w", err)
	}
	workflowIDs := make([]string, 0, len(approvals))
	for i := range approvals {
		workflowIDs = append(workflowIDs, approvals[i].InstallWorkflowStep.InstallWorkflowID)
	}
	workflowsByID := map[string]*app.Workflow{}
	if len(workflowIDs) > 0 {
		var workflows []app.Workflow
		if err := s.db.WithContext(ctx).
			Select("id", "plan_only").
			Where("id IN ?", workflowIDs).
			Find(&workflows).Error; err != nil {
			return nil, fmt.Errorf("unable to load approval workflows: %w", err)
		}
		for i := range workflows {
			workflowsByID[workflows[i].ID] = &workflows[i]
		}
	}
	out := make(map[string]*alignmentApproval, len(approvals))
	for i := range approvals {
		wf := workflowsByID[approvals[i].InstallWorkflowStep.InstallWorkflowID]
		if wf == nil {
			continue
		}
		out[approvals[i].ID] = &alignmentApproval{approval: &approvals[i], workflow: wf}
	}
	return out, nil
}

func alignmentPlanFromApproval(aa *alignmentApproval) *InstallAlignmentPlan {
	a := aa.approval
	plan := &InstallAlignmentPlan{
		WorkflowID:     aa.workflow.ID,
		StepID:         a.InstallWorkflowStepID,
		ApprovalID:     a.ID,
		OwnerID:        a.OwnerID,
		OwnerType:      a.OwnerType,
		Type:           a.Type,
		PlanOnly:       aa.workflow.PlanOnly,
		CreatedAt:      a.CreatedAt,
		ChangesCreate:  a.ChangesCreate,
		ChangesUpdate:  a.ChangesUpdate,
		ChangesDelete:  a.ChangesDelete,
		ChangesReplace: a.ChangesReplace,
		ChangesNoop:    a.ChangesNoop,
	}
	return plan
}

type componentSync struct {
	InstallComponentID      string
	ExpectedBuildID         string
	LiveID                  string
	LastBuildID             string
	LastType                app.InstallDeployType
	LastStatus              app.InstallDeployStatus
	LastPlannedAt           *time.Time
	LastPlannedInputsDigest string
	LastPolicyAction        string
	LastWorkflowID          string
	LastWorkflowType        app.WorkflowType
	LastWorkflowStatus      string
	LastWorkflowCreatedAt   *time.Time
	ScanDrifted             bool
	LiveWorkflowType        app.WorkflowType
	LiveAppBranchRunID      string
	RealPlanApprovalID      string
	ScanPlanApprovalID      string
}

type sandboxSync struct {
	InstallSandboxID        string
	LiveWorkflowType        app.WorkflowType
	LiveAppBranchRunID      string
	LastAppSandboxConfigID  string
	LastStatus              app.SandboxRunStatus
	LastPlannedAt           *time.Time
	LastPlannedInputsDigest string
	LastPolicyAction        string
	LastCreatedAt           *time.Time
	ScanID                  string
	ScanStatus              app.SandboxRunStatus
	ScanAppSandboxConfigID  string
	ScanCreatedAt           *time.Time
	RealPlanApprovalID      string
	ScanPlanApprovalID      string
}

func (st *installAlignmentState) openDeployments() map[string]*InstallAlignmentOpenDeployment {
	out := map[string]*InstallAlignmentOpenDeployment{}
	for icID, row := range st.sync {
		status := row.LastWorkflowStatus
		if row.LastWorkflowID == "" || row.LastWorkflowCreatedAt == nil ||
			(status != string(app.StatusFailedPendingRetry) && !slices.Contains(installhelpers.InstallDeploymentActiveStatuses, status)) {
			continue
		}
		if status == "" {
			status = string(app.StatusPending)
		}
		out[icID] = &InstallAlignmentOpenDeployment{ID: row.LastWorkflowID, Status: status, CreatedAt: *row.LastWorkflowCreatedAt, Type: row.LastWorkflowType}
	}
	return out
}
