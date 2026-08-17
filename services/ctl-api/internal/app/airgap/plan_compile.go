package airgap

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/render"
	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
	statepkg "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	planpkg "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

const compilePlanEnvelopeSource = "ctl-api airgap-bundle compile"

var (
	stackOutputReference          = regexp.MustCompile(`\.nuon\.install_stack\.outputs\.([A-Za-z0-9_-]+)`)
	sandboxOutputReference        = regexp.MustCompile(`\.nuon\.sandbox\.outputs\.([A-Za-z0-9_.-]+)`)
	installSandboxOutputReference = regexp.MustCompile(`\.nuon\.install\.sandbox\.outputs\.([A-Za-z0-9_.-]+)`)
	componentOutputRef            = regexp.MustCompile(`\.nuon\.components\.([A-Za-z0-9_-]+)\.outputs\.([A-Za-z0-9_.-]+)`)
)

func CompilePlanEnvelope(ctx context.Context, db *gorm.DB, v *validator.Validate, orgID, appID string, cfg *app.AppConfig, sandboxBuildID string, componentBuildIDs map[string]string, runbooks []runnerairgap.RunbookTemplate, report *QualificationReport) (*runnerairgap.Envelope, error) {
	if cfg == nil {
		return nil, fmt.Errorf("app config is required")
	}
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("serialize app config for template validation: %w", err)
	}
	connections, err := compileConnections(ctx, db, cfg)
	if err != nil {
		return nil, err
	}
	componentRefs := extractComponentOutputRefs(cfgJSON)
	if err := validateComponentOutputRefs(componentRefs, connections, report); err != nil {
		return nil, err
	}
	appConfigJSON, err := exportAppConfigJSON(ctx, db, cfg.ID)
	if err != nil {
		return nil, err
	}
	componentRefs = extractComponentOutputRefs(cfgJSON, appConfigJSON)
	if err := validateComponentOutputRefs(componentRefs, connections, report); err != nil {
		return nil, err
	}
	componentRefs = mergeComponentRefs(componentRefs, contextClusterRefs(cfg, connections))

	inputs, err := exportInputSpecs(ctx, db, cfg.ID)
	if err != nil {
		return nil, err
	}
	for i := range inputs {
		inputs[i].Bindable = true
	}
	installID := VirtualInstallID(appID)
	st, stateData, err := compileState(orgID, appID, installID, appConfigJSON, inputs, componentRefs, needsSandboxCluster(connections, cfgJSON, appConfigJSON), roleConfigNames(cfg.BreakGlassConfig.Roles), roleConfigNames(cfg.PermissionsConfig.CustomRoles), cfg.StackConfig.Type)
	if err != nil {
		return nil, err
	}
	install := compileInstall(orgID, appID, cfg.ID, installID)
	stack := compileStack(cfg, installID)
	role := compileRoleSelection(cfg, stack)
	planner := planpkg.NewPlanner(v)
	logger := zap.NewNop()

	var sandboxBuild app.AppSandboxBuild
	if err := db.WithContext(ctx).Where(app.AppSandboxBuild{ID: sandboxBuildID, OrgID: orgID, AppID: appID}).First(&sandboxBuild).Error; err != nil {
		return nil, fmt.Errorf("load pinned sandbox build %s: %w", sandboxBuildID, err)
	}

	run := &app.InstallSandboxRun{ID: "sandbox-plan", InstallID: installID}
	sandboxPlan, _, err := planner.RenderSandboxRunPlan(logger, &planpkg.RenderSandboxRunPlanInput{
		Install: install, Stack: stack, AppCfg: cfg, Run: run,
		ResolveState: func() (*statepkg.State, map[string]any, error) { return st, stateData, nil },
		ResolveSource: func() (*plantypes.GitSource, *plantypes.OCISource, error) {
			return nil, &plantypes.OCISource{Registry: compileOrgRegistry(orgID, appID), Tag: sandboxBuild.ID}, nil
		},
		HasUpdatePlansFeature: func() (bool, error) { return false, nil },
	})
	if err != nil {
		return nil, fmt.Errorf("render sandbox plan: %w", err)
	}
	sandboxComposite, err := compositeJSON(&plantypes.CompositePlan{SandboxRunPlan: sandboxPlan})
	if err != nil {
		return nil, err
	}
	sandboxType := "sandbox-terraform"
	if cfg.SandboxConfig.Type == "pulumi" {
		sandboxType = "sandbox-pulumi"
	}
	if refs := componentTokenRefs(sandboxComposite, componentRefs); len(refs) > 0 {
		return nil, fmt.Errorf("sandbox config references component outputs, which do not exist before components apply: %s", strings.Join(componentRefStrings(refs), ", "))
	}
	steps := []runnerairgap.Step{
		{ID: "sandbox-plan", Name: "sandbox create-apply-plan", JobType: sandboxType, JobOperation: "create-apply-plan", JobGroup: "sandbox", CompositePlan: sandboxComposite},
		{ID: "sandbox-apply", Name: "sandbox apply-plan", JobType: sandboxType, JobOperation: "apply-plan", JobGroup: "sandbox", DependsOn: []string{"sandbox-plan"}, PlanFromStep: "sandbox-plan", CompositePlan: sandboxComposite},
	}

	order, err := compileComponentOrder(connections)
	if err != nil {
		return nil, err
	}
	installComponents := make([]app.InstallComponent, 0, len(order))
	previous := "sandbox-apply"
	applied := map[string]bool{}
	componentApplySteps := map[string]string{}
	for _, connection := range order {
		buildID := componentBuildIDs[connection.ID]
		if buildID == "" {
			return nil, fmt.Errorf("component config connection %s has no pinned build", connection.ID)
		}
		build, err := compileComponentBuild(ctx, db, orgID, buildID, connection)
		if err != nil {
			return nil, err
		}
		installComponentID := virtualID("vic", appID+":"+connection.ComponentID)
		deployID := virtualID("vdep", appID+":"+connection.ComponentID)
		ic := app.InstallComponent{ID: installComponentID, InstallID: installID, ComponentID: connection.ComponentID, Component: connection.Component}
		installComponents = append(installComponents, ic)
		deploy := &app.InstallDeploy{ID: deployID, OrgID: orgID, InstallID: installID, InstallComponentID: installComponentID, InstallComponent: ic, ComponentBuildID: build.ID, ComponentBuild: *build, ComponentID: connection.ComponentID, ComponentName: connection.ComponentName, Type: app.InstallDeployTypeApply}
		dst, err := compileInstallRegistry(planner, logger, deploy, stack, stateData, role)
		if err != nil {
			return nil, fmt.Errorf("render install registry for component %s: %w", connection.ComponentName, err)
		}
		deploy.OCIArtifact = app.OCIArtifact{Repository: dst.Repository, Tag: deploy.ID, Digest: build.SourceDigest}
		syncPlan, err := planner.RenderSyncOCIPlan(logger, &planpkg.RenderSyncOCIPlanInput{Deploy: deploy, CompBuild: build, Install: install, SrcCfg: compileOrgRegistry(orgID, appID), DstCfg: dst})
		if err != nil {
			return nil, fmt.Errorf("render sync plan for component %s: %w", connection.ComponentName, err)
		}
		syncComposite, err := compositeJSON(&plantypes.CompositePlan{SyncOCIPlan: syncPlan})
		if err != nil {
			return nil, err
		}
		if syncComposite, err = normalizeComponentTokenPadding(syncComposite, componentRefs); err != nil {
			return nil, err
		}
		syncID := "sync-" + connection.ComponentName
		steps = append(steps, runnerairgap.Step{ID: syncID, Name: connection.ComponentName + " sync", JobType: "oci-sync", JobOperation: "exec", JobGroup: "sync", DependsOn: []string{previous}, CompositePlan: syncComposite})

		deployPlan, jobType, err := compileDeployPlan(planner, logger, cfg, stack, st, stateData, deploy, build, dst, role)
		if err != nil {
			return nil, fmt.Errorf("render deploy plan for component %s: %w", connection.ComponentName, err)
		}
		deployComposite, err := compositeJSON(&plantypes.CompositePlan{DeployPlan: deployPlan})
		if err != nil {
			return nil, err
		}
		if deployComposite, err = normalizeComponentTokenPadding(deployComposite, componentRefs); err != nil {
			return nil, err
		}
		for _, ref := range append(componentTokenRefs(syncComposite, componentRefs), componentTokenRefs(deployComposite, componentRefs)...) {
			if !applied[ref.Component] {
				return nil, fmt.Errorf("component %s consumes %s before component %s applies; declare the dependency so it orders later", connection.ComponentName, ref.String(), ref.Component)
			}
		}
		planID := "deploy-" + connection.ComponentName + "-plan"
		applyID := "deploy-" + connection.ComponentName + "-apply"
		steps = append(steps,
			runnerairgap.Step{ID: planID, Name: connection.ComponentName + " create-apply-plan", JobType: jobType, JobOperation: "create-apply-plan", JobGroup: "deploy", DependsOn: []string{syncID}, CompositePlan: deployComposite},
			runnerairgap.Step{ID: applyID, Name: connection.ComponentName + " apply-plan", JobType: jobType, JobOperation: "apply-plan", JobGroup: "deploy", DependsOn: []string{planID}, PlanFromStep: planID, CompositePlan: deployComposite},
		)
		previous = applyID
		applied[connection.ComponentName] = true
		componentApplySteps[connection.ComponentName] = applyID
	}

	bindings := make([]runnerairgap.OutputBinding, 0, len(componentRefs))
	for _, ref := range componentRefs {
		stepID := componentApplySteps[ref.Component]
		if stepID == "" {
			return nil, fmt.Errorf("component output reference %s has no producing apply step", ref.String())
		}
		bindings = append(bindings, runnerairgap.OutputBinding{Token: ref.token(), ComponentName: ref.Component, StepID: stepID, OutputPath: ref.Path})
	}

	actions, err := compileActionTemplates(cfg, planner, logger, installID, stateData, stack, role, report)
	if err != nil {
		return nil, err
	}
	drift, err := compileDriftTemplates(steps, componentSpecs(installComponents, connections))
	if err != nil {
		return nil, err
	}
	envelope := &runnerairgap.Envelope{Version: "v0", OrgID: orgID, AppID: appID, InstallID: installID, CreatedAt: time.Now().UTC(), Source: compilePlanEnvelopeSource, AppConfig: appConfigJSON, Inputs: inputs, ForceDefaultCloudAuth: true, Components: componentSpecs(installComponents, connections), Steps: steps, Actions: actions, Drift: drift, Runbooks: runbooks, OutputBindings: bindings}
	if err := envelope.Validate(); err != nil {
		return nil, fmt.Errorf("compiled plan envelope is invalid: %w", err)
	}
	return envelope, nil
}

