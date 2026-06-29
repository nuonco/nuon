package config

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/nuonco/nuon/pkg/config/diff"
)

// Diff compares the receiver (new config) against old (previous config) and
// returns a hierarchical diff tree. The structure mirrors terraform plan output:
// each section shows added, removed, changed, or unchanged fields.
func (a *AppConfig) Diff(old *AppConfig) *diff.Diff {
	if old == nil {
		old = &AppConfig{}
	}

	var children []*diff.Diff

	// Metadata
	children = append(children,
		diff.NewDiff(diff.WithKey("version"), diff.WithStringDiff(old.Version, a.Version)),
		diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(old.Description, a.Description)),
		diff.NewDiff(diff.WithKey("display_name"), diff.WithStringDiff(old.DisplayName, a.DisplayName)),
		diff.NewDiff(diff.WithKey("slack_webhook_url"), diff.WithStringDiff(old.SlackWebhookURL, a.SlackWebhookURL)),
		diff.NewDiff(diff.WithKey("readme"), diff.WithStringDiff(old.Readme, a.Readme)),
	)

	if d := diffBranch(old.Branch, a.Branch); d != nil {
		children = append(children, d)
	}
	if d := diffSandbox(old.Sandbox, a.Sandbox); d != nil {
		children = append(children, d)
	}
	if d := diffRunner(old.Runner, a.Runner); d != nil {
		children = append(children, d)
	}
	if d := diffStack(old.Stack, a.Stack); d != nil {
		children = append(children, d)
	}
	if d := diffInputs(old.Inputs, a.Inputs); d != nil {
		children = append(children, d)
	}
	if d := diffPermissions(old.Permissions, a.Permissions); d != nil {
		children = append(children, d)
	}
	if d := diffPolicies(old.Policies, a.Policies); d != nil {
		children = append(children, d)
	}
	if d := diffSecrets(old.Secrets, a.Secrets); d != nil {
		children = append(children, d)
	}
	if d := diffBreakGlass(old.BreakGlass, a.BreakGlass); d != nil {
		children = append(children, d)
	}
	if d := diffOperationRoles(old.OperationRoles, a.OperationRoles); d != nil {
		children = append(children, d)
	}
	if d := diffComponents(old.Components, a.Components); d != nil {
		children = append(children, d)
	}
	if d := diffInstalls(old.Installs, a.Installs); d != nil {
		children = append(children, d)
	}
	if d := diffActions(old.Actions, a.Actions); d != nil {
		children = append(children, d)
	}

	return diff.NewDiff(
		diff.WithKey("app_config"),
		diff.WithChildren(children...),
	)
}

// --- Branch ---

func diffBranch(old, new *AppBranchConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &AppBranchConfig{}
	}
	if new == nil {
		new = &AppBranchConfig{}
	}

	children := []*diff.Diff{
		diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(old.Name, new.Name)),
	}
	children = append(children, diffConnectedRepo("connected_repo", old.ConnectedRepo, new.ConnectedRepo)...)

	return diff.NewDiff(diff.WithKey("branch"), diff.WithChildren(children...))
}

// --- Sandbox ---

func toTOMLString(v any) string {
	b, err := ToTOML(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// sectionTOML serializes a section sub-config to TOML, treating a zero-valued
// (absent) sub-config as empty. Without this, a newly added section would read
// as a change from its empty-default TOML (e.g. "input = []") rather than an add.
func sectionTOML(v any) string {
	s := toTOMLString(v)
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		if s == toTOMLString(reflect.New(rv.Type().Elem()).Interface()) {
			return ""
		}
	}
	return s
}

// sectionDiff builds a section node from its field children and, when the
// section's TOML serialization changed, also attaches a before/after content
// diff so the dashboard can render the whole section as a TOML line diff. The
// field children are retained for change counting.
func sectionDiff(key string, old, new any, children []*diff.Diff) *diff.Diff {
	opts := []diff.DiffOption{diff.WithKey(key), diff.WithChildren(children...)}
	if oldTOML, newTOML := sectionTOML(old), sectionTOML(new); oldTOML != newTOML {
		opts = append(opts, diff.WithContentDiff(oldTOML, newTOML))
	}
	return diff.NewDiff(opts...)
}

func diffSandbox(old, new *AppSandboxConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &AppSandboxConfig{}
	}
	if new == nil {
		new = &AppSandboxConfig{}
	}

	children := []*diff.Diff{
		diff.NewDiff(diff.WithKey("terraform_version"), diff.WithStringDiff(old.TerraformVersion, new.TerraformVersion)),
		diff.NewDiff(diff.WithKey("drift_schedule"), diff.WithOptionalStringDiff(old.DriftSchedule, new.DriftSchedule)),
	}
	children = append(children, diffPublicRepo("public_repo", old.PublicRepo, new.PublicRepo)...)
	children = append(children, diffConnectedRepo("connected_repo", old.ConnectedRepo, new.ConnectedRepo)...)
	if d := diff.MapDiff("env_vars", old.EnvVarMap, new.EnvVarMap); d != nil {
		children = append(children, d)
	}
	if d := diff.MapDiff("vars", old.VarsMap, new.VarsMap); d != nil {
		children = append(children, d)
	}

	return sectionDiff("sandbox", old, new, children)
}

// --- Runner ---

func diffRunner(old, new *AppRunnerConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &AppRunnerConfig{}
	}
	if new == nil {
		new = &AppRunnerConfig{}
	}

	children := []*diff.Diff{
		diff.NewDiff(diff.WithKey("runner_type"), diff.WithStringDiff(old.RunnerType, new.RunnerType)),
		diff.NewDiff(diff.WithKey("helm_driver"), diff.WithStringDiff(old.HelmDriver, new.HelmDriver)),
		diff.NewDiff(diff.WithKey("init_script_url"), diff.WithStringDiff(old.InitScriptURL, new.InitScriptURL)),
	}
	if d := diff.MapDiff("env_vars", old.EnvVarMap, new.EnvVarMap); d != nil {
		children = append(children, d)
	}

	return sectionDiff("runner", old, new, children)
}

