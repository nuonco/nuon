package config

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"
	"strconv"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/oci/signature"
)

// Diff compares the receiver (new config) against old (previous config) and
// returns a hierarchical diff tree. The structure mirrors terraform plan output:
// each section shows added, removed, changed, or unchanged fields.
func (a *AppConfig) Diff(old *AppConfig, changedInputs ...diff.NodeID) *diff.Diff {
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
	if d := diff.MapDiff("label_colors", old.LabelColors, a.LabelColors); d != nil {
		children = append(children, d)
	}
	if d := diff.MapDiff("default_labels", old.DefaultLabels, a.DefaultLabels); d != nil {
		children = append(children, d)
	}

	if d := diffBranch(old.Branch, a.Branch); d != nil {
		children = append(children, d)
	}
	if d := diffBranches(old.Branches, a.Branches); d != nil {
		children = append(children, d)
	}
	if d := diffKubernetesContexts(old.KubernetesContexts, a.KubernetesContexts); d != nil {
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
	if d := diffInstallsConfig(old.InstallsConfig, a.InstallsConfig); d != nil {
		children = append(children, d)
	}
	if d := diffActions(old.Actions, a.Actions); d != nil {
		children = append(children, d)
	}
	children = append(children, diff.NewDiff(
		diff.WithKey("triggers"),
		diff.WithStringDiff(sectionTOML(old.Triggers), sectionTOML(a.Triggers)),
	))
	if d := diffRunbooks(old.Runbooks, a.Runbooks); d != nil {
		children = append(children, d)
	}

	result := diff.NewDiff(
		diff.WithKey("app_config"),
		diff.WithChildren(children...),
	)
	graph := appConfigDependencyGraph(old, a)
	result.ApplyImpacts(graph.Propagate(append(changedResourceIDs(result), changedInputs...)...))
	return result
}

// --- Branch ---

func diffBranch(old, new *AppBranchConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	return branchDiff("branch", old, new)
}

func diffBranches(old, new []*AppBranchConfig) *diff.Diff {
	if len(old) == 0 && len(new) == 0 {
		return nil
	}
	oldByName, newByName := map[string]*AppBranchConfig{}, map[string]*AppBranchConfig{}
	names := map[string]bool{}
	for _, branch := range old {
		oldByName[branch.Name], names[branch.Name] = branch, true
	}
	for _, branch := range new {
		newByName[branch.Name], names[branch.Name] = branch, true
	}
	var children []*diff.Diff
	for _, name := range slices.Sorted(maps.Keys(names)) {
		children = append(children, branchDiff("branch."+name, oldByName[name], newByName[name]))
	}
	return diff.NewDiff(diff.WithKey("branches"), diff.WithChildren(children...))
}

func diffKubernetesContexts(old, new *KubernetesContextsConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &KubernetesContextsConfig{}
	}
	if new == nil {
		new = &KubernetesContextsConfig{}
	}
	oldByName, newByName := map[string]KubernetesContext{}, map[string]KubernetesContext{}
	names := map[string]bool{}
	for _, context := range old.Contexts {
		oldByName[context.Name], names[context.Name] = *context, true
	}
	for _, context := range new.Contexts {
		newByName[context.Name], names[context.Name] = *context, true
	}
	var children []*diff.Diff
	for _, name := range slices.Sorted(maps.Keys(names)) {
		before, after := oldByName[name], newByName[name]
		children = append(children, diff.NewDiff(diff.WithKey("kubernetes_context."+name), diff.WithChildren(
			diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(before.Name, after.Name)),
			diff.NewDiff(diff.WithKey("component"), diff.WithStringDiff(before.Component, after.Component)),
		)))
	}
	return diff.NewDiff(diff.WithKey("kubernetes_contexts"), diff.WithChildren(children...))
}

// Diff compares the receiver (desired config) against old (the config currently
// applied to the branch) and returns a hierarchical diff tree covering every
// field an app branch config can express.
func (c *AppBranchConfig) Diff(old *AppBranchConfig) *diff.Diff {
	return branchDiff("branch", old, c)
}

func branchDiff(key string, old, new *AppBranchConfig) *diff.Diff {
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
	children = append(children, diffPublicRepo("public_repo", old.PublicRepo, new.PublicRepo)...)
	if d := diffBranchInstallGroups(old.InstallGroups, new.InstallGroups); d != nil {
		children = append(children, d)
	}
	if d := diffBranchPreview(old.Preview, new.Preview); d != nil {
		children = append(children, d)
	}
	if d := diffBranchRun(old.Run, new.Run); d != nil {
		children = append(children, d)
	}
	children = append(children,
		diff.NewDiff(diff.WithKey("post_deploy_runbooks"), diff.WithOrderedStringSliceDiff(old.PostDeployRunbooks, new.PostDeployRunbooks)),
		diff.NewDiff(diff.WithKey("ignore_changes_regex"), diff.WithStringDiff(old.IgnoreChangesRegex, new.IgnoreChangesRegex)),
		diff.NewDiff(diff.WithKey("send_statuses_on_ignore"), diff.WithBoolDiff(old.SendStatusesOnIgnore, new.SendStatusesOnIgnore)),
	)

	return diff.NewDiff(diff.WithKey(key), diff.WithChildren(children...))
}

