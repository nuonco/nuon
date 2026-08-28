package customermanaged

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	planpkg "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func TestVirtualIDsAreDeterministic(t *testing.T) {
	firstInstall := VirtualInstallID("app-test")
	secondInstall := VirtualInstallID("app-test")

	require.Equal(t, firstInstall, secondInstall)
	require.Equal(t, "vinst"+firstInstall[5:], firstInstall)
	require.Len(t, firstInstall, 21)
	require.Equal(t, virtualID("vic", "app-test:component-a"), virtualID("vic", "app-test:component-a"))
}

func TestCompileInstallSetsSandboxTerraformWorkspaceID(t *testing.T) {
	install := compileInstall("org-test", "app-test", "cfg-test", "vinst1234")
	require.NotEmpty(t, install.InstallSandbox.TerraformWorkspace.ID)
	require.Equal(t, virtualID("vtfw", "vinst1234"), install.InstallSandbox.TerraformWorkspace.ID)
}

func TestCompileComponentOrderUsesDependenciesAndNameTieBreak(t *testing.T) {
	connections := []app.ComponentConfigConnection{
		{ID: "ccc-c", ComponentID: "c", ComponentName: "charlie", ComponentDependencyIDs: []string{"a"}},
		{ID: "ccc-b", ComponentID: "b", ComponentName: "bravo"},
		{ID: "ccc-a", ComponentID: "a", ComponentName: "alpha"},
	}

	ordered, err := compileComponentOrder(connections)
	require.NoError(t, err)
	require.Equal(t, []string{"alpha", "bravo", "charlie"}, []string{ordered[0].ComponentName, ordered[1].ComponentName, ordered[2].ComponentName})
}

func TestCompileStateNestsNuonAndInjectsInputPlaceholders(t *testing.T) {
	raw := []byte(`{"value":"{{.nuon.install_stack.outputs.vpc_id}}","other":"{{.nuon.sandbox.outputs.endpoint}}","zone":"{{.nuon.install.sandbox.outputs.nuon_dns}}"}`)
	_, data, err := compileState("org-test", "app-test", "vinst-test", raw, []customermanaged.InputSpec{{Name: "region"}}, nil, false, nil, nil, "")
	require.NoError(t, err)

	nuon := data["nuon"].(map[string]any)
	install := nuon["install"].(map[string]any)
	require.Equal(t, "vinst-test", install["id"])
	require.Equal(t, "__NUON_INPUT_region__", install["inputs"].(map[string]any)["region"])
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_vpc_id__", nuon["install_stack"].(map[string]any)["outputs"].(map[string]any)["vpc_id"])
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_SANDBOX_endpoint__", nuon["sandbox"].(map[string]any)["outputs"].(map[string]any)["endpoint"])
	installSandboxOutputs := install["sandbox"].(map[string]any)["outputs"].(map[string]any)
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_SANDBOX_nuon_dns__", installSandboxOutputs["nuon_dns"])
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_SANDBOX_endpoint__", installSandboxOutputs["endpoint"])
}

func TestCompileStateSeedsBuiltinStackOutputsWithoutReferences(t *testing.T) {
	_, data, err := compileState("org-test", "app-test", "vinst-test", []byte(`{"plain":"no template references"}`), nil, nil, false, nil, nil, "")
	require.NoError(t, err)

	outputs := data["nuon"].(map[string]any)["install_stack"].(map[string]any)["outputs"].(map[string]any)
	for _, key := range builtinStackOutputKeys {
		require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_"+key+"__", outputs[key], key)
	}
}