// --- Stack ---

func diffStack(old, new *StackConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &StackConfig{}
	}
	if new == nil {
		new = &StackConfig{}
	}

	children := []*diff.Diff{
		diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(old.Type, new.Type)),
		diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(old.Name, new.Name)),
		diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(old.Description, new.Description)),
		diff.NewDiff(diff.WithKey("vpc_nested_template_url"), diff.WithStringDiff(old.VPCNestedTemplateURL, new.VPCNestedTemplateURL)),
		diff.NewDiff(diff.WithKey("runner_nested_template_url"), diff.WithStringDiff(old.RunnerNestedTemplateURL, new.RunnerNestedTemplateURL)),
	}

	// Custom nested stacks matched by name
	children = append(children, diffCustomNestedStacks(old.CustomNestedStacks, new.CustomNestedStacks)...)

	return sectionDiff("stack", old, new, children)
}

func diffCustomNestedStacks(old, new []CustomNestedStack) []*diff.Diff {
	oldByName := make(map[string]CustomNestedStack, len(old))
	for _, s := range old {
		oldByName[s.Name] = s
	}

	var diffs []*diff.Diff
	seen := make(map[string]bool)

	for _, s := range new {
		seen[s.Name] = true
		if os, ok := oldByName[s.Name]; ok {
			var stackChildren []*diff.Diff
			stackChildren = append(stackChildren,
				diff.NewDiff(diff.WithKey("template_url"), diff.WithStringDiff(os.TemplateURL, s.TemplateURL)),
				diff.NewDiff(diff.WithKey("index"), diff.WithStringDiff(fmt.Sprintf("%d", os.Index), fmt.Sprintf("%d", s.Index))),
			)
			if d := diff.MapDiff("parameters", os.Parameters, s.Parameters); d != nil {
				stackChildren = append(stackChildren, d)
			}
			diffs = append(diffs, diff.NewDiff(diff.WithKey("custom_stack."+s.Name), diff.WithChildren(stackChildren...)))
		} else {
			diffs = append(diffs, diff.NewDiff(diff.WithKey("custom_stack."+s.Name), diff.WithStringDiff("", s.TemplateURL)))
		}
	}

	for _, s := range old {
		if !seen[s.Name] {
			diffs = append(diffs, diff.NewDiff(diff.WithKey("custom_stack."+s.Name), diff.WithStringDiff(s.TemplateURL, "")))
		}
	}

	return diffs
}

// --- Inputs ---

func diffInputs(old, new *AppInputConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &AppInputConfig{}
	}
	if new == nil {
		new = &AppInputConfig{}
	}

	var children []*diff.Diff

	// Groups matched by name
	children = append(children, diffInputGroups(old.Groups, new.Groups)...)

	// Inputs matched by name
	children = append(children, diffAppInputs(old.Inputs, new.Inputs)...)

	return sectionDiff("inputs", old, new, children)
}

func diffInputGroups(old, new []AppInputGroup) []*diff.Diff {
	oldByName := make(map[string]AppInputGroup, len(old))
	for _, g := range old {
		oldByName[g.Name] = g
	}

	var diffs []*diff.Diff
	seen := make(map[string]bool)

	diffGroup := func(og, g AppInputGroup) *diff.Diff {
		return diff.NewDiff(
			diff.WithKey("group."+g.Name),
			diff.WithChildren(
				diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(og.Description, g.Description)),
				diff.NewDiff(diff.WithKey("display_name"), diff.WithStringDiff(og.DisplayName, g.DisplayName)),
			),
		)
	}

	for _, g := range new {
		seen[g.Name] = true
		if og, ok := oldByName[g.Name]; ok {
			diffs = append(diffs, diffGroup(og, g))
		} else {
			diffs = append(diffs, diffGroup(AppInputGroup{}, g))
		}
	}

	for _, g := range old {
		if !seen[g.Name] {
			diffs = append(diffs, diffGroup(g, AppInputGroup{Name: g.Name}))
		}
	}

	return diffs
}

func diffAppInputs(old, new []AppInput) []*diff.Diff {
	oldByName := make(map[string]AppInput, len(old))
	for _, i := range old {
		oldByName[i.Name] = i
	}

	var diffs []*diff.Diff
	seen := make(map[string]bool)

	diffInput := func(oi, i AppInput) *diff.Diff {
		return diff.NewDiff(
			diff.WithKey("input."+i.Name),
			diff.WithChildren(
				diff.NewDiff(diff.WithKey("display_name"), diff.WithStringDiff(oi.DisplayName, i.DisplayName)),
				diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(oi.Description, i.Description)),
				diff.NewDiff(diff.WithKey("group"), diff.WithStringDiff(oi.Group, i.Group)),
				diff.NewDiff(diff.WithKey("default"), diff.WithStringDiff(fmt.Sprintf("%v", oi.Default), fmt.Sprintf("%v", i.Default))),
				diff.NewDiff(diff.WithKey("required"), diff.WithBoolDiff(oi.Required, i.Required)),
				diff.NewDiff(diff.WithKey("sensitive"), diff.WithBoolDiff(oi.Sensitive, i.Sensitive)),
				diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(oi.Type, i.Type)),
				diff.NewDiff(diff.WithKey("user_configurable"), diff.WithBoolDiff(oi.UserConfigurable, i.UserConfigurable)),
			),
		)
	}

	for _, i := range new {
		seen[i.Name] = true
		if oi, ok := oldByName[i.Name]; ok {
			diffs = append(diffs, diffInput(oi, i))
		} else {
			diffs = append(diffs, diffInput(AppInput{}, i))
		}
	}

	for _, i := range old {
		if !seen[i.Name] {
			diffs = append(diffs, diffInput(i, AppInput{Name: i.Name}))
		}
	}

	return diffs
}