func diffBranchInstallGroups(old, new []AppBranchInstallGroupConfig) *diff.Diff {
	if len(old) == 0 && len(new) == 0 {
		return nil
	}

	oldByName := make(map[string]AppBranchInstallGroupConfig, len(old))
	for _, g := range old {
		oldByName[g.Name] = g
	}

	var children []*diff.Diff
	seen := make(map[string]bool, len(new))

	for _, g := range new {
		seen[g.Name] = true
		children = append(children, diffBranchInstallGroup(oldByName[g.Name], g))
	}

	for _, g := range old {
		if !seen[g.Name] {
			children = append(children, diffBranchInstallGroup(g, AppBranchInstallGroupConfig{Name: g.Name}))
		}
	}

	return diff.NewDiff(diff.WithKey("install_groups"), diff.WithChildren(children...))
}

func diffBranchInstallGroup(old, new AppBranchInstallGroupConfig) *diff.Diff {
	name := new.Name
	if name == "" {
		name = old.Name
	}

	children := []*diff.Diff{
		diff.NewDiff(diff.WithKey("order"), diff.WithStringDiff(strconv.Itoa(old.Order), strconv.Itoa(new.Order))),
		diff.NewDiff(diff.WithKey("default"), diff.WithBoolDiff(old.Default, new.Default)),
		diff.NewDiff(diff.WithKey("auto_approve_on_policies_passing"), diff.WithOptionalBoolDiff(
			old.AutoApproveOnPoliciesPassing, new.AutoApproveOnPoliciesPassing,
		)),
	}
	if d := diff.MapDiff("label_selector", old.LabelSelector, new.LabelSelector); d != nil {
		children = append(children, d)
	}

	return diff.NewDiff(diff.WithKey("install_group."+name), diff.WithChildren(children...))
}

func diffBranchPreview(old, new *AppBranchPreviewConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &AppBranchPreviewConfig{Mode: "plan-only"}
	}
	if new == nil {
		new = &AppBranchPreviewConfig{Mode: "plan-only"}
	}

	children := []*diff.Diff{
		diff.NewDiff(diff.WithKey("mode"), diff.WithStringDiff(old.Mode, new.Mode)),
		diff.NewDiff(diff.WithKey("install_id"), diff.WithStringDiff(old.InstallID, new.InstallID)),
		diff.NewDiff(diff.WithKey("install_name"), diff.WithStringDiff(old.InstallName, new.InstallName)),
		diff.NewDiff(diff.WithKey("set_statuses"), withTrueDefaultBoolDiff(old.SetStatuses, new.SetStatuses)),
		diff.NewDiff(diff.WithKey("comment"), withTrueDefaultBoolDiff(old.Comment, new.Comment)),
		diff.NewDiff(diff.WithKey("ignore_drafts"), withTrueDefaultBoolDiff(old.IgnoreDrafts, new.IgnoreDrafts)),
		diff.NewDiff(diff.WithKey("react"), withTrueDefaultBoolDiff(old.React, new.React)),
	}
	if d := diff.MapDiff("label_selector", old.LabelSelector, new.LabelSelector); d != nil {
		children = append(children, d)
	}

	return diff.NewDiff(diff.WithKey("preview"), diff.WithChildren(children...))
}

