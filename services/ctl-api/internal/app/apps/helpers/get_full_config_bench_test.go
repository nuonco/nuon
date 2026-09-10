package helpers_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/afterquery"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
)

// BenchmarkGetFullAppConfig measures full-config GORM allocations using in-memory
// SQL rows; no database server is required. Setup, warmup and graph checks are
// excluded from timing. This measures allocations, not SQL correctness or latency.
// Run from the repository root:
//
//	go test ./services/ctl-api/internal/app/apps/helpers/ -run '^$' -bench BenchmarkGetFullAppConfig -benchmem
func BenchmarkGetFullAppConfig(b *testing.B) {
	for _, actions := range []int{8, 120} {
		b.Run(fmt.Sprintf("actions_%d", actions), func(b *testing.B) {
			h := fullConfigBenchmarkHelper(b, actions)
			ctx := context.Background()
			got, err := h.GetFullAppConfig(ctx, "config", false)
			require.NoError(b, err)
			checkFullConfigBenchmark(b, got, "config", "helm", actions)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				got, err = h.GetFullAppConfig(ctx, "config", false)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			checkFullConfigBenchmark(b, got, "config", "helm", actions)
		})
	}
}

func TestGetFullAppConfigFixture(t *testing.T) {
	for _, actions := range []int{8, 120} {
		t.Run(fmt.Sprintf("actions_%d", actions), func(t *testing.T) {
			h := fullConfigBenchmarkHelper(t, actions)
			got, err := h.GetFullAppConfig(context.Background(), "config", false)
			require.NoError(t, err)
			checkFullConfigBenchmark(t, got, "config", "helm", actions)
		})
	}
}

func fullConfigBenchmarkHelper(t testing.TB, actions int) *appshelpers.Helpers {
	t.Helper()
	tables := make(fullConfigDriver)
	add := func(table, columns string, values ...driver.Value) {
		data := tables[table]
		data.columns = strings.Fields(columns)
		require.Len(t, values, len(data.columns))
		data.rows = append(data.rows, values)
		tables[table] = data
	}
	add("app_configs_view_v3", "id app_id component_ids status version", "config", "app", "{helm,terraform,docker,kubernetes,image,job,pulumi}", "active", int64(1))
	for _, table := range []string{"app_runner_configs", "app_stack_configs", "app_input_configs", "app_secrets_configs", "app_permissions_configs", "app_break_glass_configs", "app_policies_configs", "app_operation_role_configs", "app_kubernetes_contexts_configs", "app_sandbox_configs"} {
		add(table, "app_config_id id", "config", table)
	}
	add("app_input_groups", "app_input_config_id id name", "app_input_configs", "group", "general")
	add("app_inputs", "app_input_config_id id app_input_group_id name", "app_input_configs", "input", "group", "region")
	add("app_secret_configs", "app_secrets_config_id id name", "app_secrets_configs", "secret", "api_token")
	add("app_secret_kubernetes_sync_targets", "app_secret_config_id id name key", "secret", "target", "api-token", "token")
	for _, role := range []struct{ owner, typ string }{
		{"app_permissions_configs", string(app.AWSIAMRoleTypeRunnerProvision)},
		{"app_break_glass_configs", string(app.AWSIAMRoleTypeBreakGlass)},
	} {
		add("app_awsiam_role_configs", "owner_id id owner_type type name", role.owner, role.typ, role.owner, role.typ, "read-only")
		add("app_awsiam_policy_configs", "app_awsiam_role_config_id id name contents", role.typ, role.typ+"-policy", "read", `{"Version":"2012-10-17","Statement":[]}`)
	}
	add("app_policy_configs", "app_policies_config_id id name contents", "app_policies_configs", "policy", "allow", "package main\ndefault allow = true")
	add("app_operation_role_rules", "app_operation_role_config_id id operation role", "app_operation_role_configs", "rule", "deploy", "read-only")
	add("app_kubernetes_context_configs", "app_kubernetes_contexts_config_id id name source_component_id", "app_kubernetes_contexts_configs", "context", "primary", "terraform")
	for _, c := range []struct {
		id  string
		typ app.ComponentType
	}{
		{"helm", app.ComponentTypeHelmChart}, {"terraform", app.ComponentTypeTerraformModule},
		{"docker", app.ComponentTypeDockerBuild}, {"kubernetes", app.ComponentTypeKubernetesManifest},
		{"image", app.ComponentTypeExternalImage}, {"job", app.ComponentTypeJob}, {"pulumi", app.ComponentTypePulumi},
	} {
		add("components", "id name type", c.id, c.id, string(c.typ))
		add("component_config_connections_view_v1", "app_config_id id component_id", "config", c.id+"-connection", c.id)
	}
	add("helm_component_configs", "component_config_connection_id id chart_name", "helm-connection", "helm-config", "my-chart")
	add("terraform_module_component_configs", "component_config_connection_id id version", "terraform-connection", "terraform-config", "latest")
	add("docker_build_component_configs", "component_config_connection_id id dockerfile", "docker-connection", "docker-config", "Dockerfile")
	add("kubernetes_manifest_component_configs", "component_config_connection_id id manifest", "kubernetes-connection", "kubernetes-config", "apiVersion: v1\nkind: ConfigMap\n")
	add("external_image_component_configs", "component_config_connection_id id image_url", "image-connection", "image-config", "nginx")
	add("job_component_configs", "component_config_connection_id id image_url", "job-connection", "job-config", "ubuntu")
	add("pulumi_component_configs", "component_config_connection_id id runtime", "pulumi-connection", "pulumi-config", "go")
	for _, source := range []string{"app_sandbox_configs", "helm-config", "terraform-config", "docker-config", "pulumi-config"} {
		add("public_git_vcs_configs", "component_config_id id repo branch", source, source+"-vcs", "https://example.com/config", "main")
	}
	tables["connected_github_vcs_configs"] = fullConfigTable{columns: []string{"component_config_id", "id"}}
	for i := 0; i < actions; i++ {
		id := fmt.Sprintf("action-%03d", i)
		add("action_workflow_configs", "app_config_id id", "config", id)
		columns := "action_workflow_config_id id app_config_id type cron_schedule component_id"
		add("action_workflow_trigger_configs", columns, id, id+"-manual", "config", "manual", "", nil)
		if i%10 == 0 {
			add("action_workflow_trigger_configs", columns, id, id+"-cron", "config", "cron", "17 */2 * * *", nil)
		}
		if i%6 == 0 {
			add("action_workflow_trigger_configs", columns, id, id+"-component", "config", "post-deploy-component", "", "helm")
		}
	}
	sqlDB := sql.OpenDB(tables)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.Use(views.NewViewsPlugin(psql.AllModels())))
	require.NoError(t, db.Use(afterquery.NewAfterQueryPlugin()))
	return appshelpers.New(appshelpers.Params{DB: db, Cfg: &internal.Config{}, L: zap.NewNop()})
}