// compileInstall builds the synthetic install used for zero-install plan
// rendering. The sandbox terraform workspace ID must be non-empty: the
// rendered sandbox plan bakes it into the http-backend query string, and the
// offline runner's loopback backend rejects requests without a workspace_id.
func compileInstall(orgID, appID, appConfigID, installID string) *app.Install {
	install := &app.Install{ID: installID, OrgID: orgID, AppID: appID, AppConfigID: appConfigID, Name: "airgap"}
	install.InstallSandbox.TerraformWorkspace.ID = virtualID("vtfw", installID)
	return install
}

func virtualID(prefix, seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return prefix + hex.EncodeToString(sum[:])[:16]
}

func uniqueMatches(re *regexp.Regexp, data []byte) []string {
	seen := map[string]bool{}
	for _, match := range re.FindAllSubmatch(data, -1) {
		seen[string(match[1])] = true
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func addQualificationViolation(report *QualificationReport, code, member, message string) {
	if report == nil {
		return
	}
	report.Violations = append(report.Violations, Finding{Code: code, Member: member, Message: message})
	finish(report)
}

func addQualificationWarning(report *QualificationReport, code, member, message string) {
	if report == nil {
		return
	}
	report.Warnings = append(report.Warnings, Finding{Code: code, Member: member, Message: message})
	finish(report)
}

func firstGitSourcedActionStep(actionCfg app.ActionWorkflowConfig) string {
	for _, stepCfg := range actionCfg.Steps {
		if stepCfg.PublicGitVCSConfig != nil || stepCfg.ConnectedGithubVCSConfig != nil {
			return stepCfg.Name
		}
	}
	return ""
}

// The readme is customer-facing documentation, not a plan input: component
// output references inside it are rendered after components apply, so they
// must not disqualify the config.
func withoutDocFields(data []byte) []byte {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return data
	}
	delete(fields, "readme")
	stripped, err := json.Marshal(fields)
	if err != nil {
		return data
	}
	return stripped
}

// The planner injects builtin sandbox variables referencing well-known stack
// output keys (see plan_sandbox_run.go), so they never appear in the app
// config and must always be seeded as placeholders.
var builtinStackOutputKeys = []string{
	"vpc_id", "region", "provision_iam_role_arn", "deprovision_iam_role_arn", "maintenance_iam_role_arn",
	"project_id", "network_name", "private_subnet_name",
	"subscription_id", "subscription_tenant_id", "resource_group_location",
}

func compileState(orgID, appID, installID string, raw []byte, inputs []runnerairgap.InputSpec, componentRefs []componentOutputRefKey, seedSandboxCluster bool, breakGlassRoles, customRoles []string, stackType app.StackType) (*statepkg.State, map[string]any, error) {
	stackOutputs, sandboxOutputs := map[string]any{}, map[string]any{}
	for _, key := range builtinStackOutputKeys {
		stackOutputs[key] = "__NUON_AIRGAP_STACK_" + key + "__"
	}
	for _, key := range uniqueMatches(stackOutputReference, raw) {
		stackOutputs[key] = "__NUON_AIRGAP_STACK_" + key + "__"
	}
	sandboxPaths := map[string]bool{}
	for _, path := range uniqueMatches(sandboxOutputReference, raw) {
		sandboxPaths[strings.Trim(path, ".")] = true
	}
	for _, path := range uniqueMatches(installSandboxOutputReference, raw) {
		sandboxPaths[strings.Trim(path, ".")] = true
	}
	for _, path := range sortedKeys(sandboxPaths) {
		token := "__NUON_AIRGAP_SANDBOX_" + strings.ReplaceAll(path, ".", "_") + "__"
		if err := seedPlaceholderPath(sandboxOutputs, path, token); err != nil {
			return nil, nil, fmt.Errorf("sandbox output references: %w", err)
		}
	}
	sandboxOutputs["ecr"] = map[string]any{"repository_url": "__NUON_AIRGAP_SANDBOX_ecr_repository_url__", "registry_url": "__NUON_AIRGAP_SANDBOX_ecr_registry_url__"}
	// Cluster outputs are seeded only when something consumes them: a seeded
	// cluster makes ResolveKubernetesContextFromData treat the sandbox as the
	// implicit kubernetes default, and cluster-less sandboxes (Lambda, ECS)
	// must resolve to no ClusterInfo so deploys don't demand kube auth.
	if seedSandboxCluster {
		sandboxOutputs["cluster"] = map[string]any{"name": "__NUON_AIRGAP_CLUSTER_id__", "endpoint": "__NUON_AIRGAP_CLUSTER_endpoint__", "certificate_authority_data": "__NUON_AIRGAP_CLUSTER_ca_data__"}
	}
	inputValues := map[string]string{}
	for _, input := range inputs {
		inputValues[input.Name] = runnerairgap.InputPlaceholder(input.Name)
	}
	components, err := seedComponentOutputs(componentRefs)
	if err != nil {
		return nil, nil, err
	}
	// Components are seeded only into the render data so placeholder tokens
	// reach the plans that actually reference them. Embedding them in the
	// State carried by every plan would put every token in every composite,
	// breaking both compile-time dependency checks and offline late binding.
	st := &statepkg.State{ID: installID, Name: "airgap", Org: &statepkg.OrgState{ID: orgID, Populated: true}, App: &statepkg.AppState{ID: appID, Populated: true, Variables: map[string]string{}}, Install: &statepkg.InstallState{ID: installID, Name: "airgap", Populated: true, Inputs: inputValues, Sandbox: statepkg.SandboxState{Populated: true, Outputs: sandboxOutputs}}, Inputs: &statepkg.InputsState{Populated: true, Inputs: inputValues}, Sandbox: &statepkg.SandboxState{Populated: true, Outputs: sandboxOutputs}, InstallStack: &statepkg.InstallStackState{Populated: true, Outputs: stackOutputs}, Cloud: compileCloudAccount(stackType)}
	if err := seedRoleARNOutputs(st, stackOutputs, breakGlassRoles, customRoles); err != nil {
		return nil, nil, err
	}
	inner, err := st.AsMap()
	if err != nil {
		return nil, nil, fmt.Errorf("encode compile state: %w", err)
	}
	inner["components"] = components
	return st, map[string]any{"nuon": inner}, nil
}

func roleConfigNames(roles []app.AppAWSIAMRoleConfig) []string {
	names := make([]string, 0, len(roles))
	for _, role := range roles {
		names = append(names, role.Name)
	}
	return names
}

// seedRoleARNOutputs seeds break_glass_role_arns and custom_role_arns as
// name→token maps instead of flat string tokens: templates consume them with
// map operations (len, values, index), which fail on a scalar placeholder.
// Role names render against the compile state so the map keys carry the
// deterministic install ID; the deploy-time install-ID splice rewrites both
// the keys and the stack template's role names identically, so the runner can
// late-bind each token to the real ARN by matching phone-home output keys.
// An app config with no roles seeds empty maps, which keeps guards like
// `gt (len ...) 0` accurate.
func seedRoleARNOutputs(st *statepkg.State, stackOutputs map[string]any, breakGlassRoles, customRoles []string) error {
	inner, err := st.AsMap()
	if err != nil {
		return fmt.Errorf("encode compile state for role rendering: %w", err)
	}
	data := map[string]any{"nuon": inner}
	seed := func(key string, names []string) error {
		arns := map[string]any{}
		for i, name := range names {
			rendered, err := render.RenderTextV2(name, data)
			if err != nil {
				return fmt.Errorf("render %s role name %q: %w", key, name, err)
			}
			if rendered == "" {
				continue
			}
			arns[rendered] = fmt.Sprintf("__NUON_AIRGAP_STACK_%s_%d__", key, i)
		}
		stackOutputs[key] = arns
		return nil
	}
	if err := seed("break_glass_role_arns", breakGlassRoles); err != nil {
		return err
	}
	return seed("custom_role_arns", customRoles)
}

// compileCloudAccount seeds `.nuon.cloud_account` for the stack's cloud so
// templates like `.nuon.cloud_account.aws.region` render instead of nil-panic.
// The values reuse the stack-output placeholder tokens: plans are late-bound
// by the runner from phone-home outputs, and the root stack template's region
// token is substituted with the customer's deploy region during stack
// preparation. Only the stack's own cloud is seeded so `if .nuon.cloud_account.aws`
// style guards keep branching correctly.
func compileCloudAccount(stackType app.StackType) *statepkg.CloudAccount {
	switch stackType {
	case app.StackTypeAzure:
		return &statepkg.CloudAccount{Azure: &statepkg.AzureCloudAccount{Location: "__NUON_AIRGAP_STACK_resource_group_location__"}}
	case app.StackTypeGCP:
		return &statepkg.CloudAccount{GCP: &statepkg.GCPCloudAccount{ProjectID: "__NUON_AIRGAP_STACK_project_id__", Region: "__NUON_AIRGAP_STACK_region__"}}
	default:
		return &statepkg.CloudAccount{AWS: &statepkg.AWSCloudAccount{Region: "__NUON_AIRGAP_STACK_region__"}}
	}
}

func compileStack(cfg *app.AppConfig, installID string) *app.InstallStack {
	outputs := app.InstallStackOutputs{}
	switch cfg.StackConfig.Type {
	case app.StackTypeAzure:
		outputs.AzureStackOutputs = &app.AzureStackOutputs{
			SubscriptionID:              "__NUON_AIRGAP_STACK_subscription_id__",
			SubscriptionTenantID:        "__NUON_AIRGAP_STACK_subscription_tenant_id__",
			MaintenanceIdentityClientID: "__NUON_AIRGAP_STACK_maintenance_identity_client_id__",
		}
	case app.StackTypeGCP:
		outputs.GCPStackOutputs = &app.GCPStackOutputs{
			ProjectID:          "__NUON_AIRGAP_STACK_project_id__",
			Region:             "__NUON_AIRGAP_STACK_region__",
			ProvisionSAEmail:   "__NUON_AIRGAP_STACK_provision_sa_email__",
			DeprovisionSAEmail: "__NUON_AIRGAP_STACK_deprovision_sa_email__",
			MaintenanceSAEmail: "__NUON_AIRGAP_STACK_maintenance_sa_email__",
		}
	default:
		outputs.AWSStackOutputs = &app.AWSStackOutputs{
			AccountID:             "__NUON_AIRGAP_STACK_account_id__",
			Region:                "__NUON_AIRGAP_STACK_region__",
			VPCID:                 "__NUON_AIRGAP_STACK_vpc_id__",
			ProvisionIAMRoleARN:   "__NUON_AIRGAP_STACK_provision_iam_role_arn__",
			DeprovisionIAMRoleARN: "__NUON_AIRGAP_STACK_deprovision_iam_role_arn__",
			MaintenanceIAMRoleARN: "__NUON_AIRGAP_STACK_maintenance_iam_role_arn__",
			RunnerIAMRoleARN:      "__NUON_AIRGAP_STACK_runner_iam_role_arn__",
		}
	}
	return &app.InstallStack{ID: virtualID("vist", installID), InstallID: installID, InstallStackOutputs: outputs}
}

func compileRoleSelection(cfg *app.AppConfig, stack *app.InstallStack) *operationroles.RoleSelection {
	name := cfg.PermissionsConfig.MaintenanceRole.Name
	if name == "" {
		name = "maintenance"
	}
	arn := "__NUON_AIRGAP_STACK_maintenance_role__"
	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		arn = stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		arn = stack.InstallStackOutputs.AzureStackOutputs.MaintenanceIdentityClientID
	case stack.InstallStackOutputs.GCPStackOutputs != nil:
		arn = stack.InstallStackOutputs.GCPStackOutputs.MaintenanceSAEmail
	}
	return &operationroles.RoleSelection{RoleName: name, UnrenderedRoleName: name, RoleARN: arn, Source: operationroles.RoleSelectionSourceDefault}
}