func diffBranchRun(old, new *AppBranchRunConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &AppBranchRunConfig{}
	}
	if new == nil {
		new = &AppBranchRunConfig{}
	}

	return diff.NewDiff(diff.WithKey("run"), diff.WithChildren(
		diff.NewDiff(diff.WithKey("mode"), diff.WithStringDiff(old.Mode, new.Mode)),
		diff.NewDiff(diff.WithKey("tag_prefix"), diff.WithStringDiff(old.TagPrefix, new.TagPrefix)),
		diff.NewDiff(diff.WithKey("github_label"), diff.WithStringDiff(old.GithubLabel, new.GithubLabel)),
	))
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
	if !reflect.DeepEqual(old, new) {
		if oldTOML, newTOML := sectionTOML(old), sectionTOML(new); oldTOML != newTOML {
			opts = append(opts, diff.WithContentDiff(oldTOML, newTOML))
		}
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
		diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(old.Type, new.Type)),
		diff.NewDiff(diff.WithKey("terraform_version"), diff.WithStringDiff(old.TerraformVersion, new.TerraformVersion)),
		diff.NewDiff(diff.WithKey("runtime"), diff.WithStringDiff(old.Runtime, new.Runtime)),
		diff.NewDiff(diff.WithKey("pulumi_version"), diff.WithStringDiff(old.PulumiVersion, new.PulumiVersion)),
		diff.NewDiff(diff.WithKey("drift_schedule"), diff.WithOptionalStringDiff(old.DriftSchedule, new.DriftSchedule)),
	}
	children = append(children, diffDeployOptions(old.MaxAutoRetries, new.MaxAutoRetries, old.SkipNoops, new.SkipNoops, old.AutoApproveOnPoliciesPassing, new.AutoApproveOnPoliciesPassing)...)
	children = append(children, diffEntityOperationRoles(old.OperationRoles, new.OperationRoles))
	if d := diffTerraformVariablesFiles(old.VariablesFiles, new.VariablesFiles); d != nil {
		children = append(children, d)
	}
	if d := diff.MapDiff("pulumi_config", old.PulumiConfig, new.PulumiConfig); d != nil {
		children = append(children, d)
	}
	children = append(children, diffPublicRepo("public_repo", old.PublicRepo, new.PublicRepo)...)
	children = append(children, diffConnectedRepo("connected_repo", old.ConnectedRepo, new.ConnectedRepo)...)
	if d := diff.MapDiff("env_vars", old.EnvVarMap, new.EnvVarMap); d != nil {
		children = append(children, d)
	}
	if d := diff.MapDiff("vars", old.VarsMap, new.VarsMap); d != nil {
		children = append(children, d)
	}

	oldSandbox, newSandbox := *old, *new
	oldSandbox.References, newSandbox.References = nil, nil
	result := sectionDiff("sandbox", &oldSandbox, &newSandbox, children)
	result.ResourceID = SandboxResourceID
	return result
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
		diff.NewDiff(diff.WithKey("instance_type"), diff.WithStringDiff(old.InstanceType, new.InstanceType)),
		diff.NewDiff(diff.WithKey("runner_api_url"), diff.WithStringDiff(old.RunnerAPIURL, new.RunnerAPIURL)),
		diff.NewDiff(diff.WithKey("public_api_url"), diff.WithStringDiff(old.PublicAPIURL, new.PublicAPIURL)),
		diff.NewDiff(diff.WithKey("init_script_url"), diff.WithStringDiff(old.InitScriptURL, new.InitScriptURL)),
		diff.NewDiff(diff.WithKey("phone_home_script_url"), diff.WithStringDiff(old.PhoneHomeScriptURL, new.PhoneHomeScriptURL)),
		diff.NewDiff(diff.WithKey("image_auto_update"), diff.WithBoolDiff(old.ImageAutoUpdate, new.ImageAutoUpdate)),
		diff.NewDiff(diff.WithKey("image_id"), diff.WithStringDiff(old.ImageID, new.ImageID)),
	}
	if d := diff.MapDiff("env_vars", old.EnvVarMap, new.EnvVarMap); d != nil {
		children = append(children, d)
	}

	result := sectionDiff("runner", old, new, children)
	result.ResourceID = RunnerResourceID
	return result
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
		diff.NewDiff(diff.WithKey("deployment_scope"), diff.WithStringDiff(old.DeploymentScope, new.DeploymentScope)),
		diff.NewDiff(diff.WithKey("ui_definition"), diff.WithStringDiff(old.UIDefinition, new.UIDefinition)),
		diff.NewDiff(diff.WithKey("ui_definition_contents"), diff.WithContentDiff(old.UIDefinitionContents, new.UIDefinitionContents)),
	}

	// Custom nested stacks matched by name
	children = append(children, diffCustomNestedStacks(old.CustomNestedStacks, new.CustomNestedStacks)...)

	result := sectionDiff("stack", old, new, children)
	result.ResourceID = StackResourceID
	return result
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
				diff.NewDiff(diff.WithKey("contents"), diff.WithContentDiff(os.Contents, s.Contents)),
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

	before, after := *old, *new
	before.Inputs, after.Inputs = slices.Clone(old.Inputs), slices.Clone(new.Inputs)
	for _, inputs := range [][]AppInput{before.Inputs, after.Inputs} {
		for i := range inputs {
			inputs[i].Internal = false
		}
	}
	return sectionDiff("inputs", &before, &after, children)
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
	defaultText := func(value any) string {
		encoded, err := ToJSON(value)
		if err != nil {
			return fmt.Sprintf("%T:%#v", value, value)
		}
		return string(encoded)
	}

	diffInput := func(oi, i AppInput) *diff.Diff {
		return diff.NewDiff(
			diff.WithKey("input."+i.Name),
			diff.WithResourceID(InputResourceID(i.Name)),
			diff.WithChildren(
				diff.NewDiff(diff.WithKey("display_name"), diff.WithStringDiff(oi.DisplayName, i.DisplayName)),
				diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(oi.Description, i.Description)),
				diff.NewDiff(diff.WithKey("group"), diff.WithStringDiff(oi.Group, i.Group)),
				diff.NewDiff(diff.WithKey("default"), diff.WithStringDiff(defaultText(oi.Default), defaultText(i.Default))),
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
	children = append(children, diffNamedIAMPolicies(old.NamedPolicies, new.NamedPolicies)...)

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

	oldEnabled, newEnabled := "inherit", "inherit"
	if old.EnabledInStack != nil {
		oldEnabled = strconv.FormatBool(*old.EnabledInStack)
	}
	if new.EnabledInStack != nil {
		newEnabled = strconv.FormatBool(*new.EnabledInStack)
	}
	result := sectionDiff(
		key,
		old,
		new,
		[]*diff.Diff{
			diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(old.Type, new.Type)),
			diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(old.Name, new.Name)),
			diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(old.Description, new.Description)),
			diff.NewDiff(diff.WithKey("display_name"), diff.WithStringDiff(old.DisplayName, new.DisplayName)),
			diff.NewDiff(diff.WithKey("cloud_platform"), diff.WithStringDiff(old.CloudPlatform, new.CloudPlatform)),
			diff.NewDiff(diff.WithKey("permissions_boundary"), diff.WithStringDiff(old.PermissionsBoundary, new.PermissionsBoundary)),
			diff.NewDiff(diff.WithKey("enabled_in_stack"), diff.WithStringDiff(oldEnabled, newEnabled)),
			diff.NewDiff(diff.WithKey("named_policies"), diff.WithStringSliceDiff(NamedPolicyRefNames(old.NamedPolicies), NamedPolicyRefNames(new.NamedPolicies))),
			diffIAMPolicies(old.Policies, new.Policies),
		},
	)
	result.ResourceID = roleNodeID(old, new)
	return result
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