func TestCompileStateSeedsCloudAccountByStackType(t *testing.T) {
	_, data, err := compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, nil, false, nil, nil, "")
	require.NoError(t, err)
	cloud := data["nuon"].(map[string]any)["cloud_account"].(map[string]any)
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_region__", cloud["aws"].(map[string]any)["region"])
	require.Nil(t, cloud["azure"])
	require.Nil(t, cloud["gcp"])

	out, err := render.RenderTextV2(`{{ .nuon.cloud_account.aws.region }}`, data)
	require.NoError(t, err)
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_region__", out)

	_, data, err = compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, nil, false, nil, nil, app.StackTypeAzure)
	require.NoError(t, err)
	cloud = data["nuon"].(map[string]any)["cloud_account"].(map[string]any)
	require.Nil(t, cloud["aws"])
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_resource_group_location__", cloud["azure"].(map[string]any)["location"])

	_, data, err = compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, nil, false, nil, nil, app.StackTypeGCP)
	require.NoError(t, err)
	cloud = data["nuon"].(map[string]any)["cloud_account"].(map[string]any)
	require.Nil(t, cloud["aws"])
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_project_id__", cloud["gcp"].(map[string]any)["project_id"])
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_region__", cloud["gcp"].(map[string]any)["region"])
}

func TestCompileStateSeedsRoleARNMapsFromRenderedNames(t *testing.T) {
	_, data, err := compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, nil, false,
		[]string{"{{.nuon.install.id}}-sandbox-break-glass", ""}, []string{"{{.nuon.install.id}}-custom"}, "")
	require.NoError(t, err)

	outputs := data["nuon"].(map[string]any)["install_stack"].(map[string]any)["outputs"].(map[string]any)
	require.Equal(t, map[string]any{
		"vinst-test-sandbox-break-glass": "__NUON_CUSTOMER_MANAGED_STACK_break_glass_role_arns_0__",
	}, outputs["break_glass_role_arns"])
	require.Equal(t, map[string]any{
		"vinst-test-custom": "__NUON_CUSTOMER_MANAGED_STACK_custom_role_arns_0__",
	}, outputs["custom_role_arns"])
}

func TestCompileStateSeedsEmptyRoleARNMapsWithoutRoles(t *testing.T) {
	_, data, err := compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, nil, false, nil, nil, "")
	require.NoError(t, err)

	outputs := data["nuon"].(map[string]any)["install_stack"].(map[string]any)["outputs"].(map[string]any)
	require.Equal(t, map[string]any{}, outputs["break_glass_role_arns"])
	require.Equal(t, map[string]any{}, outputs["custom_role_arns"])
}

func TestCompileStateRoleARNMapsSupportTemplateMapOperations(t *testing.T) {
	_, data, err := compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, nil, false,
		[]string{"{{.nuon.install.id}}-sandbox-break-glass"}, nil, "")
	require.NoError(t, err)

	out, err := render.RenderTextV2(`{{ if gt (len .nuon.install_stack.outputs.break_glass_role_arns) 0 }}{{ index (values .nuon.install_stack.outputs.break_glass_role_arns) 0 }}{{ end }}`, data)
	require.NoError(t, err)
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_break_glass_role_arns_0__", out)

	out, err = render.RenderTextV2(`{{ if gt (len .nuon.install_stack.outputs.custom_role_arns) 0 }}unexpected{{ end }}`, data)
	require.NoError(t, err)
	require.Equal(t, "", out)
}

func TestCompileActionTemplatesSkipsGitSourcedActionsWithWarning(t *testing.T) {
	cfg := &app.AppConfig{ActionWorkflowConfigs: []app.ActionWorkflowConfig{
		{
			ID:               "acc-git",
			ActionWorkflowID: "acw-git",
			ActionWorkflow:   app.ActionWorkflow{Name: "git-action"},
			Steps: []app.ActionWorkflowStepConfig{
				{ID: "acs-git", Name: "fetch", PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: "org/repo"}},
			},
		},
		{
			ID:               "acc-inline",
			ActionWorkflowID: "acw-inline",
			ActionWorkflow:   app.ActionWorkflow{Name: "inline-action"},
			Steps: []app.ActionWorkflowStepConfig{
				{ID: "acs-inline", Name: "run", InlineContents: "echo hello"},
			},
		},
	}}
	stack := compileStack(&app.AppConfig{}, "vinst-test")
	role := &operationroles.RoleSelection{RoleARN: "arn:aws:iam::123:role/maintenance"}
	report := &QualificationReport{}

	result, err := compileActionTemplates(cfg, planpkg.NewPlanner(validator.New()), zap.NewNop(), "vinst-test", map[string]any{"nuon": map[string]any{}}, stack, role, report)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "acw-inline", result[0].ID)
	require.Len(t, report.Warnings, 1)
	require.Equal(t, "action.git_source_excluded", report.Warnings[0].Code)
	require.Equal(t, "action:acw-git", report.Warnings[0].Member)
}