func compileOrgRegistry(orgID, appID string) *configs.OCIRegistryRepository {
	return &configs.OCIRegistryRepository{Plugin: "oci", RegistryType: configs.OCIRegistryTypePrivateOCI, Repository: orgID + "/" + appID, LoginServer: "__NUON_AIRGAP_VENDOR_REGISTRY__"}
}

func compileConnections(ctx context.Context, db *gorm.DB, cfg *app.AppConfig) ([]app.ComponentConfigConnection, error) {
	connections := cfg.ComponentConfigConnections
	if len(connections) != len(cfg.ComponentIDs) {
		var err error
		connections, err = exportComponentConfigConnections(ctx, db, cfg.ID)
		if err != nil {
			return nil, err
		}
	}
	return connections, nil
}

func compileComponentBuild(ctx context.Context, db *gorm.DB, orgID, id string, connection app.ComponentConfigConnection) (*app.ComponentBuild, error) {
	var build app.ComponentBuild
	if err := db.WithContext(ctx).Where(app.ComponentBuild{ID: id, OrgID: orgID}).First(&build).Error; err != nil {
		return nil, fmt.Errorf("load pinned component build %s: %w", id, err)
	}
	if build.ComponentConfigConnectionID != connection.ID {
		return nil, fmt.Errorf("pinned build %s belongs to component config connection %s, not %s", id, build.ComponentConfigConnectionID, connection.ID)
	}
	build.ComponentConfigConnection = connection
	return &build, nil
}