func diffNamedIAMPolicies(old, new []NamedIAMPolicy) []*diff.Diff {
	oldByID := make(map[string]NamedIAMPolicy, len(old))
	for _, p := range old {
		oldByID[p.Name] = p
	}

	var diffs []*diff.Diff
	seen := make(map[string]bool)
	for _, p := range new {
		id := p.Name
		seen[id] = true
		diffs = append(diffs, diffNamedIAMPolicy(oldByID[id], p))
	}
	for _, p := range old {
		id := p.Name
		if seen[id] {
			continue
		}
		diffs = append(diffs, diffNamedIAMPolicy(p, NamedIAMPolicy{}))
	}
	return diffs
}

func diffNamedIAMPolicy(old, new NamedIAMPolicy) *diff.Diff {
	name := new.Name
	if name == "" {
		name = old.Name
	}

	result := diff.NewDiff(
		diff.WithKey("named_policy."+name),
		diff.WithChildren(
			diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(old.Name, new.Name)),
			diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(old.Description, new.Description)),
			diff.NewDiff(diff.WithKey("contents"), diff.WithContentDiff(old.Contents, new.Contents)),
		),
	)
	result.ResourceID = NamedPolicyResourceID(name)
	return result
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
				diff.NewDiff(diff.WithKey("apply_on_teardown"), diff.WithBoolDiff(AppliesOnTeardown(op.ApplyOnTeardown), AppliesOnTeardown(p.ApplyOnTeardown))),
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
		result := sectionDiff(
			"secret."+s.Name,
			os,
			s,
			[]*diff.Diff{
				diff.NewDiff(diff.WithKey("display_name"), diff.WithStringDiff(os.DisplayName, s.DisplayName)),
				diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(os.Description, s.Description)),
				diff.NewDiff(diff.WithKey("required"), diff.WithBoolDiff(os.Required, s.Required)),
				diff.NewDiff(diff.WithKey("auto_generate"), diff.WithBoolDiff(os.AutoGenerate, s.AutoGenerate)),
				diff.NewDiff(diff.WithKey("format"), diff.WithStringDiff(os.Format, s.Format)),
				diff.NewDiff(diff.WithKey("default"), diff.WithStringDiff(os.Default, s.Default)),
				diff.NewDiff(diff.WithKey("kubernetes_sync"), diff.WithOptionalBoolDiff(os.KubernetesSync, s.KubernetesSync)),
				diff.NewDiff(diff.WithKey("kubernetes_secret_namespace"), diff.WithStringDiff(os.KubernetesSecretNamespace, s.KubernetesSecretNamespace)),
				diff.NewDiff(diff.WithKey("kubernetes_secret_name"), diff.WithStringDiff(os.KubernetesSecretName, s.KubernetesSecretName)),
				diffSecretTargets(os.KubernetesSyncTargets, s.KubernetesSyncTargets),
			},
		)
		result.ResourceID = SecretResourceID(s.Name)
		return result
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
	// The parser defaults a missing [operation_roles] config to an empty
	// matrix config, so nil must diff as that default rather than a zero
	// value — otherwise every config shows a spurious type-only change.
	parserDefault := &OperationRolesConfig{Type: OperationRuleConfigTypeMatrix}
	if old == nil {
		old = parserDefault
	}
	if new == nil {
		new = parserDefault
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

	return diff.NewDiff(
		diff.WithKey("operation_roles"),
		diff.WithResourceID(OperationRolesResourceID),
		diff.WithChildren(children...),
	)
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
		diff.NewDiff(diff.WithKey("toggleable"), diff.WithOptionalBoolDiff(old.Toggleable, new.Toggleable)),
		diff.NewDiff(diff.WithKey("default_enabled"), diff.WithOptionalBoolDiff(old.DefaultEnabled, new.DefaultEnabled)),
		diff.NewDiff(diff.WithKey("kubernetes_context"), diff.WithStringDiff(old.KubernetesContext, new.KubernetesContext)),
		diffEntityOperationRoles(old.OperationRoles, new.OperationRoles),
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

	oldComponent, newComponent := *old, *new
	oldComponent.Checksum, newComponent.Checksum = "", ""
	oldComponent.References, newComponent.References = nil, nil
	oldComponent.Dependencies = slices.Clone(old.Dependencies)
	newComponent.Dependencies = slices.Clone(new.Dependencies)
	sort.Strings(oldComponent.Dependencies)
	sort.Strings(newComponent.Dependencies)
	result := sectionDiff("component."+new.Name, &oldComponent, &newComponent, children)
	result.ResourceID = ComponentResourceID(new.Name)
	return result
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
	diffs = append(diffs, diffDeployOptions(old.MaxAutoRetries, new.MaxAutoRetries, old.SkipNoops, new.SkipNoops, old.AutoApproveOnPoliciesPassing, new.AutoApproveOnPoliciesPassing)...)
	diffs = append(diffs, diffComponentHealth(old.Health, new.Health)...)
	diffs = append(diffs,
		diff.NewDiff(diff.WithKey("chart_name"), diff.WithStringDiff(old.ChartName, new.ChartName)),
		diff.NewDiff(diff.WithKey("namespace"), diff.WithStringDiff(old.Namespace, new.Namespace)),
		diff.NewDiff(diff.WithKey("storage_driver"), diff.WithStringDiff(old.StorageDriver, new.StorageDriver)),
		diff.NewDiff(diff.WithKey("take_ownership"), diff.WithBoolDiff(old.TakeOwnership, new.TakeOwnership)),
		diff.NewDiff(diff.WithKey("skip_crds"), diff.WithBoolDiff(old.SkipCRDs, new.SkipCRDs)),
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
	diffs = append(diffs, diffDeployOptions(old.MaxAutoRetries, new.MaxAutoRetries, old.SkipNoops, new.SkipNoops, old.AutoApproveOnPoliciesPassing, new.AutoApproveOnPoliciesPassing)...)
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
	diffs = append(diffs, diffComponentTeardown("teardown", old.Teardown, new.Teardown)...)

	return diffs
}

func diffComponentTeardown(key string, old, new *ComponentTeardownConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &ComponentTeardownConfig{}
	}
	if new == nil {
		new = &ComponentTeardownConfig{}
	}
	return []*diff.Diff{
		diff.NewDiff(diff.WithKey(key), diff.WithChildren(
			diff.NewDiff(diff.WithKey("retain"), diff.WithStringSliceDiff(old.Retain, new.Retain)),
		)),
	}
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
	diffs = append(diffs, diffDeployOptions(old.MaxAutoRetries, new.MaxAutoRetries, old.SkipNoops, new.SkipNoops, old.AutoApproveOnPoliciesPassing, new.AutoApproveOnPoliciesPassing)...)
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
	diffs = append(diffs, diffAzureACR(old.AzureACRImageConfig, new.AzureACRImageConfig)...)
	diffs = append(diffs, diffImageVerification(old.Verification, new.Verification)...)

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
			diff.NewDiff(diff.WithKey("update_policy"), diff.WithStringDiff(old.UpdatePolicy, new.UpdatePolicy)),
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
			diff.NewDiff(diff.WithKey("workload_identity_provider"), diff.WithStringDiff(old.WorkloadIdentityProvider, new.WorkloadIdentityProvider)),
			diff.NewDiff(diff.WithKey("update_policy"), diff.WithStringDiff(old.UpdatePolicy, new.UpdatePolicy)),
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
			diff.NewDiff(diff.WithKey("update_policy"), diff.WithStringDiff(old.UpdatePolicy, new.UpdatePolicy)),
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
	diffs = append(diffs, diffDeployOptions(old.MaxAutoRetries, new.MaxAutoRetries, old.SkipNoops, new.SkipNoops, old.AutoApproveOnPoliciesPassing, new.AutoApproveOnPoliciesPassing)...)
	diffs = append(diffs, diffComponentHealth(old.Health, new.Health)...)
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
		diff.NewDiff(diff.WithKey("cmd"), diff.WithOrderedStringSliceDiff(old.Cmd, new.Cmd)),
		diff.NewDiff(diff.WithKey("args"), diff.WithOrderedStringSliceDiff(old.Args, new.Args)),
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
	diffs = append(diffs, diffDeployOptions(old.MaxAutoRetries, new.MaxAutoRetries, old.SkipNoops, new.SkipNoops, old.AutoApproveOnPoliciesPassing, new.AutoApproveOnPoliciesPassing)...)
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
			empty := &Install{}
			d, _ := empty.Diff(i)
			if d != nil {
				d.Key = i.Name
				children = append(children, d)
			}
		}
	}

	return diff.NewDiff(diff.WithKey("installs"), diff.WithChildren(children...))
}