func TestCompileStackSuppliesAllAWSRolePlaceholders(t *testing.T) {
	stack := compileStack(&app.AppConfig{}, "vinst-test")

	aws := stack.InstallStackOutputs.AWSStackOutputs
	require.NotNil(t, aws)
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_provision_iam_role_arn__", aws.ProvisionIAMRoleARN)
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_deprovision_iam_role_arn__", aws.DeprovisionIAMRoleARN)
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_STACK_maintenance_iam_role_arn__", aws.MaintenanceIAMRoleARN)
	require.Equal(t, aws.MaintenanceIAMRoleARN, compileRoleSelection(&app.AppConfig{}, stack).RoleARN)
}

func TestUnknownComponentOutputReferenceAddsQualificationViolation(t *testing.T) {
	report := &QualificationReport{}
	value := "{{.nuon.components.database.outputs.endpoint}}"
	cfg := &app.AppConfig{SandboxConfig: app.AppSandboxConfig{Variables: map[string]*string{"endpoint": &value}}}

	_, err := CompilePlanEnvelope(context.Background(), nil, nil, "org-test", "app-test", cfg, "sandbox-build", nil, nil, report)
	require.ErrorContains(t, err, "unknown component")
	require.ErrorContains(t, err, "database.endpoint")
	require.False(t, report.Qualified)
	require.Len(t, report.Violations, 1)
	require.Equal(t, "template.component_output_unknown_component", report.Violations[0].Code)
}

func TestExtractComponentOutputRefsSkipsReadmeAndDeduplicates(t *testing.T) {
	data := []byte(`{"readme":"curl https://{{.nuon.components.api_gateway.outputs.api_gateway.domain_name_id}}/widgets/7","vars":"{{.nuon.components.database.outputs.endpoint}}","again":"{{.nuon.components.database.outputs.endpoint}}","nested":"{{.nuon.components.lambda_function.outputs.lambda_function.lambda_function_arn}} "}`)

	refs := extractComponentOutputRefs(data)
	require.Equal(t, []componentOutputRefKey{
		{Component: "database", Path: "endpoint"},
		{Component: "lambda_function", Path: "lambda_function.lambda_function_arn"},
	}, refs)
}

func TestValidateComponentOutputRefs(t *testing.T) {
	connections := []app.ComponentConfigConnection{{ComponentName: "database"}}
	refs := []componentOutputRefKey{{Component: "database", Path: "endpoint"}}
	require.NoError(t, validateComponentOutputRefs(refs, connections, nil))

	report := &QualificationReport{}
	refs = append(refs, componentOutputRefKey{Component: "missing", Path: "url"})
	err := validateComponentOutputRefs(refs, connections, report)
	require.ErrorContains(t, err, "missing.url")
	require.NotContains(t, err.Error(), "database.endpoint")
	require.Len(t, report.Violations, 1)
}