func compileComponentOrder(connections []app.ComponentConfigConnection) ([]app.ComponentConfigConnection, error) {
	byID := map[string]app.ComponentConfigConnection{}
	indegree := map[string]int{}
	children := map[string][]string{}
	for _, c := range connections {
		byID[c.ComponentID] = c
		indegree[c.ComponentID] = 0
	}
	for _, c := range connections {
		for _, dep := range c.ComponentDependencyIDs {
			if _, ok := byID[dep]; !ok {
				continue
			}
			indegree[c.ComponentID]++
			children[dep] = append(children[dep], c.ComponentID)
		}
	}
	result := make([]app.ComponentConfigConnection, 0, len(connections))
	for len(result) < len(connections) {
		ready := []app.ComponentConfigConnection{}
		for id, degree := range indegree {
			if degree == 0 {
				ready = append(ready, byID[id])
			}
		}
		if len(ready) == 0 {
			return nil, fmt.Errorf("component dependency graph contains a cycle")
		}
		sort.Slice(ready, func(i, j int) bool { return ready[i].ComponentName < ready[j].ComponentName })
		for _, c := range ready {
			result = append(result, c)
			delete(indegree, c.ComponentID)
			for _, child := range children[c.ComponentID] {
				indegree[child]--
			}
		}
	}
	return result, nil
}