// --- Installs Config ---

func diffInstallsConfig(old, new *InstallsConfig) *diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &InstallsConfig{}
	}
	if new == nil {
		new = &InstallsConfig{}
	}

	var children []*diff.Diff
	children = append(children, diffConnectedRepo("connected_repo", old.ConnectedRepo, new.ConnectedRepo)...)
	children = append(children, diffPublicRepo("public_repo", old.PublicRepo, new.PublicRepo)...)

	return sectionDiff("installs_config", old, new, children)
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
		diff.NewDiff(diff.WithKey("enable_kube_config"), withTrueDefaultBoolDiff(old.EnableKubeConfig, new.EnableKubeConfig)),
		diff.NewDiff(diff.WithKey("dependencies"), diff.WithStringSliceDiff(old.Dependencies, new.Dependencies)),
		diff.NewDiff(diff.WithKey("is_healthcheck"), diff.WithBoolDiff(old.IsHealthcheck, new.IsHealthcheck)),
		diff.NewDiff(diff.WithKey("image"), diff.WithStringDiff(old.Image, new.Image)),
		diff.NewDiff(diff.WithKey("kubernetes_context"), diff.WithStringDiff(old.KubernetesContext, new.KubernetesContext)),
	)

	if d := diff.MapDiff("labels", old.Labels, new.Labels); d != nil {
		children = append(children, d)
	}

	// Steps matched by name
	children = append(children, diffActionSteps(old.Steps, new.Steps)...)

	// Triggers matched by index
	children = append(children, diffActionTriggers(old.Triggers, new.Triggers)...)

	// Persisted configs can predate deterministic dependency ordering.
	oldAction, newAction := *old, *new
	for _, action := range []*ActionConfig{&oldAction, &newAction} {
		action.References = nil
		action.Steps = slices.Clone(action.Steps)
		for i, step := range action.Steps {
			stepCopy := *step
			stepCopy.References = nil
			action.Steps[i] = &stepCopy
		}
	}
	oldAction.Dependencies = slices.Clone(old.Dependencies)
	newAction.Dependencies = slices.Clone(new.Dependencies)
	sort.Strings(oldAction.Dependencies)
	sort.Strings(newAction.Dependencies)
	result := sectionDiff("action."+new.Name, &oldAction, &newAction, children)
	result.ResourceID = ActionResourceID(new.Name)
	return result
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
					diff.NewDiff(diff.WithKey("index"), diff.WithStringDiff(strconv.FormatInt(old[i].Index, 10), strconv.FormatInt(new[i].Index, 10))),
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