func TestCompileStateSeedsNestedSandboxOutputPlaceholders(t *testing.T) {
	raw := []byte(`{"domain":"{{.nuon.install.sandbox.outputs.nuon_dns.public_domain.name }}","zone":"{{.nuon.install.sandbox.outputs.nuon_dns.public_domain.zone_id}}","flat":"{{.nuon.sandbox.outputs.endpoint}}"}`)
	_, data, err := compileState("org-test", "app-test", "vinst-test", raw, nil, nil, false, nil, nil, "")
	require.NoError(t, err)

	outputs := data["nuon"].(map[string]any)["sandbox"].(map[string]any)["outputs"].(map[string]any)
	publicDomain := outputs["nuon_dns"].(map[string]any)["public_domain"].(map[string]any)
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_SANDBOX_nuon_dns_public_domain_name__", publicDomain["name"])
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_SANDBOX_nuon_dns_public_domain_zone_id__", publicDomain["zone_id"])
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_SANDBOX_endpoint__", outputs["endpoint"])
}

func TestActionStepTemplateReferencesSeedCompileState(t *testing.T) {
	region := "{{.nuon.install.sandbox.outputs.account.region}}"
	lambdaARN := "{{.nuon.components.lambda.outputs.function.arn}}"
	cfg := app.AppConfig{ActionWorkflowConfigs: []app.ActionWorkflowConfig{{
		Steps: []app.ActionWorkflowStepConfig{{
			EnvVars: pgtype.Hstore{"REGION": &region, "LAMBDA_ARN": &lambdaARN},
		}},
	}}}
	references := templateReferenceData(&cfg)
	require.Equal(t, []componentOutputRefKey{{Component: "lambda", Path: "function.arn"}}, extractComponentOutputRefs(references))

	_, data, err := compileState("org-test", "app-test", "vinst-test", references, nil, nil, false, nil, nil, "")
	require.NoError(t, err)

	outputs := data["nuon"].(map[string]any)["install"].(map[string]any)["sandbox"].(map[string]any)["outputs"].(map[string]any)
	account := outputs["account"].(map[string]any)
	require.Equal(t, "__NUON_CUSTOMER_MANAGED_SANDBOX_account_region__", account["region"])
}

func TestCompileStateSeedsComponentOutputPlaceholders(t *testing.T) {
	refs := []componentOutputRefKey{
		{Component: "certificate", Path: "public_domain_certificate_arn"},
		{Component: "lambda_function", Path: "lambda_function.lambda_function_arn"},
	}
	_, data, err := compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, refs, false, nil, nil, "")
	require.NoError(t, err)

	components := data["nuon"].(map[string]any)["components"].(map[string]any)
	cert := components["certificate"].(map[string]any)["outputs"].(map[string]any)
	require.Equal(t, customermanaged.ComponentOutputPlaceholder("certificate", "public_domain_certificate_arn"), cert["public_domain_certificate_arn"])
	lambda := components["lambda_function"].(map[string]any)["outputs"].(map[string]any)["lambda_function"].(map[string]any)
	require.Equal(t, customermanaged.ComponentOutputPlaceholder("lambda_function", "lambda_function.lambda_function_arn"), lambda["lambda_function_arn"])
}

func TestCompileStateKeepsComponentPlaceholdersOutOfEmbeddedState(t *testing.T) {
	refs := []componentOutputRefKey{
		{Component: "certificate", Path: "public_domain_certificate_arn"},
		{Component: "lambda_function", Path: "lambda_function.lambda_function_arn"},
	}
	st, data, err := compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, refs, false, nil, nil, "")
	require.NoError(t, err)

	require.Empty(t, st.Components)
	raw, err := json.Marshal(st)
	require.NoError(t, err)
	for _, ref := range refs {
		token := customermanaged.ComponentOutputPlaceholder(ref.Component, ref.Path)
		require.NotContains(t, string(raw), token)
		require.Contains(t, mustJSON(t, data), token)
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return string(raw)
}