type fullConfigTable struct {
	columns []string
	rows    [][]driver.Value
}

// Each fixture's first column is its lookup ID. This driver supplies matching rows
// for GORM's bound IDs; it does not implement SQL semantics or database validation.
type fullConfigDriver map[string]fullConfigTable

func (d fullConfigDriver) Driver() driver.Driver                        { return d }
func (d fullConfigDriver) Open(string) (driver.Conn, error)             { return d, nil }
func (d fullConfigDriver) Connect(context.Context) (driver.Conn, error) { return d, nil }
func (d fullConfigDriver) Close() error                                 { return nil }
func (d fullConfigDriver) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("fixture driver is read-only")
}
func (d fullConfigDriver) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("fixture driver does not prepare statements")
}

func (d fullConfigDriver) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	_, from, ok := strings.Cut(query, ` FROM "`)
	if !ok {
		return nil, fmt.Errorf("unexpected fixture query: %s", query)
	}
	name, _, _ := strings.Cut(from, `"`)
	table, ok := d[name]
	if !ok {
		return nil, fmt.Errorf("missing fixture for table %q", name)
	}
	rows := &fullConfigRows{columns: table.columns}
	for _, row := range table.rows {
		for _, arg := range args {
			if row[0] == arg.Value {
				rows.data = append(rows.data, row)
				break
			}
		}
	}
	return rows, nil
}

type fullConfigRows struct {
	columns []string
	data    [][]driver.Value
}

func (r *fullConfigRows) Columns() []string { return r.columns }
func (r *fullConfigRows) Close() error      { return nil }
func (r *fullConfigRows) Next(dest []driver.Value) error {
	if len(r.data) == 0 {
		return io.EOF
	}
	copy(dest, r.data[0])
	r.data = r.data[1:]
	return nil
}