// --- Runbooks ---

func diffRunbooks(old, new []*RunbookConfig) *diff.Diff {
	if len(old) == 0 && len(new) == 0 {
		return nil
	}

	oldByName := make(map[string]*RunbookConfig, len(old))
	for _, r := range old {
		oldByName[r.Name] = r
	}

	var children []*diff.Diff
	seen := make(map[string]bool)

	for _, r := range new {
		seen[r.Name] = true
		if or, ok := oldByName[r.Name]; ok {
			children = append(children, diffRunbook(or, r))
		} else {
			children = append(children, diffRunbook(&RunbookConfig{}, r))
		}
	}

	for _, r := range old {
		if !seen[r.Name] {
			children = append(children, diffRunbook(r, &RunbookConfig{}))
		}
	}

	return diff.NewDiff(diff.WithKey("runbooks"), diff.WithChildren(children...))
}

func diffRunbook(old, new *RunbookConfig) *diff.Diff {
	var children []*diff.Diff
	children = append(children,
		diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(old.Name, new.Name)),
		diff.NewDiff(diff.WithKey("description"), diff.WithStringDiff(old.Description, new.Description)),
		diff.NewDiff(diff.WithKey("readme"), diff.WithStringDiff(old.Readme, new.Readme)),
		diff.NewDiff(diff.WithKey("dependencies"), diff.WithStringSliceDiff(old.Dependencies, new.Dependencies)),
		diff.NewDiff(diff.WithKey("triggers"), diff.WithContentDiff(toTOMLString(map[string]any{"triggers": old.Triggers}), toTOMLString(map[string]any{"triggers": new.Triggers}))),
		diff.NewDiff(diff.WithKey("input"), diff.WithContentDiff(toTOMLString(map[string]any{"input": old.Inputs}), toTOMLString(map[string]any{"input": new.Inputs}))),
	)

	if d := diff.MapDiff("labels", old.Labels, new.Labels); d != nil {
		children = append(children, d)
	}

	oldOrder, newOrder := make([]string, len(old.Steps)), make([]string, len(new.Steps))
	for i, step := range old.Steps {
		oldOrder[i] = step.Name
	}
	for i, step := range new.Steps {
		newOrder[i] = step.Name
	}
	children = append(children, diff.NewDiff(
		diff.WithKey("step_order"),
		diff.WithOrderedStringSliceDiff(oldOrder, newOrder),
	))
	children = append(children, diffRunbookSteps(old.Steps, new.Steps)...)

	name := new.Name
	if name == "" {
		name = old.Name
	}
	return diff.NewDiff(diff.WithKey("runbook."+name), diff.WithChildren(children...))
}