// --- Permissions ---

func diffPermissions(old, new *PermissionsConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &PermissionsConfig{}
	}
	if new == nil {
		new = &PermissionsConfig{}
	}

	var children []*diff.Diff

	// Diff the standard roles
	if d := diffIAMRole("provision_role", old.ProvisionRole, new.ProvisionRole); d != nil {
		children = append(children, d)
	}
	if d := diffIAMRole("deprovision_role", old.DeprovisionRole, new.DeprovisionRole); d != nil {
		children = append(children, d)
	}
	if d := diffIAMRole("maintenance_role", old.MaintenanceRole, new.MaintenanceRole); d != nil {
		children = append(children, d)
	}

	// Custom roles matched by name
	children = append(children, diffIAMRoles("custom_role", old.CustomRoles, new.CustomRoles)...)

	return sectionDiff("permissions", old, new, children)
}

func diffIAMRole(key string, old, new *AppAWSIAMRole) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &AppAWSIAMRole{}
	}
	if new == nil {
		new = &AppAWSIAMRole{}
	}

	return diff.NewDiff(
		diff.WithKey(key),
		diff.WithChildren(
			diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(old.Name, new.Name)),
			diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(old.Description, new.Description)),
			diff.NewDiff(diff.WithKey("display_name"), diff.WithStringDiff(old.DisplayName, new.DisplayName)),
			diff.NewDiff(diff.WithKey("cloud_platform"), diff.WithStringDiff(old.CloudPlatform, new.CloudPlatform)),
			diff.NewDiff(diff.WithKey("permissions_boundary"), diff.WithStringDiff(old.PermissionsBoundary, new.PermissionsBoundary)),
			diff.NewDiff(diff.WithKey("enabled_in_stack"), diff.WithOptionalBoolDiff(old.EnabledInStack, new.EnabledInStack)),
		),
	)
}

func diffIAMRoles(prefix string, old, new []*AppAWSIAMRole) []*diff.Diff {
	oldByName := make(map[string]*AppAWSIAMRole, len(old))
	for _, r := range old {
		oldByName[r.Name] = r
	}

	var diffs []*diff.Diff
	seen := make(map[string]bool)

	for _, r := range new {
		seen[r.Name] = true
		if or, ok := oldByName[r.Name]; ok {
			diffs = append(diffs, diffIAMRole(prefix+"."+r.Name, or, r))
		} else {
			diffs = append(diffs, diffIAMRole(prefix+"."+r.Name, &AppAWSIAMRole{}, r))
		}
	}

	for _, r := range old {
		if !seen[r.Name] {
			diffs = append(diffs, diffIAMRole(prefix+"."+r.Name, r, &AppAWSIAMRole{Name: r.Name}))
		}
	}

	return diffs
}

// --- Policies ---

func diffPolicies(old, new *PoliciesConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &PoliciesConfig{}
	}
	if new == nil {
		new = &PoliciesConfig{}
	}

	oldByName := make(map[string]AppPolicy, len(old.Policies))
	for _, p := range old.Policies {
		oldByName[p.Name] = p
	}

	var children []*diff.Diff
	seen := make(map[string]bool)

	diffPolicy := func(op, p AppPolicy) *diff.Diff {
		return diff.NewDiff(
			diff.WithKey("policy."+p.Name),
			diff.WithChildren(
				diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(string(op.Type), string(p.Type))),
				diff.NewDiff(diff.WithKey("engine"), diff.WithStringDiff(string(op.Engine), string(p.Engine))),
				diff.NewDiff(diff.WithKey("contents"), diff.WithStringDiff(op.Contents, p.Contents)),
				diff.NewDiff(diff.WithKey("components"), diff.WithStringSliceDiff(op.Components, p.Components)),
			),
		)
	}

	for _, p := range new.Policies {
		seen[p.Name] = true
		if op, ok := oldByName[p.Name]; ok {
			children = append(children, diffPolicy(op, p))
		} else {
			children = append(children, diffPolicy(AppPolicy{}, p))
		}
	}

	for _, p := range old.Policies {
		if !seen[p.Name] {
			children = append(children, diffPolicy(p, AppPolicy{Name: p.Name}))
		}
	}

	return diff.NewDiff(diff.WithKey("policies"), diff.WithChildren(children...))
}

// --- Secrets ---