func compileInstallRegistry(planner *planpkg.Planner, l *zap.Logger, deploy *app.InstallDeploy, stack *app.InstallStack, stateData map[string]any, role *operationroles.RoleSelection) (*configs.OCIRegistryRepository, error) {
	auth, err := planner.AuthForDeploy(l, role, stack, "oci-sync-"+deploy.ID)
	if err != nil {
		return nil, err
	}
	return planner.RenderInstallRegistryRepository(l, &planpkg.RenderInstallRegistryRepositoryInput{InstallDeploy: deploy, Stack: stack, StateData: stateData, CloudAuth: auth})
}

// needsSandboxCluster reports whether the compiled state should present the
// sandbox as emitting kubernetes cluster outputs. Mirrors the connected
// flow's implicit-default contract: helm and kubernetes-manifest components
// without a declared kubernetes_context target the sandbox cluster, and any
// explicit config reference to the sandbox cluster outputs also requires it.
func needsSandboxCluster(connections []app.ComponentConfigConnection, raws ...[]byte) bool {
	for _, connection := range connections {
		switch connection.Type {
		case app.ComponentTypeHelmChart, app.ComponentTypeKubernetesManifest:
			if connection.KubernetesContextName == "" {
				return true
			}
		}
	}
	for _, raw := range raws {
		for _, re := range []*regexp.Regexp{sandboxOutputReference, installSandboxOutputReference} {
			for _, path := range uniqueMatches(re, raw) {
				path = strings.Trim(path, ".")
				if path == "cluster" || strings.HasPrefix(path, "cluster.") {
					return true
				}
			}
		}
	}
	return false
}