func diffRunbookSteps(old, new []*RunbookStepConfig) []*diff.Diff {
	oldByName := make(map[string]*RunbookStepConfig, len(old))
	for _, s := range old {
		oldByName[s.Name] = s
	}

	var diffs []*diff.Diff
	seen := make(map[string]bool)

	diffStep := func(os, s *RunbookStepConfig) *diff.Diff {
		name := s.Name
		if name == "" {
			name = os.Name
		}
		children := []*diff.Diff{
			diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(os.Name, s.Name)),
			diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(string(os.Type), string(s.Type))),
			diff.NewDiff(diff.WithKey("plan_only"), diff.WithBoolDiff(os.PlanOnly, s.PlanOnly)),
			diff.NewDiff(diff.WithKey("component_name"), diff.WithStringDiff(os.ComponentName, s.ComponentName)),
			diff.NewDiff(diff.WithKey("deploy_dependents"), diff.WithBoolDiff(os.DeployDependents, s.DeployDependents)),
			diff.NewDiff(diff.WithKey("tear_down_dependents"), diff.WithBoolDiff(os.TearDownDependents, s.TearDownDependents)),
			diff.NewDiff(diff.WithKey("skip_component_deploys"), diff.WithBoolDiff(os.SkipComponentDeploys, s.SkipComponentDeploys)),
			diff.NewDiff(diff.WithKey("action_name"), diff.WithStringDiff(os.ActionName, s.ActionName)),
			diff.NewDiff(diff.WithKey("command"), diff.WithStringDiff(os.Command, s.Command)),
			diff.NewDiff(diff.WithKey("inline_contents"), diff.WithContentDiff(os.InlineContents, s.InlineContents)),
			diff.NewDiff(diff.WithKey("timeout"), diff.WithStringDiff(os.Timeout, s.Timeout)),
			diff.NewDiff(diff.WithKey("role"), diff.WithStringDiff(os.Role, s.Role)),
			diff.NewDiff(diff.WithKey("trigger"), diff.WithStringDiff(os.Trigger, s.Trigger)),
			diff.NewDiff(diff.WithKey("event_types"), diff.WithStringSliceDiff(os.EventTypes, s.EventTypes)),
			diff.NewDiff(diff.WithKey("filters"), diff.WithContentDiff(toTOMLString(map[string]any{"filters": os.Filters}), toTOMLString(map[string]any{"filters": s.Filters}))),
			diff.NewDiff(diff.WithKey("match_all"), diff.WithBoolDiff(os.MatchAll, s.MatchAll)),
		}
		if d := diff.MapDiff("env_vars", os.EnvVarMap, s.EnvVarMap); d != nil {
			children = append(children, d)
		}
		return diff.NewDiff(
			diff.WithKey("step."+name),
			diff.WithChildren(children...),
		)
	}

	for _, s := range new {
		seen[s.Name] = true
		if os, ok := oldByName[s.Name]; ok {
			diffs = append(diffs, diffStep(os, s))
		} else {
			diffs = append(diffs, diffStep(&RunbookStepConfig{}, s))
		}
	}

	for _, s := range old {
		if !seen[s.Name] {
			diffs = append(diffs, diffStep(s, &RunbookStepConfig{}))
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

func diffDeployOptions(oldRetries, newRetries *int, oldSkip, newSkip, oldApprove, newApprove *bool) []*diff.Diff {
	before, after := 0, 0
	if oldRetries != nil {
		before = *oldRetries
	}
	if newRetries != nil {
		after = *newRetries
	}
	return []*diff.Diff{
		diff.NewDiff(diff.WithKey("max_auto_retries"), diff.WithStringDiff(strconv.Itoa(before), strconv.Itoa(after))),
		diff.NewDiff(diff.WithKey("skip_noops"), diff.WithOptionalBoolDiff(oldSkip, newSkip)),
		diff.NewDiff(diff.WithKey("auto_approve_on_policies_passing"), diff.WithOptionalBoolDiff(oldApprove, newApprove)),
	}
}

func diffEntityOperationRoles(old, new []EntityOperationRole) *diff.Diff {
	var children []*diff.Diff
	for i := 0; i < max(len(old), len(new)); i++ {
		var before, after EntityOperationRole
		if i < len(old) {
			before = old[i]
		}
		if i < len(new) {
			after = new[i]
		}
		children = append(children, diff.NewDiff(diff.WithKey(fmt.Sprintf("operation_role.%d", i)), diff.WithChildren(
			diff.NewDiff(diff.WithKey("operation"), diff.WithStringDiff(string(before.Operation), string(after.Operation))),
			diff.NewDiff(diff.WithKey("role"), diff.WithStringDiff(before.RoleName, after.RoleName)),
		)))
	}
	return diff.NewDiff(diff.WithKey("operation_roles"), diff.WithChildren(children...))
}

func diffIAMPolicies(old, new []AppAWSIAMPolicy) *diff.Diff {
	var children []*diff.Diff
	for i := 0; i < max(len(old), len(new)); i++ {
		var before, after AppAWSIAMPolicy
		if i < len(old) {
			before = old[i]
		}
		if i < len(new) {
			after = new[i]
		}
		children = append(children, diff.NewDiff(diff.WithKey(fmt.Sprintf("policy.%d", i)), diff.WithChildren(
			diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(before.Name, after.Name)),
			diff.NewDiff(diff.WithKey("managed_policy_name"), diff.WithStringDiff(before.ManagedPolicyName, after.ManagedPolicyName)),
			diff.NewDiff(diff.WithKey("contents"), diff.WithContentDiff(before.Contents, after.Contents)),
			diff.NewDiff(diff.WithKey("gcp_permissions"), diff.WithStringSliceDiff(before.GCPPermissions, after.GCPPermissions)),
			diff.NewDiff(diff.WithKey("gcp_predefined_role"), diff.WithStringDiff(before.GCPPredefinedRole, after.GCPPredefinedRole)),
			diff.NewDiff(diff.WithKey("azure_actions"), diff.WithStringSliceDiff(before.AzureActions, after.AzureActions)),
			diff.NewDiff(diff.WithKey("azure_built_in_roles"), diff.WithStringSliceDiff(before.AzureBuiltInRoles, after.AzureBuiltInRoles)),
		)))
	}
	return diff.NewDiff(diff.WithKey("policies"), diff.WithChildren(children...))
}

func diffSecretTargets(old, new []*KubernetesSyncTarget) *diff.Diff {
	var children []*diff.Diff
	for i := 0; i < max(len(old), len(new)); i++ {
		before, after := &KubernetesSyncTarget{}, &KubernetesSyncTarget{}
		if i < len(old) && old[i] != nil {
			before = old[i]
		}
		if i < len(new) && new[i] != nil {
			after = new[i]
		}
		children = append(children, diff.NewDiff(diff.WithKey(fmt.Sprintf("target.%d", i)), diff.WithChildren(
			diff.NewDiff(diff.WithKey("namespaces"), diff.WithStringSliceDiff(before.Namespaces, after.Namespaces)),
			diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(before.Name, after.Name)),
			diff.NewDiff(diff.WithKey("key"), diff.WithStringDiff(before.Key, after.Key)),
		)))
	}
	return diff.NewDiff(diff.WithKey("kubernetes_sync_targets"), diff.WithChildren(children...))
}