func diffSecrets(old, new *SecretsConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &SecretsConfig{}
	}
	if new == nil {
		new = &SecretsConfig{}
	}

	oldByName := make(map[string]*AppSecret, len(old.Secrets))
	for _, s := range old.Secrets {
		oldByName[s.Name] = s
	}

	var children []*diff.Diff
	seen := make(map[string]bool)

	diffSecret := func(os, s *AppSecret) *diff.Diff {
		return diff.NewDiff(
			diff.WithKey("secret."+s.Name),
			diff.WithChildren(
				diff.NewDiff(diff.WithKey("display_name"), diff.WithStringDiff(os.DisplayName, s.DisplayName)),
				diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(os.Description, s.Description)),
				diff.NewDiff(diff.WithKey("required"), diff.WithBoolDiff(os.Required, s.Required)),
				diff.NewDiff(diff.WithKey("auto_generate"), diff.WithBoolDiff(os.AutoGenerate, s.AutoGenerate)),
				diff.NewDiff(diff.WithKey("format"), diff.WithStringDiff(os.Format, s.Format)),
				diff.NewDiff(diff.WithKey("default"), diff.WithStringDiff(os.Default, s.Default)),
				diff.NewDiff(diff.WithKey("kubernetes_sync"), diff.WithOptionalBoolDiff(os.KubernetesSync, s.KubernetesSync)),
				diff.NewDiff(diff.WithKey("kubernetes_secret_namespace"), diff.WithStringDiff(os.KubernetesSecretNamespace, s.KubernetesSecretNamespace)),
				diff.NewDiff(diff.WithKey("kubernetes_secret_name"), diff.WithStringDiff(os.KubernetesSecretName, s.KubernetesSecretName)),
			),
		)
	}

	for _, s := range new.Secrets {
		seen[s.Name] = true
		if os, ok := oldByName[s.Name]; ok {
			children = append(children, diffSecret(os, s))
		} else {
			children = append(children, diffSecret(&AppSecret{}, s))
		}
	}

	for _, s := range old.Secrets {
		if !seen[s.Name] {
			children = append(children, diffSecret(s, &AppSecret{Name: s.Name}))
		}
	}

	return sectionDiff("secrets", old, new, children)
}

// --- BreakGlass ---

func diffBreakGlass(old, new *BreakGlass) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &BreakGlass{}
	}
	if new == nil {
		new = &BreakGlass{}
	}

	children := diffIAMRoles("role", old.Roles, new.Roles)
	return diff.NewDiff(diff.WithKey("break_glass"), diff.WithChildren(children...))
}

// --- OperationRoles ---

func diffOperationRoles(old, new *OperationRolesConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &OperationRolesConfig{}
	}
	if new == nil {
		new = &OperationRolesConfig{}
	}

	children := []*diff.Diff{
		diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(string(old.Type), string(new.Type))),
	}

	// Rules matched by composite key (principal + operation)
	type ruleKey struct {
		principal string
		operation OperationType
	}
	oldRules := make(map[ruleKey]*OperationRoleRule, len(old.RuleMatrix))
	for _, r := range old.RuleMatrix {
		oldRules[ruleKey{r.Principal, r.Operation}] = r
	}

	seen := make(map[ruleKey]bool)
	for _, r := range new.RuleMatrix {
		k := ruleKey{r.Principal, r.Operation}
		seen[k] = true
		key := fmt.Sprintf("rule.%s.%s", r.Principal, r.Operation)
		if or, ok := oldRules[k]; ok {
			children = append(children, diff.NewDiff(
				diff.WithKey(key),
				diff.WithChildren(
					diff.NewDiff(diff.WithKey("role_name"), diff.WithStringDiff(or.RoleName, r.RoleName)),
				),
			))
		} else {
			children = append(children, diff.NewDiff(diff.WithKey(key), diff.WithStringDiff("", r.RoleName)))
		}
	}
	for _, r := range old.RuleMatrix {
		k := ruleKey{r.Principal, r.Operation}
		if !seen[k] {
			key := fmt.Sprintf("rule.%s.%s", r.Principal, r.Operation)
			children = append(children, diff.NewDiff(diff.WithKey(key), diff.WithStringDiff(r.RoleName, "")))
		}
	}

	return diff.NewDiff(diff.WithKey("operation_roles"), diff.WithChildren(children...))
}

// --- Components ---

func diffComponents(old, new ComponentList) *diff.Diff {
	if len(old) == 0 && len(new) == 0 {
		return nil
	}

	oldByName := make(map[string]*Component, len(old))
	for _, c := range old {
		oldByName[c.Name] = c
	}

	var children []*diff.Diff
	seen := make(map[string]bool)

	// Sort new components for deterministic output
	sorted := make([]*Component, len(new))
	copy(sorted, new)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	for _, c := range sorted {
		seen[c.Name] = true
		if oc, ok := oldByName[c.Name]; ok {
			children = append(children, diffComponent(oc, c))
		} else {
			// Added component — diff against empty to show all fields
			children = append(children, diffComponent(&Component{}, c))
		}
	}

	// Sort old components for deterministic removed order
	sortedOld := make([]*Component, len(old))
	copy(sortedOld, old)
	sort.Slice(sortedOld, func(i, j int) bool { return sortedOld[i].Name < sortedOld[j].Name })

	for _, c := range sortedOld {
		if !seen[c.Name] {
			// Removed component — diff against empty to show all fields
			children = append(children, diffComponent(c, &Component{Name: c.Name}))
		}
	}

	return diff.NewDiff(diff.WithKey("components"), diff.WithChildren(children...))
}

func diffComponent(old, new *Component) *diff.Diff {
	var children []*diff.Diff

	children = append(children,
		diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(string(old.Type), string(new.Type))),
		diff.NewDiff(diff.WithKey("var_name"), diff.WithStringDiff(old.VarName, new.VarName)),
		diff.NewDiff(diff.WithKey("dependencies"), diff.WithStringSliceDiff(old.Dependencies, new.Dependencies)),
	)

	if d := diff.MapDiff("labels", old.Labels, new.Labels); d != nil {
		children = append(children, d)
	}

	// Type-specific diffs
	children = append(children, diffHelmChart(old.HelmChart, new.HelmChart)...)
	children = append(children, diffTerraformModule(old.TerraformModule, new.TerraformModule)...)
	children = append(children, diffDockerBuild(old.DockerBuild, new.DockerBuild)...)
	children = append(children, diffExternalImage(old.ExternalImage, new.ExternalImage)...)
	children = append(children, diffKubernetesManifest(old.KubernetesManifest, new.KubernetesManifest)...)
	children = append(children, diffJob(old.Job, new.Job)...)
	children = append(children, diffPulumi(old.Pulumi, new.Pulumi)...)

	return diff.NewDiff(diff.WithKey("component."+new.Name), diff.WithChildren(children...))
}