func TestNormalizeComponentTokenPaddingTrimsWholeStringTokens(t *testing.T) {
	ref := componentOutputRefKey{Component: "certificate", Path: "public_domain_certificate_arn"}
	token := ref.token()
	raw := mustJSON(t, map[string]any{
		"padded":   token + " ",
		"both":     "  " + token + "\t",
		"exact":    token,
		"embedded": "prefix " + token + " suffix",
		"nested":   map[string]any{"vars": []any{" " + token}},
		"plain":    " unrelated ",
	})

	normalized, err := normalizeComponentTokenPadding(json.RawMessage(raw), []componentOutputRefKey{ref})
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(normalized, &decoded))
	require.Equal(t, token, decoded["padded"])
	require.Equal(t, token, decoded["both"])
	require.Equal(t, token, decoded["exact"])
	require.Equal(t, "prefix "+token+" suffix", decoded["embedded"])
	require.Equal(t, token, decoded["nested"].(map[string]any)["vars"].([]any)[0])
	require.Equal(t, " unrelated ", decoded["plain"])
}

func TestNormalizeComponentTokenPaddingNoRefsIsIdentity(t *testing.T) {
	raw := json.RawMessage(`{"a":" b "}`)
	normalized, err := normalizeComponentTokenPadding(raw, nil)
	require.NoError(t, err)
	require.Equal(t, raw, normalized)
}

func TestSeedComponentOutputsRejectsConflictingPaths(t *testing.T) {
	_, err := seedComponentOutputs([]componentOutputRefKey{
		{Component: "lambda_function", Path: "lambda_function"},
		{Component: "lambda_function", Path: "lambda_function.lambda_function_arn"},
	})
	require.ErrorContains(t, err, "conflicting references under lambda_function")
}

func TestNeedsSandboxCluster(t *testing.T) {
	terraform := app.ComponentConfigConnection{Type: app.ComponentTypeTerraformModule}
	helm := app.ComponentConfigConnection{Type: app.ComponentTypeHelmChart}
	helmWithContext := app.ComponentConfigConnection{Type: app.ComponentTypeHelmChart, KubernetesContextName: "data-cluster"}
	manifest := app.ComponentConfigConnection{Type: app.ComponentTypeKubernetesManifest}

	require.False(t, needsSandboxCluster([]app.ComponentConfigConnection{terraform}, []byte(`{}`)))
	require.True(t, needsSandboxCluster([]app.ComponentConfigConnection{terraform, helm}, []byte(`{}`)))
	require.True(t, needsSandboxCluster([]app.ComponentConfigConnection{manifest}, []byte(`{}`)))
	require.False(t, needsSandboxCluster([]app.ComponentConfigConnection{terraform, helmWithContext}, []byte(`{}`)))
	require.True(t, needsSandboxCluster([]app.ComponentConfigConnection{terraform}, []byte(`{"v":"{{.nuon.sandbox.outputs.cluster.name}}"}`)))
	require.True(t, needsSandboxCluster([]app.ComponentConfigConnection{terraform}, []byte(`{"v":"{{.nuon.install.sandbox.outputs.cluster.endpoint}}"}`)))
	require.False(t, needsSandboxCluster([]app.ComponentConfigConnection{terraform}, []byte(`{"v":"{{.nuon.sandbox.outputs.cluster_name}}"}`)))
}

func TestContextClusterRefsSeedsUsedContextsOnly(t *testing.T) {
	cfg := &app.AppConfig{
		KubernetesContextsConfig: app.AppKubernetesContextsConfig{Contexts: []app.AppKubernetesContextConfig{
			{Name: "used", SourceComponentName: "eks"},
			{Name: "unused", SourceComponentName: "other"},
		}},
	}
	connections := []app.ComponentConfigConnection{{Type: app.ComponentTypeHelmChart, KubernetesContextName: "used"}}

	refs := contextClusterRefs(cfg, connections)
	require.Equal(t, []componentOutputRefKey{
		{Component: "eks", Path: "cluster.name"},
		{Component: "eks", Path: "cluster.endpoint"},
		{Component: "eks", Path: "cluster.certificate_authority_data"},
	}, refs)

	cfg.StackConfig = app.AppStackConfig{Type: app.StackTypeAzure}
	refs = contextClusterRefs(cfg, connections)
	require.Equal(t, []componentOutputRefKey{
		{Component: "eks", Path: "cluster.name"},
		{Component: "eks", Path: "cluster.host"},
		{Component: "eks", Path: "cluster.cluster_ca_certificate"},
	}, refs)

	require.Nil(t, contextClusterRefs(cfg, []app.ComponentConfigConnection{{Type: app.ComponentTypeTerraformModule}}))
}

