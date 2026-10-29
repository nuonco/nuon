package config

import (
	"encoding/json"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/config/refs"
)

func TestAppConfigDiff_RootFields(t *testing.T) {
	t.Parallel()
	fields := map[string]struct {
		section string
		set     func(*AppConfig, string)
	}{
		"Version": {"version", func(c *AppConfig, v string) {
			c.Version = "1"
			if len(v) > 6 {
				c.Version = "2"
			}
		}},
		"Description":     {"description", func(c *AppConfig, v string) { c.Description = v }},
		"DisplayName":     {"display_name", func(c *AppConfig, v string) { c.DisplayName = v }},
		"SlackWebhookURL": {"slack_webhook_url", func(c *AppConfig, v string) { c.SlackWebhookURL = "https://example.com/" + v }},
		"Readme":          {"readme", func(c *AppConfig, v string) { c.Readme = v }},
		"LabelColors":     {"label_colors", func(c *AppConfig, v string) { c.LabelColors = map[string]string{"env": "#" + v[:6]} }},
		"DefaultLabels":   {"default_labels", func(c *AppConfig, v string) { c.DefaultLabels = map[string]string{"env": v} }},
		"Branch":          {"branch", func(c *AppConfig, v string) { c.Branch = &AppBranchConfig{Name: "stable", IgnoreChangesRegex: v} }},
		"Branches":        {"branches", func(c *AppConfig, v string) { c.Branches = []*AppBranchConfig{{Name: "stable", IgnoreChangesRegex: v}} }},
		"Inputs":          {"inputs", func(c *AppConfig, v string) { c.Inputs.Inputs[0].Description = v }},
		"Sandbox":         {"sandbox", func(c *AppConfig, v string) { c.Sandbox.VarsMap = map[string]string{"region": v} }},
		"Runner":          {"runner", func(c *AppConfig, v string) { c.Runner.EnvVarMap = map[string]string{"REGION": v} }},
		"Permissions":     {"permissions", func(c *AppConfig, v string) { c.Permissions.MaintenanceRole.Description = v }},
		"Policies":        {"policies", func(c *AppConfig, v string) { c.Policies.Policies[0].Contents = v }},
		"Secrets":         {"secrets", func(c *AppConfig, v string) { c.Secrets.Secrets[0].Description = v }},
		"BreakGlass": {"break_glass", func(c *AppConfig, v string) {
			c.BreakGlass = &BreakGlass{Roles: []*AppAWSIAMRole{{Name: "emergency", Description: v}}}
		}},
		"Stack": {"stack", func(c *AppConfig, v string) { c.Stack.Description = v }},
		"OperationRoles": {"operation_roles", func(c *AppConfig, v string) {
			c.OperationRoles = &OperationRolesConfig{Type: OperationRuleConfigTypeMatrix, RuleMatrix: []*OperationRoleRule{{Principal: "nuon::component:*", Operation: OperationDeploy, RoleName: v}}}
		}},
		"KubernetesContexts": {"kubernetes_contexts", func(c *AppConfig, v string) {
			component := "terraform"
			if len(v) > 6 {
				component = "cluster"
			}
			c.Components = ComponentList{{Name: "terraform", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}}, {Name: "cluster", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}}}
			c.KubernetesContexts = &KubernetesContextsConfig{Contexts: []*KubernetesContext{{Name: "compute", Component: component}}}
		}},
		"Components": {"components", func(c *AppConfig, v string) {
			c.Components = ComponentList{{Name: "terraform", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0", VarsMap: map[string]string{"region": v}}}}
		}},
		"Installs": {"installs", func(c *AppConfig, v string) {
			c.Installs = []*Install{{Name: "acme", Labels: map[string]string{"env": v}}}
		}},
		"InstallsConfig": {"installs_config", func(c *AppConfig, v string) {
			c.InstallsConfig = &InstallsConfig{PublicRepo: &PublicRepoConfig{Repo: "https://github.com/acme/config", Directory: v}}
		}},
		"Actions": {"actions", func(c *AppConfig, v string) {
			c.Actions = []*ActionConfig{{Name: "inspect", Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo " + v}}}}
		}},
		"Runbooks": {"runbooks", func(c *AppConfig, v string) { c.Runbooks[0].Description = v }},
		"Triggers": {"triggers", func(c *AppConfig, v string) {
			c.Triggers = &TriggersConfig{Rules: []*TriggerRuleConfig{{Name: "release", Trigger: v}}}
		}},
		"SourceArchive": {"", func(c *AppConfig, v string) {
			c.SourceArchive = &SourceArchive{Files: map[string]string{"nuon.toml": v}}
		}},
	}
	typ := reflect.TypeFor[AppConfig]()
	require.Len(t, fields, typ.NumField(), "every new root field needs an authored/ignored contract")
	for _, field := range reflect.VisibleFields(typ) {
		_, ok := fields[field.Name]
		require.True(t, ok, "unclassified field %s", field.Name)
	}
	baseline := baseConfig()
	baseline.Stack = &StackConfig{Name: "acme"}
	baseline.Inputs = &AppInputConfig{Inputs: []AppInput{{Name: "region", Type: "string"}}}
	baseline.Secrets = &SecretsConfig{Secrets: []*AppSecret{{Name: "api_key"}}}
	baseline.Permissions = &PermissionsConfig{MaintenanceRole: &AppAWSIAMRole{Name: "maintenance"}}
	baseline.Policies = &PoliciesConfig{Policies: []AppPolicy{{Name: "acme", Type: AppPolicyTypeKubernetesCluster, Engine: AppPolicyEngineOPA}}}
	baseline.Runbooks = []*RunbookConfig{{Name: "release", Steps: []*RunbookStepConfig{{Name: "inspect", Type: RunbookStepTypeAction, Command: "echo ready"}}}}
	encoded, err := json.Marshal(baseline)
	require.NoError(t, err)
	value := "123456"
	for name, field := range fields {
		var old, next AppConfig
		require.NoError(t, json.Unmarshal(encoded, &old))
		require.NoError(t, json.Unmarshal(encoded, &next))
		field.set(&old, value)
		field.set(&next, "abcdef"+value)
		for _, result := range []*diff.Diff{next.Diff(&old), old.Diff(&next)} {
			var changed []string
			for _, child := range result.Children {
				if child.DirectSummary().HasChanged {
					changed = append(changed, child.Key)
				}
			}
			if field.section == "" {
				require.Empty(t, changed, "%s is authored provenance, not semantic configuration", name)
			} else {
				require.Equal(t, []string{field.section}, changed, "%s must be visible and local", name)
			}
		}
	}
}

func TestAppConfigDiff_RunbookFields(t *testing.T) {
	fields := map[string]func(*RunbookConfig, string){
		"Name":        func(c *RunbookConfig, v string) { c.Name = "release_" + v },
		"Description": func(c *RunbookConfig, v string) { c.Description = v },
		"Readme":      func(c *RunbookConfig, v string) { c.Readme = v },
		"Labels":      func(c *RunbookConfig, v string) { c.Labels = map[string]string{"env": v} },
		"Triggers": func(c *RunbookConfig, v string) {
			c.Triggers = []RunbookTrigger{{Type: RunbookTriggerTypeCron, CronSchedule: v + " * * * *"}}
		},
		"Steps": func(c *RunbookConfig, v string) { c.Steps[0].Command = "echo " + v },
		"Inputs": func(c *RunbookConfig, v string) {
			c.Inputs = []RunbookInput{{Name: "region", Type: "string", Default: v}}
		},
		"Dependencies":        func(c *RunbookConfig, v string) { c.Dependencies = []string{v} },
		"References":          nil,
		"DeprecationWarnings": nil,
	}
	typ := reflect.TypeFor[RunbookConfig]()
	require.Len(t, fields, typ.NumField())
	for _, field := range reflect.VisibleFields(typ) {
		_, ok := fields[field.Name]
		require.True(t, ok, "unclassified runbook field %s", field.Name)
	}
	minute := 7
	for field, set := range fields {
		if set == nil {
			continue
		}
		makeConfig := func(v string) *AppConfig {
			c := &AppConfig{Runbooks: []*RunbookConfig{{Name: "release", Steps: []*RunbookStepConfig{{Name: "inspect", Type: RunbookStepTypeAction, Command: "echo ready"}}}}}
			set(c.Runbooks[0], v)
			return c
		}
		old, next := makeConfig(strconv.Itoa(minute)), makeConfig(strconv.Itoa(minute+30))
		for _, result := range []*diff.Diff{next.Diff(old), old.Diff(next)} {
			require.True(t, findChild(result, "runbooks").DirectSummary().HasChanged, "%s must not be missed", field)
			require.NotEmpty(t, result.FormatChanged(""))
		}
	}
}

func TestAppConfigDiff_RunbookStepFields(t *testing.T) {
	fields := map[string]string{
		"Name": "identity", "Type": "orchestration", "PlanOnly": "orchestration",
		"ComponentName": "orchestration", "DeployDependents": "orchestration", "TearDownDependents": "orchestration",
		"DeployDependenciesLegacy": "legacy-alias", "SkipComponentDeploys": "orchestration", "ActionName": "orchestration",
		"Command": "execution", "InlineContents": "execution", "EnvVarMap": "execution", "Timeout": "execution", "Role": "execution",
		"Trigger": "orchestration", "EventTypes": "orchestration", "Filters": "orchestration", "MatchAll": "orchestration",
		"References": "derived",
	}
	typ := reflect.TypeFor[RunbookStepConfig]()
	require.Len(t, fields, typ.NumField())
	for _, field := range reflect.VisibleFields(typ) {
		_, ok := fields[field.Name]
		require.True(t, ok, "unclassified runbook step field %s", field.Name)
	}
	value := "acme"
	for name, category := range fields {
		if category == "derived" || category == "legacy-alias" {
			continue
		}
		makeConfig := func(changed bool) *AppConfig {
			s := &RunbookStepConfig{Name: "inspect", Type: RunbookStepTypeAction, Command: "echo ready"}
			switch name {
			case "Type":
				s.Type, s.Command = RunbookStepTypeSandboxReprovision, ""
			case "PlanOnly", "ComponentName", "DeployDependents":
				s.Type, s.ComponentName, s.Command = RunbookStepTypeComponentDeploy, "api", ""
			case "TearDownDependents":
				s.Type, s.ComponentName, s.Command = RunbookStepTypeComponentTearDown, "api", ""
			case "SkipComponentDeploys":
				s.Type, s.Command = RunbookStepTypeSandboxReprovision, ""
			case "Trigger", "EventTypes", "Filters", "MatchAll":
				s.Type, s.Command, s.Trigger, s.EventTypes = RunbookStepTypeWaitForEvent, "", "release", []string{"ready"}
			case "InlineContents":
				s.Command, s.InlineContents = "", "echo ready"
			case "ActionName":
				s.Command, s.ActionName = "", "inspect"
			}
			if changed {
				f := reflect.ValueOf(s).Elem().FieldByName(name)
				switch f.Kind() {
				case reflect.Bool:
					f.SetBool(true)
				case reflect.String:
					f.SetString("changed_" + value)
					if name == "Type" {
						s.Type = RunbookStepTypeSandboxDeprovision
					}
					if name == "Timeout" {
						s.Timeout = "30s"
					}
				case reflect.Map:
					f.Set(reflect.ValueOf(map[string]string{"REGION": value}))
				case reflect.Slice:
					if name == "EventTypes" {
						s.EventTypes = []string{"deployed"}
					} else {
						s.Filters = []TriggerFilterConfig{{Op: "eq", Path: "$.region", Value: value}}
					}
				}
			}
			return &AppConfig{Runbooks: []*RunbookConfig{{Name: "release", Steps: []*RunbookStepConfig{s}}}}
		}
		old, next := makeConfig(false), makeConfig(true)
		for _, result := range []*diff.Diff{next.Diff(old), old.Diff(next)} {
			require.True(t, findChild(result, "runbooks").DirectSummary().HasChanged, "%s must not be missed", name)
		}
	}
}

func TestAppConfigDiff_JobSequenceFieldProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		values := drawDistinctStrings(t, hegel.Text(), 2, 6)
		makeConfig := func(v []string) *AppConfig {
			return &AppConfig{Components: ComponentList{{Name: "job", Type: JobComponentType, Job: &JobComponentConfig{Cmd: slices.Clone(v), Args: slices.Clone(v)}}}}
		}
		reversed := slices.Clone(values)
		slices.Reverse(reversed)
		for _, pair := range [][2][]string{{values, reversed}, {values, values}, {nil, nil}, {nil, {""}}, {nil, values}, {values, nil}, {{"a, b"}, {"a", "b"}}} {
			result := makeConfig(pair[1]).Diff(makeConfig(pair[0])).FindResource(ComponentResourceID("job"))
			operation := diff.OpChange
			switch {
			case slices.Equal(pair[0], pair[1]):
				operation = diff.OpNoop
			case len(pair[0]) == 0:
				operation = diff.OpAdd
			case len(pair[1]) == 0:
				operation = diff.OpRemove
			}
			for _, field := range []string{"cmd", "args"} {
				require.Equal(t, operation, findChild(result, field).Diff.Op, "%s is an ordered sequence with presence", field)
			}
		}
	})
}