// contextClusterRefs returns the component-output references implied by
// declared kubernetes_contexts that components actually use: the peer's
// cluster connection fields must be seeded and late-bound like any other
// cross-component output so context-targeted deploys can resolve offline.
func contextClusterRefs(cfg *app.AppConfig, connections []app.ComponentConfigConnection) []componentOutputRefKey {
	used := map[string]bool{}
	for _, connection := range connections {
		if connection.KubernetesContextName != "" {
			used[connection.KubernetesContextName] = true
		}
	}
	if len(used) == 0 {
		return nil
	}
	fields := []string{"name", "endpoint", "certificate_authority_data"}
	if cfg.StackConfig.Type == app.StackTypeAzure {
		fields = []string{"name", "host", "cluster_ca_certificate"}
	}
	refs := make([]componentOutputRefKey, 0, 3*len(used))
	for _, kctx := range cfg.KubernetesContextsConfig.Contexts {
		if !used[kctx.Name] {
			continue
		}
		for _, field := range fields {
			refs = append(refs, componentOutputRefKey{Component: kctx.SourceComponentName, Path: "cluster." + field})
		}
	}
	return refs
}

func mergeComponentRefs(base, extra []componentOutputRefKey) []componentOutputRefKey {
	seen := map[componentOutputRefKey]bool{}
	merged := make([]componentOutputRefKey, 0, len(base)+len(extra))
	for _, ref := range append(base, extra...) {
		if seen[ref] {
			continue
		}
		seen[ref] = true
		merged = append(merged, ref)
	}
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Component != merged[j].Component {
			return merged[i].Component < merged[j].Component
		}
		return merged[i].Path < merged[j].Path
	})
	return merged
}