// --- Component type-specific diffs ---

func diffHelmChart(old, new *HelmChartComponentConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &HelmChartComponentConfig{}
	}
	if new == nil {
		new = &HelmChartComponentConfig{}
	}

	var diffs []*diff.Diff
	diffs = append(diffs,
		diff.NewDiff(diff.WithKey("chart_name"), diff.WithStringDiff(old.ChartName, new.ChartName)),
		diff.NewDiff(diff.WithKey("namespace"), diff.WithStringDiff(old.Namespace, new.Namespace)),
		diff.NewDiff(diff.WithKey("storage_driver"), diff.WithStringDiff(old.StorageDriver, new.StorageDriver)),
		diff.NewDiff(diff.WithKey("take_ownership"), diff.WithBoolDiff(old.TakeOwnership, new.TakeOwnership)),
		diff.NewDiff(diff.WithKey("drift_schedule"), diff.WithOptionalStringDiff(old.DriftSchedule, new.DriftSchedule)),
		diff.NewDiff(diff.WithKey("build_timeout"), diff.WithStringDiff(old.BuildTimeout, new.BuildTimeout)),
		diff.NewDiff(diff.WithKey("deploy_timeout"), diff.WithStringDiff(old.DeployTimeout, new.DeployTimeout)),
	)

	if d := diff.MapDiff("values", old.ValuesMap, new.ValuesMap); d != nil {
		diffs = append(diffs, d)
	}

	if d := diffHelmValuesFiles(old.ValuesFiles, new.ValuesFiles); d != nil {
		diffs = append(diffs, d)
	}

	diffs = append(diffs, diffPublicRepo("public_repo", old.PublicRepo, new.PublicRepo)...)
	diffs = append(diffs, diffConnectedRepo("connected_repo", old.ConnectedRepo, new.ConnectedRepo)...)
	diffs = append(diffs, diffHelmRepo(old.HelmRepo, new.HelmRepo)...)

	return diffs
}

func diffHelmValuesFiles(old, new []HelmValuesFile) *diff.Diff {
	if len(old) == 0 && len(new) == 0 {
		return nil
	}

	oldByPath := make(map[string]HelmValuesFile, len(old))
	for _, f := range old {
		key := f.Path
		if key == "" {
			key = f.Source
		}
		oldByPath[key] = f
	}

	var children []*diff.Diff
	seen := make(map[string]bool)

	for _, f := range new {
		key := f.Path
		if key == "" {
			key = f.Source
		}
		seen[key] = true

		of := oldByPath[key]
		if of.Contents != f.Contents {
			children = append(children, diff.NewDiff(diff.WithKey(key), diff.WithContentDiff(of.Contents, f.Contents)))
		}
	}

	for key, f := range oldByPath {
		if !seen[key] {
			children = append(children, diff.NewDiff(diff.WithKey(key), diff.WithContentDiff(f.Contents, "")))
		}
	}

	if len(children) == 0 {
		return nil
	}
	return diff.NewDiff(diff.WithKey("values_files"), diff.WithChildren(children...))
}

func diffTerraformModule(old, new *TerraformModuleComponentConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &TerraformModuleComponentConfig{}
	}
	if new == nil {
		new = &TerraformModuleComponentConfig{}
	}

	var diffs []*diff.Diff
	diffs = append(diffs,
		diff.NewDiff(diff.WithKey("terraform_version"), diff.WithStringDiff(old.TerraformVersion, new.TerraformVersion)),
		diff.NewDiff(diff.WithKey("drift_schedule"), diff.WithOptionalStringDiff(old.DriftSchedule, new.DriftSchedule)),
		diff.NewDiff(diff.WithKey("build_timeout"), diff.WithStringDiff(old.BuildTimeout, new.BuildTimeout)),
		diff.NewDiff(diff.WithKey("deploy_timeout"), diff.WithStringDiff(old.DeployTimeout, new.DeployTimeout)),
	)

	if d := diff.MapDiff("env_vars", old.EnvVarMap, new.EnvVarMap); d != nil {
		diffs = append(diffs, d)
	}
	if d := diff.MapDiff("vars", old.VarsMap, new.VarsMap); d != nil {
		diffs = append(diffs, d)
	}

	if d := diffTerraformVariablesFiles(old.VariablesFiles, new.VariablesFiles); d != nil {
		diffs = append(diffs, d)
	}

	diffs = append(diffs, diffPublicRepo("public_repo", old.PublicRepo, new.PublicRepo)...)
	diffs = append(diffs, diffConnectedRepo("connected_repo", old.ConnectedRepo, new.ConnectedRepo)...)

	return diffs
}

func diffTerraformVariablesFiles(old, new []TerraformVariablesFile) *diff.Diff {
	maxLen := len(old)
	if len(new) > maxLen {
		maxLen = len(new)
	}
	if maxLen == 0 {
		return nil
	}

	var children []*diff.Diff
	for i := 0; i < maxLen; i++ {
		var oldContents, newContents string
		if i < len(old) {
			oldContents = old[i].Contents
		}
		if i < len(new) {
			newContents = new[i].Contents
		}
		if oldContents == newContents {
			continue
		}
		children = append(children, diff.NewDiff(diff.WithKey(fmt.Sprintf("var_file[%d]", i)), diff.WithContentDiff(oldContents, newContents)))
	}

	if len(children) == 0 {
		return nil
	}
	return diff.NewDiff(diff.WithKey("var_file"), diff.WithChildren(children...))
}