func TestAppConfigDiff_TrueDefaultBooleans(t *testing.T) {
	name := "acme"
	for label, tc := range map[string]struct {
		old, next *bool
		changed   bool
	}{
		"both omitted":     {},
		"omitted to false": {nil, boolPtr(false), true},
		"omitted to true":  {nil, boolPtr(true), false},
		"false to omitted": {boolPtr(false), nil, true},
		"false unchanged":  {boolPtr(false), boolPtr(false), false},
		"false to true":    {boolPtr(false), boolPtr(true), true},
		"true to omitted":  {boolPtr(true), nil, false},
		"true to false":    {boolPtr(true), boolPtr(false), true},
		"true unchanged":   {boolPtr(true), boolPtr(true), false},
	} {
		t.Run(label, func(t *testing.T) {
			makeConfig := func(value *bool) *AppConfig {
				return &AppConfig{
					Actions:    []*ActionConfig{{Name: name, EnableKubeConfig: value, Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo ready"}}}},
					Components: ComponentList{{Name: name, Type: HelmChartComponentType, HelmChart: &HelmChartComponentConfig{Health: &ComponentHealthConfig{Enabled: value}}}},
					Policies:   &PoliciesConfig{Policies: []AppPolicy{{Name: name, Type: AppPolicyTypeKubernetesCluster, Engine: AppPolicyEngineOPA, Contents: "package acme\n", ApplyOnTeardown: value}}},
					Branch:     &AppBranchConfig{Name: name, Preview: &AppBranchPreviewConfig{SetStatuses: value, Comment: value, IgnoreDrafts: value, React: value}},
				}
			}
			result := makeConfig(tc.next).Diff(makeConfig(tc.old))
			preview := findChild(findChild(result, "branch"), "preview")
			for _, field := range []*diff.Diff{
				findChild(result.FindResource(ActionResourceID(name)), "enable_kube_config"),
				findChild(findChild(result.FindResource(ComponentResourceID(name)), "health"), "enabled"),
				findChild(findChild(findChild(result, "policies"), "policy."+name), "apply_on_teardown"),
				findChild(preview, "set_statuses"),
				findChild(preview, "comment"),
				findChild(preview, "ignore_drafts"),
				findChild(preview, "react"),
			} {
				require.NotNil(t, field)
				require.Equal(t, tc.changed, field.DirectSummary().HasChanged, "%s: omission defaults to true", field.Key)
			}
		})
	}
}

func TestAppConfigDiff_DerivedFields(t *testing.T) {
	value := "acme"
	reference := []refs.Ref{{Type: refs.RefTypeInputs, Name: "region", Input: "nuon.inputs.inputs.region"}}
	for field, set := range map[string]func(*AppConfig){
		"component checksum":      func(c *AppConfig) { c.Components[0].Checksum = value },
		"component source file":   func(c *AppConfig) { c.Components[0].SourceFile = value },
		"component references":    func(c *AppConfig) { c.Components[0].References = reference },
		"sandbox references":      func(c *AppConfig) { c.Sandbox.References = reference },
		"action references":       func(c *AppConfig) { c.Actions[0].References = reference },
		"action step references":  func(c *AppConfig) { c.Actions[0].Steps[0].References = reference },
		"runbook references":      func(c *AppConfig) { c.Runbooks[0].References = reference },
		"runbook warnings":        func(c *AppConfig) { c.Runbooks[0].DeprecationWarnings = []string{value} },
		"runbook step references": func(c *AppConfig) { c.Runbooks[0].Steps[0].References = reference },
	} {
		makeConfig := func() *AppConfig {
			c := baseConfig()
			c.Components = ComponentList{{Name: "api", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{VarsMap: map[string]string{"region": "{{ .nuon.inputs.inputs.region }}"}}}}
			c.Actions = []*ActionConfig{{Name: "inspect", Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo {{ .nuon.inputs.inputs.region }}"}}}}
			c.Runbooks = []*RunbookConfig{{Name: "release", Steps: []*RunbookStepConfig{{Name: "inspect", Type: RunbookStepTypeAction, Command: "echo {{ .nuon.inputs.inputs.region }}"}}}}
			return c
		}
		old, next := makeConfig(), makeConfig()
		set(next)
		before, err := json.Marshal(next)
		require.NoError(t, err)
		require.False(t, next.Diff(old).DirectSummary().HasChanged, "%s is derived, not authored", field)
		require.False(t, old.Diff(next).DirectSummary().HasChanged, "%s reverse comparison", field)
		after, err := json.Marshal(next)
		require.NoError(t, err)
		require.Equal(t, before, after, "comparison must not erase provenance")
	}
}