func compileDeployPlan(planner *planpkg.Planner, l *zap.Logger, cfg *app.AppConfig, stack *app.InstallStack, st *statepkg.State, stateData map[string]any, deploy *app.InstallDeploy, build *app.ComponentBuild, src *configs.OCIRegistryRepository, role *operationroles.RoleSelection) (*plantypes.DeployPlan, string, error) {
	tag, digest := planpkg.DeploySrcRef(deploy, build)
	result := &plantypes.DeployPlan{Src: src, SrcTag: tag, SrcDigest: digest, AppID: cfg.AppID, AppConfigID: cfg.ID, InstallID: deploy.InstallID, ComponentName: deploy.ComponentName, ComponentID: deploy.ComponentID}
	// ResolveKubernetesContextFromData expects the connected flow's unwrapped
	// state shape (top-level sandbox/install keys), while the compile state is
	// pre-wrapped under "nuon" for template rendering.
	innerState, _ := stateData["nuon"].(map[string]any)
	cluster := func(cloudAuth *planpkg.CloudAuth) (*kube.ClusterInfo, error) {
		return planner.ResolveKubernetesContextFromData(l, build.ComponentConfigConnection.KubernetesContextName, cfg, stack, innerState, cloudAuth)
	}
	jobType := "noop-deploy"
	var err error
	switch build.ComponentConfigConnection.Type {
	case app.ComponentTypeTerraformModule:
		jobType = "terraform-deploy"
		result.TerraformDeployPlan, err = planner.RenderTerraformDeployPlan(l, &planpkg.RenderTerraformDeployPlanInput{Stack: stack, State: st, StateData: stateData, InstallDeploy: deploy, CompBuild: build, WorkspaceID: virtualID("vtfw", deploy.ID), RoleSelection: role, ResolveClusterInfo: cluster})
	case app.ComponentTypeHelmChart:
		jobType = "helm-chart-deploy"
		result.HelmDeployPlan, err = planner.RenderHelmDeployPlan(l, &planpkg.RenderHelmDeployPlanInput{Stack: stack, State: st, StateData: stateData, InstallDeploy: deploy, CompBuild: build, RoleSelection: role, GetHelmChartID: func(string) (string, error) { return virtualID("vhlm", deploy.ID), nil }, ResolveClusterInfo: cluster})
	case app.ComponentTypeKubernetesManifest:
		jobType = "kubernetes-manifest-deploy"
		result.KubernetesManifestDeployPlan, err = planner.RenderKubernetesManifestDeployPlan(l, &planpkg.RenderKubernetesManifestDeployPlanInput{Stack: stack, State: st, StateData: stateData, InstallDeploy: deploy, CompBuild: build, RoleSelection: role, ResolveClusterInfo: cluster})
	case app.ComponentTypePulumi:
		jobType = "pulumi-deploy"
		result.PulumiDeployPlan, err = planner.RenderPulumiDeployPlan(l, &planpkg.RenderPulumiDeployPlanInput{Stack: stack, State: st, StateData: stateData, InstallDeploy: deploy, CompBuild: build, WorkspaceID: virtualID("vpul", deploy.ID), RoleSelection: role, ResolveClusterInfo: cluster, HasUpdatePlansFeature: func() (bool, error) { return false, nil }})
	case app.ComponentTypeDockerBuild, app.ComponentTypeExternalImage:
		result.NoopDeployPlan = &plantypes.NoopDeployPlan{}
	default:
		err = fmt.Errorf("unsupported component type %s", build.ComponentConfigConnection.Type)
	}
	return result, jobType, err
}