func diffDockerBuild(old, new *DockerBuildComponentConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &DockerBuildComponentConfig{}
	}
	if new == nil {
		new = &DockerBuildComponentConfig{}
	}

	var diffs []*diff.Diff
	diffs = append(diffs,
		diff.NewDiff(diff.WithKey("dockerfile"), diff.WithStringDiff(old.Dockerfile, new.Dockerfile)),
		diff.NewDiff(diff.WithKey("build_timeout"), diff.WithStringDiff(old.BuildTimeout, new.BuildTimeout)),
		diff.NewDiff(diff.WithKey("deploy_timeout"), diff.WithStringDiff(old.DeployTimeout, new.DeployTimeout)),
	)

	if d := diff.MapDiff("env_vars", old.EnvVarMap, new.EnvVarMap); d != nil {
		diffs = append(diffs, d)
	}

	diffs = append(diffs, diffPublicRepo("public_repo", old.PublicRepo, new.PublicRepo)...)
	diffs = append(diffs, diffConnectedRepo("connected_repo", old.ConnectedRepo, new.ConnectedRepo)...)

	return diffs
}

func diffExternalImage(old, new *ExternalImageComponentConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &ExternalImageComponentConfig{}
	}
	if new == nil {
		new = &ExternalImageComponentConfig{}
	}

	var diffs []*diff.Diff
	diffs = append(diffs,
		diff.NewDiff(diff.WithKey("build_timeout"), diff.WithStringDiff(old.BuildTimeout, new.BuildTimeout)),
		diff.NewDiff(diff.WithKey("deploy_timeout"), diff.WithStringDiff(old.DeployTimeout, new.DeployTimeout)),
	)

	// AWS ECR
	diffs = append(diffs, diffAWSECR(old.AWSECRImageConfig, new.AWSECRImageConfig)...)
	// GCP GAR
	diffs = append(diffs, diffGCPGAR(old.GCPGARImageConfig, new.GCPGARImageConfig)...)
	// Public image
	diffs = append(diffs, diffPublicImage(old.PublicImageConfig, new.PublicImageConfig)...)

	return diffs
}

func diffAWSECR(old, new *AWSECRConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &AWSECRConfig{}
	}
	if new == nil {
		new = &AWSECRConfig{}
	}
	return []*diff.Diff{
		diff.NewDiff(diff.WithKey("aws_ecr"), diff.WithChildren(
			diff.NewDiff(diff.WithKey("iam_role_arn"), diff.WithStringDiff(old.IAMRoleARN, new.IAMRoleARN)),
			diff.NewDiff(diff.WithKey("region"), diff.WithStringDiff(old.AWSRegion, new.AWSRegion)),
			diff.NewDiff(diff.WithKey("image_url"), diff.WithStringDiff(old.ImageURL, new.ImageURL)),
			diff.NewDiff(diff.WithKey("tag"), diff.WithStringDiff(old.Tag, new.Tag)),
		)),
	}
}

func diffGCPGAR(old, new *GCPGARConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &GCPGARConfig{}
	}
	if new == nil {
		new = &GCPGARConfig{}
	}
	return []*diff.Diff{
		diff.NewDiff(diff.WithKey("gcp_gar"), diff.WithChildren(
			diff.NewDiff(diff.WithKey("gcp_project_id"), diff.WithStringDiff(old.GCPProjectID, new.GCPProjectID)),
			diff.NewDiff(diff.WithKey("region"), diff.WithStringDiff(old.GCPRegion, new.GCPRegion)),
			diff.NewDiff(diff.WithKey("image_url"), diff.WithStringDiff(old.ImageURL, new.ImageURL)),
			diff.NewDiff(diff.WithKey("tag"), diff.WithStringDiff(old.Tag, new.Tag)),
			diff.NewDiff(diff.WithKey("service_account_email"), diff.WithStringDiff(old.ServiceAccountEmail, new.ServiceAccountEmail)),
		)),
	}
}

func diffPublicImage(old, new *PublicImageConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &PublicImageConfig{}
	}
	if new == nil {
		new = &PublicImageConfig{}
	}
	return []*diff.Diff{
		diff.NewDiff(diff.WithKey("public_image"), diff.WithChildren(
			diff.NewDiff(diff.WithKey("image_url"), diff.WithStringDiff(old.ImageURL, new.ImageURL)),
			diff.NewDiff(diff.WithKey("tag"), diff.WithStringDiff(old.Tag, new.Tag)),
		)),
	}
}

func diffKubernetesManifest(old, new *KubernetesManifestComponentConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &KubernetesManifestComponentConfig{}
	}
	if new == nil {
		new = &KubernetesManifestComponentConfig{}
	}

	var diffs []*diff.Diff
	diffs = append(diffs,
		diff.NewDiff(diff.WithKey("manifest"), diff.WithContentDiff(old.Manifest, new.Manifest)),
		diff.NewDiff(diff.WithKey("namespace"), diff.WithStringDiff(old.Namespace, new.Namespace)),
		diff.NewDiff(diff.WithKey("drift_schedule"), diff.WithOptionalStringDiff(old.DriftSchedule, new.DriftSchedule)),
		diff.NewDiff(diff.WithKey("build_timeout"), diff.WithStringDiff(old.BuildTimeout, new.BuildTimeout)),
		diff.NewDiff(diff.WithKey("deploy_timeout"), diff.WithStringDiff(old.DeployTimeout, new.DeployTimeout)),
	)

	// Kustomize
	diffs = append(diffs, diffKustomize(old.Kustomize, new.Kustomize)...)
	diffs = append(diffs, diffPublicRepo("public_repo", old.PublicRepo, new.PublicRepo)...)
	diffs = append(diffs, diffConnectedRepo("connected_repo", old.ConnectedRepo, new.ConnectedRepo)...)

	return diffs
}