func diffComponentHealth(old, new *ComponentHealthConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &ComponentHealthConfig{}
	}
	if new == nil {
		new = &ComponentHealthConfig{}
	}
	children := []*diff.Diff{
		diff.NewDiff(diff.WithKey("enabled"), withTrueDefaultBoolDiff(old.Enabled, new.Enabled)),
		diff.NewDiff(diff.WithKey("stabilization_window"), diff.WithStringDiff(old.StabilizationWindow, new.StabilizationWindow)),
		diff.NewDiff(diff.WithKey("block_deploy"), diff.WithOptionalBoolDiff(old.BlockDeploy, new.BlockDeploy)),
		diff.NewDiff(diff.WithKey("required_checks"), diff.WithStringSliceDiff(old.RequiredChecks, new.RequiredChecks)),
	}
	for i := 0; i < max(len(old.Probes), len(new.Probes)); i++ {
		var before, after ComponentHealthProbeConfig
		if i < len(old.Probes) {
			before = old.Probes[i]
		}
		if i < len(new.Probes) {
			after = new.Probes[i]
		}
		children = append(children, diff.NewDiff(diff.WithKey(fmt.Sprintf("probe.%d", i)), diff.WithChildren(
			diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(before.Type, after.Type)),
			diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(before.Name, after.Name)),
			diff.NewDiff(diff.WithKey("url"), diff.WithStringDiff(before.URL, after.URL)),
			diff.NewDiff(diff.WithKey("command"), diff.WithOrderedStringSliceDiff(before.Command, after.Command)),
		)))
	}
	return []*diff.Diff{diff.NewDiff(diff.WithKey("health"), diff.WithChildren(children...))}
}

func diffAzureACR(old, new *AzureACRConfig) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &AzureACRConfig{}
	}
	if new == nil {
		new = &AzureACRConfig{}
	}
	return []*diff.Diff{diff.NewDiff(diff.WithKey("azure_acr"), diff.WithChildren(
		diff.NewDiff(diff.WithKey("image_url"), diff.WithStringDiff(old.ImageURL, new.ImageURL)),
		diff.NewDiff(diff.WithKey("tag"), diff.WithStringDiff(old.Tag, new.Tag)),
		diff.NewDiff(diff.WithKey("registry_url"), diff.WithStringDiff(old.RegistryURL, new.RegistryURL)),
		diff.NewDiff(diff.WithKey("tenant_id"), diff.WithStringDiff(old.TenantID, new.TenantID)),
		diff.NewDiff(diff.WithKey("client_id"), diff.WithStringDiff(old.ClientID, new.ClientID)),
		diff.NewDiff(diff.WithKey("client_secret_name"), diff.WithStringDiff(old.ClientSecretName, new.ClientSecretName)),
		diff.NewDiff(diff.WithKey("client_certificate_name"), diff.WithStringDiff(old.ClientCertificateName, new.ClientCertificateName)),
		diff.NewDiff(diff.WithKey("update_policy"), diff.WithStringDiff(old.UpdatePolicy, new.UpdatePolicy)),
	))}
}

func diffImageVerification(old, new *signature.Verification) []*diff.Diff {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		old = &signature.Verification{}
	}
	if new == nil {
		new = &signature.Verification{}
	}
	children := []*diff.Diff{diff.NewDiff(diff.WithKey("require_signature"), diff.WithBoolDiff(old.RequireSignature, new.RequireSignature))}
	for i := 0; i < max(len(old.Authorities), len(new.Authorities)); i++ {
		var before, after signature.Authority
		if i < len(old.Authorities) {
			before = old.Authorities[i]
		}
		if i < len(new.Authorities) {
			after = new.Authorities[i]
		}
		children = append(children, diff.NewDiff(diff.WithKey(fmt.Sprintf("authority.%d", i)), diff.WithChildren(
			diff.NewDiff(diff.WithKey("type"), diff.WithStringDiff(string(before.Type), string(after.Type))),
			diff.NewDiff(diff.WithKey("issuer"), diff.WithStringDiff(before.Issuer, after.Issuer)),
			diff.NewDiff(diff.WithKey("subject"), diff.WithStringDiff(before.Subject, after.Subject)),
			diff.NewDiff(diff.WithKey("subject_regexp"), diff.WithStringDiff(before.SubjectRegexp, after.SubjectRegexp)),
			diff.NewDiff(diff.WithKey("public_key"), diff.WithContentDiff(before.PublicKey, after.PublicKey)),
		)))
	}
	return []*diff.Diff{diff.NewDiff(diff.WithKey("verification"), diff.WithChildren(children...))}
}

func withTrueDefaultBoolDiff(old, new *bool) diff.DiffOption {
	return diff.WithBoolDiff(old == nil || *old, new == nil || *new)
}