func compositeJSON(plan *plantypes.CompositePlan) (json.RawMessage, error) {
	raw, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("marshal composite plan: %w", err)
	}
	return raw, nil
}

// Git-sourced actions are excluded from the bundle with a qualification
// warning rather than failing the publish: their scripts are fetched from Git
// at run time, which the offline runner can never do, but the rest of the app
// (day-1 deploy and inline actions) remains fully bundleable.
func compileActionTemplates(cfg *app.AppConfig, planner *planpkg.Planner, l *zap.Logger, installID string, stateData map[string]any, stack *app.InstallStack, role *operationroles.RoleSelection, report *QualificationReport) ([]runnerairgap.ActionTemplate, error) {
	result := make([]runnerairgap.ActionTemplate, 0, len(cfg.ActionWorkflowConfigs))
	for _, actionCfg := range cfg.ActionWorkflowConfigs {
		if step := firstGitSourcedActionStep(actionCfg); step != "" {
			addQualificationWarning(report, "action.git_source_excluded", "action:"+actionCfg.ActionWorkflowID,
				fmt.Sprintf("action %s was excluded from the bundle: step %s sources its contents from Git, and air-gap v1 actions must be inline-only", actionCfg.ActionWorkflow.Name, step))
			continue
		}
		steps := make([]*plantypes.ActionWorkflowRunStepPlan, 0, len(actionCfg.Steps))
		for _, stepCfg := range actionCfg.Steps {
			step := &app.InstallActionWorkflowRunStep{ID: virtualID("vast", actionCfg.ID+":"+stepCfg.ID), Step: stepCfg}
			rendered, err := planner.RenderActionWorkflowStepPlan(l, step, stateData, nil)
			if err != nil {
				return nil, fmt.Errorf("render action %s step %s: %w", actionCfg.ActionWorkflowID, stepCfg.ID, err)
			}
			steps = append(steps, rendered)
		}
		auth, err := planner.AuthForDeploy(l, role, stack, "action-"+actionCfg.ID)
		if err != nil {
			return nil, err
		}
		plan := &plantypes.ActionWorkflowRunPlan{InstallID: installID, ID: actionCfg.ID, Steps: steps, BuiltinEnvVars: map[string]string{}, OverrideEnvVars: map[string]string{}, Attrs: map[string]string{"action.name": actionCfg.ActionWorkflow.Name, "action.id": actionCfg.ActionWorkflowID}, AWSAuth: auth.AWS, AzureAuth: auth.Azure, GCPAuth: auth.GCP}
		composite, err := compositeJSON(&plantypes.CompositePlan{ActionWorkflowRunPlan: plan})
		if err != nil {
			return nil, err
		}
		cron := ""
		for _, trigger := range actionCfg.Triggers {
			if trigger.Type == app.ActionWorkflowTriggerTypeCron {
				cron = trigger.CronSchedule
				break
			}
		}
		result = append(result, runnerairgap.ActionTemplate{ID: actionCfg.ActionWorkflowID, Name: actionCfg.ActionWorkflow.Name, CronSchedule: cron, JobType: "actions-workflow", JobGroup: "actions", JobOperation: "exec", CompositePlan: composite})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func compileDriftTemplates(steps []runnerairgap.Step, components []runnerairgap.ComponentSpec) ([]runnerairgap.DriftTemplate, error) {
	byName := map[string]runnerairgap.ComponentSpec{}
	for _, component := range components {
		byName[component.ComponentName] = component
	}
	result := []runnerairgap.DriftTemplate{}
	for _, step := range steps {
		if step.JobType != "terraform-deploy" || step.JobOperation != "create-apply-plan" {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(step.ID, "deploy-"), "-plan")
		component := byName[name]
		plan, err := clearTerraformPlanArtifacts(step.CompositePlan)
		if err != nil {
			return nil, err
		}
		result = append(result, runnerairgap.DriftTemplate{ID: "drift-" + component.ComponentID, ComponentID: component.ComponentID, ComponentName: name, JobType: step.JobType, JobGroup: step.JobGroup, JobOperation: "create-apply-plan", CompositePlan: plan})
	}
	return result, nil
}