func diffKustomize(old, new *KustomizeConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &KustomizeConfig{}
	}
	if new == nil {
		new = &KustomizeConfig{}
	}
	return []*diff.Diff{
		diff.NewDiff(diff.WithKey("kustomize"), diff.WithChildren(
			diff.NewDiff(diff.WithKey("path"), diff.WithStringDiff(old.Path, new.Path)),
			diff.NewDiff(diff.WithKey("patches"), diff.WithStringSliceDiff(old.Patches, new.Patches)),
			diff.NewDiff(diff.WithKey("enable_helm"), diff.WithBoolDiff(old.EnableHelm, new.EnableHelm)),
			diff.NewDiff(diff.WithKey("load_restrictor"), diff.WithStringDiff(old.LoadRestrictor, new.LoadRestrictor)),
		)),
	}
}

func diffJob(old, new *JobComponentConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &JobComponentConfig{}
	}
	if new == nil {
		new = &JobComponentConfig{}
	}

	var diffs []*diff.Diff
	diffs = append(diffs,
		diff.NewDiff(diff.WithKey("image_url"), diff.WithStringDiff(old.ImageURL, new.ImageURL)),
		diff.NewDiff(diff.WithKey("tag"), diff.WithStringDiff(old.Tag, new.Tag)),
		diff.NewDiff(diff.WithKey("cmd"), diff.WithStringSliceDiff(old.Cmd, new.Cmd)),
		diff.NewDiff(diff.WithKey("args"), diff.WithStringSliceDiff(old.Args, new.Args)),
		diff.NewDiff(diff.WithKey("build_timeout"), diff.WithStringDiff(old.BuildTimeout, new.BuildTimeout)),
		diff.NewDiff(diff.WithKey("deploy_timeout"), diff.WithStringDiff(old.DeployTimeout, new.DeployTimeout)),
	)

	if d := diff.MapDiff("env_vars", old.EnvVarMap, new.EnvVarMap); d != nil {
		diffs = append(diffs, d)
	}

	return diffs
}

func diffPulumi(old, new *PulumiComponentConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &PulumiComponentConfig{}
	}
	if new == nil {
		new = &PulumiComponentConfig{}
	}

	var diffs []*diff.Diff
	diffs = append(diffs,
		diff.NewDiff(diff.WithKey("runtime"), diff.WithStringDiff(old.Runtime, new.Runtime)),
		diff.NewDiff(diff.WithKey("pulumi_version"), diff.WithStringDiff(old.PulumiVersion, new.PulumiVersion)),
		diff.NewDiff(diff.WithKey("drift_schedule"), diff.WithOptionalStringDiff(old.DriftSchedule, new.DriftSchedule)),
		diff.NewDiff(diff.WithKey("build_timeout"), diff.WithStringDiff(old.BuildTimeout, new.BuildTimeout)),
		diff.NewDiff(diff.WithKey("deploy_timeout"), diff.WithStringDiff(old.DeployTimeout, new.DeployTimeout)),
	)

	if d := diff.MapDiff("config", old.ConfigMap, new.ConfigMap); d != nil {
		diffs = append(diffs, d)
	}
	if d := diff.MapDiff("env_vars", old.EnvVarMap, new.EnvVarMap); d != nil {
		diffs = append(diffs, d)
	}

	diffs = append(diffs, diffPublicRepo("public_repo", old.PublicRepo, new.PublicRepo)...)
	diffs = append(diffs, diffConnectedRepo("connected_repo", old.ConnectedRepo, new.ConnectedRepo)...)

	return diffs
}

// --- Installs ---

func diffInstalls(old, new []*Install) *diff.Diff {
	if len(old) == 0 && len(new) == 0 {
		return nil
	}

	oldByName := make(map[string]*Install, len(old))
	for _, i := range old {
		oldByName[i.Name] = i
	}

	var children []*diff.Diff
	seen := make(map[string]bool)

	for _, i := range new {
		seen[i.Name] = true
		if oi, ok := oldByName[i.Name]; ok {
			d, _ := i.Diff(oi)
			if d != nil {
				children = append(children, d)
			}
		} else {
			d, _ := i.Diff(nil)
			if d != nil {
				children = append(children, d)
			}
		}
	}

	for _, i := range old {
		if !seen[i.Name] {
			// Removed install — diff showing all fields being removed
			empty := &Install{Name: i.Name}
			d, _ := empty.Diff(i)
			if d != nil {
				children = append(children, d)
			}
		}
	}

	return diff.NewDiff(diff.WithKey("installs"), diff.WithChildren(children...))
}

// --- Actions ---

func diffActions(old, new []*ActionConfig) *diff.Diff {
	if len(old) == 0 && len(new) == 0 {
		return nil
	}

	oldByName := make(map[string]*ActionConfig, len(old))
	for _, a := range old {
		oldByName[a.Name] = a
	}

	var children []*diff.Diff
	seen := make(map[string]bool)

	for _, a := range new {
		seen[a.Name] = true
		if oa, ok := oldByName[a.Name]; ok {
			children = append(children, diffAction(oa, a))
		} else {
			children = append(children, diffAction(&ActionConfig{}, a))
		}
	}

	for _, a := range old {
		if !seen[a.Name] {
			children = append(children, diffAction(a, &ActionConfig{Name: a.Name}))
		}
	}

	return diff.NewDiff(diff.WithKey("actions"), diff.WithChildren(children...))
}