func checkFullConfigBenchmark(b testing.TB, got *app.AppConfig, configID, componentID string, actions int) {
	b.Helper()
	require.Equal(b, configID, got.ID)
	require.NotEmpty(b, got.RunnerConfig.ID)
	require.NotEmpty(b, got.StackConfig.ID)
	require.NotNil(b, got.SandboxConfig.PublicGitVCSConfig)
	require.NotEmpty(b, got.SandboxConfig.PublicGitVCSConfig.ID)
	require.Len(b, got.InputConfig.AppInputGroups, 1)
	require.Len(b, got.InputConfig.AppInputs, 1)
	require.Equal(b, "region", got.InputConfig.AppInputs[0].Name)
	require.Len(b, got.SecretsConfig.Secrets, 1)
	require.Len(b, got.SecretsConfig.Secrets[0].KubernetesSyncTargets, 1)
	require.Equal(b, "token", got.SecretsConfig.Secrets[0].KubernetesSyncTargets[0].Key)
	require.Len(b, got.PermissionsConfig.Roles, 1)
	require.Len(b, got.PermissionsConfig.Roles[0].Policies, 1)
	require.NotEmpty(b, got.PermissionsConfig.ProvisionRole.ID, "AfterQuery must run")
	require.Len(b, got.BreakGlassConfig.Roles, 1)
	require.Len(b, got.BreakGlassConfig.Roles[0].Policies, 1)
	require.Len(b, got.PoliciesConfig.Policies, 1)
	require.Equal(b, "allow", got.PoliciesConfig.Policies[0].Name)
	require.NotNil(b, got.OperationRoleConfig)
	require.Len(b, got.OperationRoleConfig.Rules, 1)
	require.Equal(b, app.OperationDeploy, got.OperationRoleConfig.Rules[0].Operation)
	require.NotNil(b, got.KubernetesContextsConfig)
	require.Len(b, got.KubernetesContextsConfig.Contexts, 1)
	require.Equal(b, "primary", got.KubernetesContextsConfig.Contexts[0].Name)
	require.Len(b, got.ComponentIDs, 7)
	require.Len(b, got.ComponentConfigConnections, 7)
	seen := make(map[app.ComponentType]bool)
	for _, c := range got.ComponentConfigConnections {
		require.Equal(b, c.ComponentID, c.Component.ID)
		require.NotEmpty(b, c.Component.Name)
		require.False(b, seen[c.Component.Type], "duplicate component type")
		seen[c.Component.Type] = true
		switch c.Component.Type {
		case app.ComponentTypeHelmChart:
			require.NotNil(b, c.HelmComponentConfig)
			require.NotNil(b, c.HelmComponentConfig.PublicGitVCSConfig)
			require.Equal(b, "my-chart", c.HelmComponentConfig.ChartName)
		case app.ComponentTypeTerraformModule:
			require.NotNil(b, c.TerraformModuleComponentConfig)
			require.NotNil(b, c.TerraformModuleComponentConfig.PublicGitVCSConfig)
		case app.ComponentTypeDockerBuild:
			require.NotNil(b, c.DockerBuildComponentConfig)
			require.NotNil(b, c.DockerBuildComponentConfig.PublicGitVCSConfig)
		case app.ComponentTypeKubernetesManifest:
			require.NotNil(b, c.KubernetesManifestComponentConfig)
			require.Contains(b, c.KubernetesManifestComponentConfig.Manifest, "ConfigMap")
		case app.ComponentTypeExternalImage:
			require.NotNil(b, c.ExternalImageComponentConfig)
			require.Equal(b, "nginx", c.ExternalImageComponentConfig.ImageURL)
		case app.ComponentTypeJob:
			require.NotNil(b, c.JobComponentConfig)
			require.Equal(b, "ubuntu", c.JobComponentConfig.ImageURL)
		case app.ComponentTypePulumi:
			require.NotNil(b, c.PulumiComponentConfig)
			require.NotNil(b, c.PulumiComponentConfig.PublicGitVCSConfig)
			require.Equal(b, "go", c.PulumiComponentConfig.Runtime)
		default:
			b.Fatalf("unexpected component type %q", c.Component.Type)
		}
	}
	require.Len(b, got.ActionWorkflowConfigs, actions)
	counts := make(map[app.ActionWorkflowTriggerType]int)
	for _, action := range got.ActionWorkflowConfigs {
		require.Equal(b, configID, action.AppConfigID)
		require.NotEmpty(b, action.Triggers)
		manual := 0
		cron, lifecycle := 0, 0
		for _, trigger := range action.Triggers {
			require.Equal(b, action.ID, trigger.ActionWorkflowConfigID)
			require.Equal(b, configID, trigger.AppConfigID)
			counts[trigger.Type]++
			switch trigger.Type {
			case app.ActionWorkflowTriggerTypeManual:
				manual++
			case app.ActionWorkflowTriggerTypeCron:
				cron++
				require.Equal(b, "17 */2 * * *", trigger.CronSchedule)
				require.NotNil(b, action.CronTrigger)
				require.Equal(b, trigger.ID, action.CronTrigger.ID)
			case app.ActionWorkflowTriggerTypePostDeployComponent:
				lifecycle++
				require.Equal(b, componentID, trigger.ComponentID.ValueString())
				require.Len(b, action.LifecycleTriggers, 1)
				require.Equal(b, trigger.ID, action.LifecycleTriggers[0].ID)
			default:
				b.Fatalf("unexpected trigger type %q", trigger.Type)
			}
		}
		require.Equal(b, 1, manual)
		require.Len(b, action.LifecycleTriggers, lifecycle)
		if cron == 0 {
			require.Nil(b, action.CronTrigger)
		}
	}
	require.Equal(b, map[app.ActionWorkflowTriggerType]int{
		app.ActionWorkflowTriggerTypeManual:              actions,
		app.ActionWorkflowTriggerTypeCron:                (actions + 9) / 10,
		app.ActionWorkflowTriggerTypePostDeployComponent: (actions + 5) / 6,
	}, counts)
}