func TestMergeComponentRefsDeduplicatesAndSorts(t *testing.T) {
	merged := mergeComponentRefs(
		[]componentOutputRefKey{{Component: "b", Path: "x"}, {Component: "a", Path: "y"}},
		[]componentOutputRefKey{{Component: "a", Path: "y"}, {Component: "a", Path: "x"}},
	)
	require.Equal(t, []componentOutputRefKey{
		{Component: "a", Path: "x"},
		{Component: "a", Path: "y"},
		{Component: "b", Path: "x"},
	}, merged)
}

func TestCompileDeployPlanClusterInfo(t *testing.T) {
	planner := planpkg.NewPlanner(validator.New())
	logger := zap.NewNop()
	cfg := &app.AppConfig{ID: "apc-test", AppID: "app-test"}
	stack := &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{AWSStackOutputs: &app.AWSStackOutputs{Region: "us-east-1"}}}
	role := &operationroles.RoleSelection{RoleARN: "arn:aws:iam::123:role/deploy"}
	deploy := &app.InstallDeploy{ID: "vdep-test", InstallID: "vinst-test", ComponentName: "lambda", ComponentID: "cmp-test", Type: app.InstallDeployTypeApply}
	build := &app.ComponentBuild{ComponentConfigConnection: app.ComponentConfigConnection{
		Type:                           app.ComponentTypeTerraformModule,
		TerraformModuleComponentConfig: &app.TerraformModuleComponentConfig{},
	}}
	src := &configs.OCIRegistryRepository{}

	t.Run("cluster-less sandbox omits cluster info", func(t *testing.T) {
		st, stateData, err := compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, nil, false, nil, nil, "")
		require.NoError(t, err)
		result, jobType, err := compileDeployPlan(planner, logger, cfg, stack, st, stateData, deploy, build, src, role)
		require.NoError(t, err)
		require.Equal(t, "terraform-deploy", jobType)
		require.Nil(t, result.TerraformDeployPlan.ClusterInfo)
	})

	t.Run("cluster sandbox carries cluster info with cloud auth", func(t *testing.T) {
		st, stateData, err := compileState("org-test", "app-test", "vinst-test", []byte(`{}`), nil, nil, true, nil, nil, "")
		require.NoError(t, err)
		result, _, err := compileDeployPlan(planner, logger, cfg, stack, st, stateData, deploy, build, src, role)
		require.NoError(t, err)
		info := result.TerraformDeployPlan.ClusterInfo
		require.NotNil(t, info)
		require.Equal(t, "__NUON_CUSTOMER_MANAGED_CLUSTER_id__", info.ID)
		require.Equal(t, "__NUON_CUSTOMER_MANAGED_CLUSTER_endpoint__", info.Endpoint)
		require.Equal(t, "__NUON_CUSTOMER_MANAGED_CLUSTER_ca_data__", info.CAData)
		require.NotNil(t, info.AWSAuth)
	})
}

func TestCompositeJSONContainsExactlyOnePlanKey(t *testing.T) {
	raw, err := compositeJSON(&plantypes.CompositePlan{DeployPlan: &plantypes.DeployPlan{InstallID: "vinst-test"}})
	require.NoError(t, err)

	var decoded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Equal(t, []string{"deploy_plan"}, mapKeys(decoded))
}

func mapKeys(values map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