func diffAction(old, new *ActionConfig) *diff.Diff {
	var children []*diff.Diff
	children = append(children,
		diff.NewDiff(diff.WithKey("timeout"), diff.WithStringDiff(old.Timeout, new.Timeout)),
		diff.NewDiff(diff.WithKey("role"), diff.WithStringDiff(old.Role, new.Role)),
		diff.NewDiff(diff.WithKey("break_glass_role"), diff.WithStringDiff(old.BreakGlassRole, new.BreakGlassRole)),
		diff.NewDiff(diff.WithKey("enable_kube_config"), diff.WithOptionalBoolDiff(old.EnableKubeConfig, new.EnableKubeConfig)),
	)

	if d := diff.MapDiff("labels", old.Labels, new.Labels); d != nil {
		children = append(children, d)
	}

	// Steps matched by name
	children = append(children, diffActionSteps(old.Steps, new.Steps)...)

	// Triggers matched by index
	children = append(children, diffActionTriggers(old.Triggers, new.Triggers)...)

	return diff.NewDiff(diff.WithKey("action."+new.Name), diff.WithChildren(children...))
}

func diffActionSteps(old, new []*ActionStepConfig) []*diff.Diff {
	oldByName := make(map[string]*ActionStepConfig, len(old))
	for _, s := range old {
		oldByName[s.Name] = s
	}

	var diffs []*diff.Diff
	seen := make(map[string]bool)

	diffStep := func(os, s *ActionStepConfig) *diff.Diff {
		stepChildren := []*diff.Diff{
			diff.NewDiff(diff.WithKey("command"), diff.WithStringDiff(os.Command, s.Command)),
			diff.NewDiff(diff.WithKey("inline_contents"), diff.WithContentDiff(os.InlineContents, s.InlineContents)),
		}
		if d := diff.MapDiff("env_vars", os.EnvVarMap, s.EnvVarMap); d != nil {
			stepChildren = append(stepChildren, d)
		}
		stepChildren = append(stepChildren, diffPublicRepo("public_repo", os.PublicRepo, s.PublicRepo)...)
		stepChildren = append(stepChildren, diffConnectedRepo("connected_repo", os.ConnectedRepo, s.ConnectedRepo)...)
		return diff.NewDiff(diff.WithKey("step."+s.Name), diff.WithChildren(stepChildren...))
	}

	for _, s := range new {
		seen[s.Name] = true
		if os, ok := oldByName[s.Name]; ok {
			diffs = append(diffs, diffStep(os, s))
		} else {
			diffs = append(diffs, diffStep(&ActionStepConfig{}, s))
		}
	}

	for _, s := range old {
		if !seen[s.Name] {
			diffs = append(diffs, diffStep(s, &ActionStepConfig{Name: s.Name}))
		}
	}

	return diffs
}

func diffActionTriggers(old, new []*ActionTriggerConfig) []*diff.Diff {
	var diffs []*diff.Diff

	maxLen := len(old)
	if len(new) > maxLen {
		maxLen = len(new)
	}

	for i := 0; i < maxLen; i++ {
		key := fmt.Sprintf("trigger.%d", i)
		if i < len(old) && i < len(new) {
			diffs = append(diffs, diff.NewDiff(
				diff.WithKey(key),
				diff.WithChildren(
					diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(old[i].Type, new[i].Type)),
					diff.NewDiff(diff.WithKey("cron_schedule"), diff.WithStringDiff(old[i].CronSchedule, new[i].CronSchedule)),
					diff.NewDiff(diff.WithKey("component_name"), diff.WithStringDiff(old[i].ComponentName, new[i].ComponentName)),
				),
			))
		} else if i < len(new) {
			diffs = append(diffs, diff.NewDiff(diff.WithKey(key), diff.WithStringDiff("", new[i].Type)))
		} else {
			diffs = append(diffs, diff.NewDiff(diff.WithKey(key), diff.WithStringDiff(old[i].Type, "")))
		}
	}

	return diffs
}

// --- Shared repo helpers ---

func diffPublicRepo(key string, old, new *PublicRepoConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &PublicRepoConfig{}
	}
	if new == nil {
		new = &PublicRepoConfig{}
	}
	return []*diff.Diff{
		diff.NewDiff(diff.WithKey(key), diff.WithChildren(
			diff.NewDiff(diff.WithKey("repo"), diff.WithStringDiff(old.Repo, new.Repo)),
			diff.NewDiff(diff.WithKey("directory"), diff.WithStringDiff(old.Directory, new.Directory)),
			diff.NewDiff(diff.WithKey("branch"), diff.WithStringDiff(old.Branch, new.Branch)),
		)),
	}
}

func diffConnectedRepo(key string, old, new *ConnectedRepoConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &ConnectedRepoConfig{}
	}
	if new == nil {
		new = &ConnectedRepoConfig{}
	}
	return []*diff.Diff{
		diff.NewDiff(diff.WithKey(key), diff.WithChildren(
			diff.NewDiff(diff.WithKey("repo"), diff.WithStringDiff(old.Repo, new.Repo)),
			diff.NewDiff(diff.WithKey("directory"), diff.WithStringDiff(old.Directory, new.Directory)),
			diff.NewDiff(diff.WithKey("branch"), diff.WithStringDiff(old.Branch, new.Branch)),
		)),
	}
}

func diffHelmRepo(old, new *HelmRepoConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &HelmRepoConfig{}
	}
	if new == nil {
		new = &HelmRepoConfig{}
	}
	return []*diff.Diff{
		diff.NewDiff(diff.WithKey("helm_repo"), diff.WithChildren(
			diff.NewDiff(diff.WithKey("repo_url"), diff.WithStringDiff(old.RepoURL, new.RepoURL)),
			diff.NewDiff(diff.WithKey("chart"), diff.WithStringDiff(old.Chart, new.Chart)),
			diff.NewDiff(diff.WithKey("version"), diff.WithStringDiff(old.Version, new.Version)),
		)),
	}
}
