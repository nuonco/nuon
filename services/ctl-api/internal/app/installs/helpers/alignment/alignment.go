package alignment

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitinstallstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/stackrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/appconfiggraph"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
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

type InstallAlignmentAttempt struct {
	ID         string                 `json:"id"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	WorkflowID string                 `json:"workflow_id,omitempty"`
	Type       app.WorkflowType       `json:"type,omitempty"`
	BranchRun  *InstallOverviewCommit `json:"branch_run,omitempty"`
} // @name service.InstallAlignmentAttempt

type InstallAlignmentOpenDeployment struct {
	ID        string           `json:"id"`
	Status    string           `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	Type      app.WorkflowType `json:"type,omitempty"`
} // @name service.InstallAlignmentOpenDeployment

type InstallAlignmentExplanation struct {
	Summary    string `json:"summary"`
	Status     string `json:"status,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
	StepID     string `json:"step_id,omitempty"`
} // @name service.InstallAlignmentExplanation

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
	ExpectedConfigID string                          `json:"expected_config_id,omitempty"`
	AppliedConfigID  string                          `json:"applied_config_id,omitempty"`
	LastAttempt      *InstallAlignmentAttempt        `json:"last_attempt,omitempty"`
	OpenDeployment   *InstallAlignmentOpenDeployment `json:"open_deployment,omitempty"`
	Plan             *InstallAlignmentPlan           `json:"plan,omitempty"`
	NewerBranchRuns  *int                            `json:"newer_branch_runs,omitempty" extensions:"x-nullable"`
	Explanation      *InstallAlignmentExplanation    `json:"explanation,omitempty"`

	appliedAt     *time.Time
	inputsDiffer  bool
	inputsUnknown bool
} // @name service.InstallAlignmentResource

type InstallAlignment struct {
	Error     string                     `json:"error,omitempty"`
	Resources []InstallAlignmentResource `json:"resources"`
} // @name service.InstallAlignment

type alignmentTarget struct {
	kind             InstallAlignmentResourceKind
	installComponent *app.InstallComponent
	sandbox          *app.InstallSandbox
	stack            *app.InstallStack
}

type installAlignmentState struct {
	install          *app.Install
	components       []app.InstallComponent
	targets          map[string]*installhelpers.ComponentTarget
	enablement       *appconfiggraph.ComponentEnablementResolver
	appliedDeploys   map[string]*app.InstallDeploy
	latestDeploys    map[string]*app.InstallDeploy
	sandbox          *app.InstallSandbox
	appliedSandbox   *app.InstallSandboxRun
	expectedSandbox  *app.AppSandboxConfig
	stack            *app.InstallStack
	stackHasVersions bool
	compositeDiff    *app.InstallConfigDiff
	compositeErr     error
	inputs           *app.InstallInputs
	appConfigs       map[string]*app.AppConfig
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

	targets, err := s.installs.EffectiveComponentTargets(ctx, install.AppConfigID)
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

	st.appliedDeploys, err = s.installs.AppliedInstallDeploys(ctx, st.components)
	if err != nil {
		return nil, err
	}

	if len(st.components) > 0 {
		ids := make([]string, 0, len(st.components))
		for _, ic := range st.components {
			ids = append(ids, ic.ID)
		}
		st.latestDeploys, err = s.latestInstallDeploysByComponent(ctx, ids, false)
		if err != nil {
			return nil, err
		}
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

	if install.AppConfigID != "" {
		diffCtx := blobstore.WithBlobService(ctx, s.blobSvc)
		baselines, err := s.installs.InstallCompositeBaselines(diffCtx, install)
		st.compositeErr = err
		if err == nil {
			st.compositeDiff, st.compositeErr = configdiff.ComputeCompositeInstallConfigDiff(diffCtx, s.db, baselines, install.AppConfigID)
		}
		if st.compositeErr != nil {
			s.l.Warn("unable to compute install alignment config diff",
				zap.String("install_id", install.ID),
				zap.Error(st.compositeErr))
		}
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
		res.AppliedConfigID = applied.AppConfigID
		if res.AppliedConfigID == "" {
			res.AppliedConfigID = ic.AppConfigRef.AppliedConfigID
		}
		if res.AppliedConfigID == "" {
			res.AppliedConfigID = applied.ComponentBuild.ComponentConfigConnection.AppConfigID
		}
		res.appliedAt = applied.AppliedAt
	}
	inactive := ic.Status == app.InstallComponentStatusInactive || ic.Status == app.InstallComponentStatusDeleted || ic.Status == app.InstallComponentStatusDisabled
	last := st.latestDeploys[ic.ID]
	tornDown := last != nil && last.Type == app.InstallDeployTypeTeardown && last.Status == app.InstallDeployStatusInactive &&
		(last.InstallWorkflow == nil || (last.InstallWorkflow.Type != app.WorkflowTypeInputUpdate && last.InstallWorkflow.Type != app.WorkflowTypeComponentDisabled))

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
	case target.Config.Checksum == "" || applied.ComponentBuild.ComponentConfigConnection.Checksum != target.Config.Checksum ||
		!configdiff.ComponentSourcesEqual(&applied.ComponentBuild, target.Build):
		res.Category = InstallAlignmentCategoryChangesNotApplied
		res.Reason = "The deployed version doesn’t match this install’s configuration."
	default:
		res.Category = InstallAlignmentCategoryInSync
	}
	if res.Category == InstallAlignmentCategoryInSync {
		if st.compositeErr != nil {
			res.Category = InstallAlignmentCategoryUnknown
			res.Reason = "unable to compare the applied component dependencies with the selected config"
		} else if st.compositeDiff != nil {
			for _, entry := range st.compositeDiff.Changed {
				if entry.ComponentID == ic.ComponentID && len(entry.ImpactReasons) > 0 {
					res.Category = InstallAlignmentCategoryChangesNotApplied
					res.Reason = "Component dependency changes haven’t been applied."
					break
				}
			}
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
	case st.compositeErr != nil || st.compositeDiff == nil:
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "unable to compare the applied sandbox config with the selected config"
	case st.compositeDiff.SandboxChanged:
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
	case st.compositeErr != nil || st.compositeDiff == nil:
		res.Category = InstallAlignmentCategoryUnknown
		res.Reason = "unable to compare the applied stack config with the selected config"
	case st.compositeDiff.StackChanged:
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
	open, err := s.openDeployments(ctx, st)
	if err != nil {
		return nil, err
	}

	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		res.LastAttempt = ev.lastAttempts[res.ID]
		res.OpenDeployment = open[res.ID]
		res.Plan = ev.plans[res.ID]
		if res.Plan != nil && res.appliedAt != nil && res.Plan.CreatedAt.Before(*res.appliedAt) {
			res.Plan = nil
		}
		res.applyDriftEvidence(ev)
	}
	if err := s.explainInstallAlignment(ctx, st, alignment); err != nil {
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
	lastAttempts  map[string]*InstallAlignmentAttempt
	plans         map[string]*InstallAlignmentPlan
	drifted       map[string]bool
	inputsAligned map[string]bool
}

func (s *Helpers) loadAlignmentEvidence(ctx context.Context, st *installAlignmentState) (*alignmentEvidence, error) {
	ev := &alignmentEvidence{
		lastAttempts:  map[string]*InstallAlignmentAttempt{},
		plans:         map[string]*InstallAlignmentPlan{},
		drifted:       map[string]bool{},
		inputsAligned: map[string]bool{},
	}
	attemptWorkflowIDs := []string{}

	icIDs := make([]string, 0, len(st.components))
	targetByResource := map[string]*installhelpers.ComponentTarget{}
	for i := range st.components {
		icIDs = append(icIDs, st.components[i].ID)
		targetByResource[st.components[i].ID] = st.targets[st.components[i].ComponentID]
	}
	if len(icIDs) > 0 {
		lastDeploys := st.latestDeploys
		for icID, d := range lastDeploys {
			attempt := &InstallAlignmentAttempt{
				ID:        d.ID,
				Status:    string(d.Status),
				CreatedAt: d.CreatedAt,
			}
			if d.InstallWorkflowID != nil {
				attempt.WorkflowID = *d.InstallWorkflowID
				attemptWorkflowIDs = append(attemptWorkflowIDs, *d.InstallWorkflowID)
			}
			ev.lastAttempts[icID] = attempt
			applied := st.appliedDeploys[icID]
			target := targetByResource[icID]
			if target != nil && target.Build != nil && target.Build.ID == d.ComponentBuildID &&
				d.Status == app.InstallDeployStatusAutoSkipped && d.PlannedAt != nil &&
				d.StatusV2.Metadata["policy_action"] != "skip_apply" &&
				(applied == nil || applied.AppliedAt == nil || !d.PlannedAt.Before(*applied.AppliedAt)) {
				cfg := &target.Build.ComponentConfigConnection
				inputs := st.desiredInputs(cfg.AppConfigID)
				ev.inputsAligned[icID] = d.PlannedInputsDigest == installhelpers.InputsDigest(installhelpers.ComponentInputNames(cfg, inputs), inputs)
			}
		}

		scans, err := s.latestInstallDeploysByComponent(ctx, icIDs, true)
		if err != nil {
			return nil, err
		}
		for icID, scan := range scans {
			applied := st.appliedDeploys[icID]
			if applied == nil || scan.ComponentBuildID != applied.ComponentBuildID {
				continue
			}
			if applied.AppliedAt != nil && scan.CreatedAt.Before(*applied.AppliedAt) {
				continue
			}
			if last := lastDeploys[icID]; last != nil && scan.CreatedAt.Before(last.CreatedAt) {
				continue
			}
			ev.drifted[icID] = scan.Status == app.InstallDeployStatus(app.WorkflowStepDrifted)
		}

		plans, err := s.latestPlansByOwner(ctx, `
			SELECT DISTINCT ON (d.install_component_id) a.id AS approval_id, d.install_component_id AS resource_id
			FROM install_workflow_step_approvals a
			JOIN install_deploys d ON d.id = a.owner_id AND d.deleted_at = 0
			WHERE a.deleted_at = 0
			  AND a.owner_type = 'install_deploys'
			  AND d.install_component_id IN ?
			ORDER BY d.install_component_id, a.created_at DESC`, icIDs)
		if err != nil {
			return nil, err
		}
		for k, v := range plans {
			ev.plans[k] = v
		}
	}

	if st.sandbox != nil {
		last, err := s.latestSandboxRun(ctx, st.sandbox.ID, false)
		if err != nil {
			return nil, err
		}
		if last != nil {
			attempt := &InstallAlignmentAttempt{
				ID:        last.ID,
				Status:    string(last.Status),
				CreatedAt: last.CreatedAt,
			}
			if last.InstallWorkflowID != nil {
				attempt.WorkflowID = *last.InstallWorkflowID
				attemptWorkflowIDs = append(attemptWorkflowIDs, *last.InstallWorkflowID)
			}
			ev.lastAttempts[st.sandbox.ID] = attempt
			if st.expectedSandbox != nil && last.AppSandboxConfigID == st.expectedSandbox.ID &&
				last.Status == app.SandboxRunAutoSkipped && last.PlannedAt != nil &&
				last.StatusV2.Metadata["policy_action"] != "skip_apply" &&
				(st.appliedSandbox == nil || st.appliedSandbox.AppliedAt == nil || !last.PlannedAt.Before(*st.appliedSandbox.AppliedAt)) {
				ev.inputsAligned[st.sandbox.ID] = last.PlannedInputsDigest == installhelpers.InputsDigest(installhelpers.InputNames(st.expectedSandbox.Refs), st.desiredInputs(""))
			}
		}
		scan, err := s.latestSandboxRun(ctx, st.sandbox.ID, true)
		if err != nil {
			return nil, err
		}
		if scan != nil && st.appliedSandbox != nil && sandboxScanMatchesApplied(st, scan) &&
			(st.appliedSandbox.AppliedAt == nil || !scan.CreatedAt.Before(*st.appliedSandbox.AppliedAt)) &&
			(last == nil || !scan.CreatedAt.Before(last.CreatedAt)) {
			ev.drifted[st.sandbox.ID] = scan.Status == app.SandboxRunStatus(app.WorkflowStepDrifted)
		}

		plans, err := s.latestPlansByOwner(ctx, `
			SELECT DISTINCT ON (r.install_sandbox_id) a.id AS approval_id, r.install_sandbox_id AS resource_id
			FROM install_workflow_step_approvals a
			JOIN install_sandbox_runs r ON r.id = a.owner_id AND r.deleted_at = 0
			WHERE a.deleted_at = 0
			  AND a.owner_type = 'install_sandbox_runs'
			  AND r.install_sandbox_id IN ?
			ORDER BY r.install_sandbox_id, a.created_at DESC`, []string{st.sandbox.ID})
		if err != nil {
			return nil, err
		}
		for k, v := range plans {
			ev.plans[k] = v
		}
	}

	if st.stack != nil {
		var version app.InstallStackVersion
		err := s.db.WithContext(ctx).
			Select("id", "status", "created_at").
			Where(app.InstallStackVersion{InstallStackID: st.stack.ID}).
			Order("created_at DESC").
			First(&version).Error
		if err != nil && !isNotFound(err) {
			return nil, fmt.Errorf("unable to get latest install stack version: %w", err)
		}
		if err == nil {
			attempt := &InstallAlignmentAttempt{
				ID:        version.ID,
				Status:    string(version.Status.Status),
				CreatedAt: version.CreatedAt,
			}
			var step app.WorkflowStep
			err = s.db.WithContext(ctx).
				Select("id", "install_workflow_id").
				Where(app.WorkflowStep{StepTargetID: version.ID, StepTargetType: string(app.WorkflowStepTargetTypeInstallStackVersions)}).
				Order("created_at DESC").
				First(&step).Error
			if err != nil && !isNotFound(err) {
				return nil, fmt.Errorf("unable to get install stack version workflow: %w", err)
			}
			if err == nil && step.InstallWorkflowID != "" {
				attempt.WorkflowID = step.InstallWorkflowID
				attemptWorkflowIDs = append(attemptWorkflowIDs, step.InstallWorkflowID)
			}
			ev.lastAttempts[st.stack.ID] = attempt
		}
	}

	if err := s.addInputWorkflowAttempts(ctx, st, ev); err != nil {
		return nil, err
	}
	if len(attemptWorkflowIDs) > 0 {
		var workflows []app.Workflow
		if err := s.db.WithContext(ctx).Select("id", "type").
			Where("id IN ?", attemptWorkflowIDs).Find(&workflows).Error; err != nil {
			return nil, fmt.Errorf("unable to get attempt workflow types: %w", err)
		}
		types := make(map[string]app.WorkflowType, len(workflows))
		for _, wf := range workflows {
			types[wf.ID] = wf.Type
		}
		for _, attempt := range ev.lastAttempts {
			if t := types[attempt.WorkflowID]; t != "" {
				attempt.Type = t
			}
		}
		var versions []app.InstallAppConfigVersion
		if err := s.db.WithContext(ctx).
			Preload("AppBranchRun").
			Preload("AppBranchRun.VCSConnectionCommit").
			Where(app.InstallAppConfigVersion{InstallID: st.install.ID, OrgID: st.install.OrgID}).
			Where("workflow_id IN ?", attemptWorkflowIDs).
			Where("app_branch_run_id IS NOT NULL").
			Find(&versions).Error; err != nil {
			return nil, fmt.Errorf("unable to get install app config versions for attempts: %w", err)
		}
		byWorkflow := map[string]*app.AppBranchRun{}
		for i := range versions {
			if versions[i].WorkflowID != nil && versions[i].AppBranchRun.ID != "" {
				byWorkflow[*versions[i].WorkflowID] = &versions[i].AppBranchRun
			}
		}
		for _, attempt := range ev.lastAttempts {
			if run := byWorkflow[attempt.WorkflowID]; run != nil {
				attempt.BranchRun = CommitFromRun(run)
			}
		}
	}

	return ev, nil
}

func sandboxScanMatchesApplied(st *installAlignmentState, scan *app.InstallSandboxRun) bool {
	if scan.AppSandboxConfigID == st.appliedSandbox.AppSandboxConfigID {
		return true
	}
	if st.expectedSandbox == nil || scan.AppSandboxConfigID != st.expectedSandbox.ID {
		return false
	}
	if st.appliedSandbox.AppSandboxConfigID == st.expectedSandbox.ID {
		return true
	}
	return st.compositeErr == nil && st.compositeDiff != nil &&
		!st.compositeDiff.SandboxChanged
}

func (s *Helpers) latestInstallDeploysByComponent(ctx context.Context, icIDs []string, planOnly bool) (map[string]*app.InstallDeploy, error) {
	if len(icIDs) == 0 {
		return map[string]*app.InstallDeploy{}, nil
	}
	join := "LEFT JOIN install_workflows w ON w.id = d.install_workflow_id AND w.deleted_at = 0"
	filter := "(w.id IS NULL OR w.plan_only = false)"
	if planOnly {
		join = "JOIN install_workflows w ON w.id = d.install_workflow_id AND w.deleted_at = 0"
		filter = "w.plan_only = true"
	}
	var ids []string
	if err := s.db.WithContext(ctx).Raw(`
		SELECT latest.id
		FROM unnest(?::text[]) AS component_ids(id)
		CROSS JOIN LATERAL (
			SELECT d.id FROM install_deploys d
			`+join+`
			WHERE d.deleted_at = 0 AND d.install_component_id = component_ids.id
			  AND `+filter+`
			ORDER BY d.created_at DESC
			LIMIT 1
		) latest`, pq.Array(icIDs)).
		Scan(&ids).Error; err != nil {
		return nil, fmt.Errorf("unable to get latest install deploys: %w", err)
	}
	out := map[string]*app.InstallDeploy{}
	if len(ids) == 0 {
		return out, nil
	}
	var deploys []app.InstallDeploy
	if err := s.db.WithContext(ctx).Preload("InstallWorkflow", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "type")
	}).Where("id IN ?", ids).Find(&deploys).Error; err != nil {
		return nil, fmt.Errorf("unable to load install deploys: %w", err)
	}
	for i := range deploys {
		out[deploys[i].InstallComponentID] = &deploys[i]
	}
	return out, nil
}

func (s *Helpers) latestSandboxRun(ctx context.Context, sandboxID string, planOnly bool) (*app.InstallSandboxRun, error) {
	query := s.db.WithContext(ctx).
		Where(app.InstallSandboxRun{InstallSandboxID: &sandboxID})
	if planOnly {
		query = query.
			Joins("JOIN install_workflows w ON w.id = install_sandbox_runs.install_workflow_id AND w.deleted_at = 0").
			Where("w.plan_only = ?", true)
	} else {
		query = query.
			Joins("LEFT JOIN install_workflows w ON w.id = install_sandbox_runs.install_workflow_id AND w.deleted_at = 0").
			Where("(w.id IS NULL OR w.plan_only = ?)", false)
	}
	var run app.InstallSandboxRun
	err := query.Order("install_sandbox_runs.created_at DESC").First(&run).Error
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get latest install sandbox run: %w", err)
	}
	return &run, nil
}

func (s *Helpers) latestPlansByOwner(ctx context.Context, query string, ids []string) (map[string]*InstallAlignmentPlan, error) {
	var rows []struct {
		ApprovalID string
		ResourceID string
	}
	if err := s.db.WithContext(ctx).Raw(query, ids).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to get latest plans: %w", err)
	}
	out := map[string]*InstallAlignmentPlan{}
	if len(rows) == 0 {
		return out, nil
	}
	approvalIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		approvalIDs = append(approvalIDs, row.ApprovalID)
	}
	approvals, err := s.loadAlignmentApprovals(ctx, approvalIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if a := approvals[row.ApprovalID]; a != nil {
			out[row.ResourceID] = alignmentPlanFromApproval(a)
		}
	}
	return out, nil
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

var alignmentOpenStepStatuses = map[app.Status]struct{}{
	"":                                     {},
	app.StatusPending:                      {},
	app.StatusQueued:                       {},
	app.StatusInProgress:                   {},
	app.StatusRetrying:                     {},
	app.StatusPlanning:                     {},
	app.StatusApplying:                     {},
	app.StatusCheckPlan:                    {},
	app.StatusFailedPendingRetry:           {},
	app.AwaitingApproval:                   {},
	app.WorkflowStepApprovalStatusApproved: {},
}

func workflowAwaitingRetry(status app.Status) bool {
	return status == app.StatusFailedPendingRetry || status == app.StatusRetrying
}

func workflowCoversKind(wf *app.Workflow, kind InstallAlignmentResourceKind) bool {
	componentKind := kind == InstallAlignmentResourceKindComponent || kind == InstallAlignmentResourceKindImage
	switch wf.Type {
	case app.WorkflowTypeProvision,
		app.WorkflowTypeReprovision,
		app.WorkflowTypeAppBranchConfigUpdate,
		app.WorkflowTypeAppBranchesRun,
		app.WorkflowTypeAppBranchesConfigRepoUpdate,
		app.WorkflowTypeAppBranchesComponentRepoUpdate,
		app.WorkflowTypeAppInstallSync:
		return true
	case app.WorkflowTypeDeployComponents, app.WorkflowTypeInputUpdate:
		return componentKind
	case app.WorkflowTypeTeardownComponents:
		return kind == InstallAlignmentResourceKindComponent
	case app.WorkflowTypeReprovisionSandbox:
		if kind == InstallAlignmentResourceKindSandbox {
			return true
		}
		skip := wf.Metadata["skip_components"]
		return componentKind && (skip == nil || *skip != "true")
	case app.WorkflowTypeReprovisionStack:
		return kind == InstallAlignmentResourceKindStack
	default:
		return false
	}
}

func (s *Helpers) openDeployments(ctx context.Context, st *installAlignmentState) (map[string]*InstallAlignmentOpenDeployment, error) {
	out := map[string]*InstallAlignmentOpenDeployment{}

	var workflows []app.Workflow
	if err := s.db.WithContext(ctx).
		Select("id", "type", "status", "metadata", "created_at", "plan_only", "finished_at").
		Where(app.Workflow{OwnerID: st.install.ID, OrgID: st.install.OrgID, OwnerType: "installs"}).
		Where("plan_only = ?", false).
		Where("(status->>'status' IN ? OR status->>'status' = ? OR status->>'status' IS NULL)", installhelpers.InstallDeploymentActiveStatuses, app.StatusFailedPendingRetry).
		Order("created_at ASC").
		Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("unable to get active install workflows: %w", err)
	}
	if len(workflows) == 0 {
		return out, nil
	}
	workflowIDs := make([]string, 0, len(workflows))
	for i := range workflows {
		workflowIDs = append(workflowIDs, workflows[i].ID)
	}

	var steps []alignmentStep
	if err := s.db.WithContext(ctx).
		Model(&app.WorkflowStep{}).
		Select("id", "install_workflow_id", "status", "step_target_type", "step_target_id", "metadata", "retried",
			"queue_signal->>'type' AS signal_type",
			"queue_signal->'data'->>'ComponentID' AS signal_component_id").
		Where("install_workflow_id IN ?", workflowIDs).
		Scan(&steps).Error; err != nil {
		return nil, fmt.Errorf("unable to get active workflow steps: %w", err)
	}
	stepsByWorkflow := map[string][]*alignmentStep{}
	for i := range steps {
		stepsByWorkflow[steps[i].InstallWorkflowID] = append(stepsByWorkflow[steps[i].InstallWorkflowID], &steps[i])
	}

	var deploys []app.InstallDeploy
	if err := s.db.WithContext(ctx).
		Select("id", "install_component_id", "install_workflow_id").
		Where("install_workflow_id IN ?", workflowIDs).
		Find(&deploys).Error; err != nil {
		return nil, fmt.Errorf("unable to get active workflow deploys: %w", err)
	}
	deployComponent := map[string]string{}
	workflowComponents := map[string]map[string]struct{}{}
	for i := range deploys {
		deployComponent[deploys[i].ID] = deploys[i].InstallComponentID
		if deploys[i].InstallWorkflowID == nil {
			continue
		}
		wfID := *deploys[i].InstallWorkflowID
		if workflowComponents[wfID] == nil {
			workflowComponents[wfID] = map[string]struct{}{}
		}
		workflowComponents[wfID][deploys[i].InstallComponentID] = struct{}{}
	}

	targets := make([]alignmentTarget, 0, len(st.components)+2)
	if st.stack != nil {
		targets = append(targets, alignmentTarget{kind: InstallAlignmentResourceKindStack, stack: st.stack})
	}
	if st.sandbox != nil {
		targets = append(targets, alignmentTarget{kind: InstallAlignmentResourceKindSandbox, sandbox: st.sandbox})
	}
	for i := range st.components {
		kind := InstallAlignmentResourceKindComponent
		if st.components[i].Component.Type.IsImage() {
			kind = InstallAlignmentResourceKindImage
		}
		targets = append(targets, alignmentTarget{kind: kind, installComponent: &st.components[i]})
	}

	for i := range workflows {
		wf := &workflows[i]
		wfSteps := stepsByWorkflow[wf.ID]
		var inputTargets map[string]bool
		if slices.Contains(alignmentInputWorkflowTypes, wf.Type) {
			inputTargets = st.inputWorkflowTargets(wf)
		}
		for _, target := range targets {
			id := target.id()
			if _, ok := out[id]; ok {
				continue
			}
			isOpen := workflowOpenForTarget(wf, wfSteps, target, deployComponent, workflowComponents[wf.ID])
			if len(wfSteps) == 0 && inputTargets != nil {
				isOpen = inputTargets[id]
			}
			if !isOpen {
				continue
			}
			status := string(wf.Status.Status)
			if status == "" {
				status = string(app.StatusPending)
			}
			out[id] = &InstallAlignmentOpenDeployment{
				ID:        wf.ID,
				Status:    status,
				CreatedAt: wf.CreatedAt,
				Type:      wf.Type,
			}
		}
	}
	return out, nil
}

func (t alignmentTarget) id() string {
	switch {
	case t.installComponent != nil:
		return t.installComponent.ID
	case t.sandbox != nil:
		return t.sandbox.ID
	case t.stack != nil:
		return t.stack.ID
	default:
		return ""
	}
}

type alignmentStep struct {
	ID                string
	InstallWorkflowID string
	Status            app.CompositeStatus
	StepTargetType    string
	StepTargetID      string
	Metadata          pgtype.Hstore `gorm:"type:hstore"`
	Retried           bool
	SignalType        string
	SignalComponentID string
	SignalBuildID     string
	ResourceID        string
}

var (
	alignmentSandboxSignalTypes = map[string]struct{}{
		string(provisionsandboxplan.SignalType):        {},
		string(provisionsandboxapplyplan.SignalType):   {},
		string(reprovisionsandboxplan.SignalType):      {},
		string(reprovisionsandboxapplyplan.SignalType): {},
		string(deprovisionsandboxplan.SignalType):      {},
		string(deprovisionsandboxapplyplan.SignalType): {},
	}
	alignmentStackSignalTypes = map[string]struct{}{
		string(generateinstallstackversion.SignalType): {},
		string(awaitinstallstackversionrun.SignalType): {},
		string(stackrun.SignalType):                    {},
	}
)

func workflowOpenForTarget(
	wf *app.Workflow,
	steps []*alignmentStep,
	target alignmentTarget,
	deployComponent map[string]string,
	workflowComponents map[string]struct{},
) bool {
	if len(steps) == 0 {
		if target.installComponent != nil {
			if _, ok := workflowComponents[target.installComponent.ID]; ok {
				return true
			}
			if wf.Type == app.WorkflowTypeTeardownComponent && target.kind == InstallAlignmentResourceKindComponent {
				id := wf.Metadata["component_id"]
				return id != nil && *id == target.installComponent.ComponentID
			}
		}
		return workflowCoversKind(wf, target.kind)
	}
	for _, step := range steps {
		if step.Retried || !stepRelatesToTarget(step, target, deployComponent) {
			continue
		}
		status := step.Status.Status
		if _, ok := alignmentOpenStepStatuses[status]; ok {
			return true
		}
		if status == app.StatusError && workflowAwaitingRetry(wf.Status.Status) {
			return true
		}
	}
	return false
}

func stepRelatesToTarget(step *alignmentStep, target alignmentTarget, deployComponent map[string]string) bool {
	switch target.kind {
	case InstallAlignmentResourceKindStack:
		_, ok := alignmentStackSignalTypes[step.SignalType]
		return ok || step.StepTargetType == string(app.WorkflowStepTargetTypeInstallStackVersions)
	case InstallAlignmentResourceKindSandbox:
		_, ok := alignmentSandboxSignalTypes[step.SignalType]
		return ok || step.StepTargetType == string(app.WorkflowStepTargetTypeInstallSandboxRuns)
	default:
		ic := target.installComponent
		if step.SignalComponentID != "" && step.SignalComponentID == ic.ComponentID {
			return true
		}
		if step.StepTargetType == string(app.WorkflowStepTargetTypeInstallDeploys) && step.StepTargetID != "" {
			if deployComponent[step.StepTargetID] == ic.ID {
				return true
			}
		}
		if name, ok := step.Metadata["component_name"]; ok && name != nil && *name == ic.Component.Name {
			return true
		}
		return false
	}
}
