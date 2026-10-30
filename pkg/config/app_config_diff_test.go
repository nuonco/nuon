package config

import (
	"encoding/json"
	"maps"
	"slices"
	"strconv"
	"testing"

	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/config/refs"
)

type AppConfigDiffSuite struct {
	suite.Suite
}

func TestAppConfigDiffSuite(t *testing.T) {
	suite.Run(t, new(AppConfigDiffSuite))
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func (s *AppConfigDiffSuite) TestRemovingPreviewClearsInstallTarget() {
	got := diffBranchPreview(
		&AppBranchPreviewConfig{Mode: "plan-only", InstallName: "example"},
		nil,
	)

	s.NotContains(got.FormatChanged(""), "mode:")
	s.Contains(got.FormatChanged(""), "install_name: 'example' -> ''")
}

// baseConfig returns a minimal valid AppConfig for use in tests.
func baseConfig() *AppConfig {
	return &AppConfig{
		Version:     "1",
		Description: "Test app",
		DisplayName: "Test",
		Sandbox: &AppSandboxConfig{
			TerraformVersion: "1.5.0",
			PublicRepo: &PublicRepoConfig{
				Repo:      "https://github.com/test/sandbox",
				Directory: "terraform",
				Branch:    "main",
			},
		},
		Runner: &AppRunnerConfig{
			RunnerType: "aws",
		},
	}
}

// --- Test: Identical configs ---

func (s *AppConfigDiffSuite) TestIdenticalConfigs() {
	a := baseConfig()
	b := baseConfig()
	d := a.Diff(b)
	s.Require().NotNil(d)
	summary := d.Summary()
	s.False(summary.HasChanged)
	s.Equal(0, summary.Added)
	s.Equal(0, summary.Removed)
	s.Equal(0, summary.Changed)
}

func (s *AppConfigDiffSuite) TestTriggerOnlyChanges() {
	old := baseConfig()
	new := baseConfig()
	new.Triggers = &TriggersConfig{Rules: []*TriggerRuleConfig{{
		Name: "deploy", Trigger: "registry", EventTypes: []string{"push"},
		Target: &TriggerTargetConfig{Type: "app_branch_run", AppBranch: "main"},
	}}}

	d := new.Diff(old)
	s.True(d.Summary().HasChanged)
	triggers := findChild(d, "triggers")
	s.Require().NotNil(triggers)
	s.Equal(diff.OpAdd, triggers.Diff.Op)
}

// --- Test: Metadata changes ---

func (s *AppConfigDiffSuite) TestMetadataChanges() {
	old := baseConfig()
	new := baseConfig()
	new.Version = "2"
	new.Description = "Updated app"
	new.DisplayName = "Updated"

	d := new.Diff(old)
	summary := d.Summary()
	s.True(summary.HasChanged)

	// Find version diff
	found := findChild(d, "version")
	s.Require().NotNil(found)
	s.Equal(diff.OpChange, found.Diff.Op)
	s.Contains(found.Diff.Diff, "'1' -> '2'")

	// Find description diff
	found = findChild(d, "description")
	s.Require().NotNil(found)
	s.Equal(diff.OpChange, found.Diff.Op)
}

// --- Test: Sandbox changes ---

func (s *AppConfigDiffSuite) TestSandboxChanges() {
	old := baseConfig()
	new := baseConfig()
	new.Sandbox.TerraformVersion = "1.6.0"
	new.Sandbox.EnvVarMap = map[string]string{"NEW_VAR": "value"}
	new.Sandbox.PublicRepo.Branch = "develop"

	d := new.Diff(old)
	sandbox := findChild(d, "sandbox")
	s.Require().NotNil(sandbox)

	tfVer := findChild(sandbox, "terraform_version")
	s.Require().NotNil(tfVer)
	s.Equal(diff.OpChange, tfVer.Diff.Op)
	s.Contains(tfVer.Diff.Diff, "'1.5.0' -> '1.6.0'")

	envVars := findChild(sandbox, "env_vars")
	s.Require().NotNil(envVars)
	newVar := findChild(envVars, "NEW_VAR")
	s.Require().NotNil(newVar)
	s.Equal(diff.OpAdd, newVar.Diff.Op)

	pubRepo := findChild(sandbox, "public_repo")
	s.Require().NotNil(pubRepo)
	branch := findChild(pubRepo, "branch")
	s.Require().NotNil(branch)
	s.Equal(diff.OpChange, branch.Diff.Op)
}

// --- Test: Runner changes ---

func (s *AppConfigDiffSuite) TestRunnerChanges() {
	old := baseConfig()
	new := baseConfig()
	new.Runner.RunnerType = "gcp"
	new.Runner.HelmDriver = "secret"
	new.Runner.EnvVarMap = map[string]string{"KEY": "val"}

	d := new.Diff(old)
	runner := findChild(d, "runner")
	s.Require().NotNil(runner)

	rt := findChild(runner, "runner_type")
	s.Require().NotNil(rt)
	s.Equal(diff.OpChange, rt.Diff.Op)

	hd := findChild(runner, "helm_driver")
	s.Require().NotNil(hd)
	s.Equal(diff.OpAdd, hd.Diff.Op)

	ev := findChild(runner, "env_vars")
	s.Require().NotNil(ev)
}

// --- Test: Component addition ---

func (s *AppConfigDiffSuite) TestComponentAddition() {
	old := baseConfig()
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "database",
			Type: TerraformModuleComponentType,
			TerraformModule: &TerraformModuleComponentConfig{
				TerraformVersion: "1.5.0",
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	s.Require().NotNil(comps)

	db := findChild(comps, "component.database")
	s.Require().NotNil(db)
	// Added component now shows full details as children
	s.Require().NotNil(db.Children)
	typeChild := findChild(db, "type")
	s.Require().NotNil(typeChild)
	s.Equal(diff.OpAdd, typeChild.Diff.Op)
}

// --- Test: Component removal ---

func (s *AppConfigDiffSuite) TestComponentRemoval() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "legacy",
			Type: HelmChartComponentType,
			HelmChart: &HelmChartComponentConfig{
				ChartName: "legacy-chart",
			},
		},
	}
	new := baseConfig()

	d := new.Diff(old)
	comps := findChild(d, "components")
	s.Require().NotNil(comps)

	legacy := findChild(comps, "component.legacy")
	s.Require().NotNil(legacy)
	// Removed component shows full details as children
	s.Require().NotNil(legacy.Children)
	typeChild := findChild(legacy, "type")
	s.Require().NotNil(typeChild)
	s.Equal(diff.OpRemove, typeChild.Diff.Op)
}

// --- Test: Component field changes ---

func (s *AppConfigDiffSuite) TestComponentChanges() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "infra",
			Type: TerraformModuleComponentType,
			TerraformModule: &TerraformModuleComponentConfig{
				TerraformVersion: "1.5.0",
				VarsMap:          map[string]string{"region": "us-east-1"},
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "infra",
			Type: TerraformModuleComponentType,
			TerraformModule: &TerraformModuleComponentConfig{
				TerraformVersion: "1.6.0",
				VarsMap:          map[string]string{"region": "us-west-2"},
				EnvVarMap:        map[string]string{"TF_LOG": "debug"},
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	s.Require().NotNil(comps)

	infra := findChild(comps, "component.infra")
	s.Require().NotNil(infra)

	tfVer := findChild(infra, "terraform_version")
	s.Require().NotNil(tfVer)
	s.Equal(diff.OpChange, tfVer.Diff.Op)

	vars := findChild(infra, "vars")
	s.Require().NotNil(vars)
	region := findChild(vars, "region")
	s.Require().NotNil(region)
	s.Equal(diff.OpChange, region.Diff.Op)

	envVars := findChild(infra, "env_vars")
	s.Require().NotNil(envVars)
}

func (s *AppConfigDiffSuite) TestTerraformTeardownDiff() {
	component := func(retain ...string) ComponentList {
		cfg := &TerraformModuleComponentConfig{TerraformVersion: "1.11.3"}
		if len(retain) > 0 {
			cfg.Teardown = &ComponentTeardownConfig{Retain: retain}
		}
		return ComponentList{{Name: "storage", Type: TerraformModuleComponentType, TerraformModule: cfg}}
	}

	old := baseConfig()
	old.Components = component("aws_rds_cluster.main")
	new := baseConfig()
	new.Components = component("aws_rds_cluster.main", "module.buckets")

	storage := findChild(findChild(new.Diff(old), "components"), "component.storage")
	s.Require().NotNil(storage)
	teardown := findChild(storage, "teardown")
	s.Require().NotNil(teardown)
	retain := findChild(teardown, "retain")
	s.Require().NotNil(retain)
	s.Equal(diff.OpChange, retain.Diff.Op)

	unchanged := baseConfig()
	unchanged.Components = component("aws_rds_cluster.main")
	s.False(unchanged.Diff(old).Summary().HasChanged)
}

// --- Test: Helm chart diff ---

func (s *AppConfigDiffSuite) TestHelmChartDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "chart",
			Type: HelmChartComponentType,
			HelmChart: &HelmChartComponentConfig{
				ChartName: "prometheus",
				ValuesMap: map[string]string{"replicas": "1"},
				Namespace: "monitoring",
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "chart",
			Type: HelmChartComponentType,
			HelmChart: &HelmChartComponentConfig{
				ChartName: "prometheus",
				ValuesMap: map[string]string{"replicas": "3", "log_level": "info"},
				Namespace: "observability",
				SkipCRDs:  true,
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	chart := findChild(comps, "component.chart")
	s.Require().NotNil(chart)

	ns := findChild(chart, "namespace")
	s.Require().NotNil(ns)
	s.Equal(diff.OpChange, ns.Diff.Op)
	skipCRDs := findChild(chart, "skip_crds")
	s.Require().NotNil(skipCRDs)
	s.Equal(diff.OpChange, skipCRDs.Diff.Op)

	vals := findChild(chart, "values")
	s.Require().NotNil(vals)
	replicas := findChild(vals, "replicas")
	s.Equal(diff.OpChange, replicas.Diff.Op)

	logLevel := findChild(vals, "log_level")
	s.Equal(diff.OpAdd, logLevel.Diff.Op)
}

// --- Test: ungrouped section TOML content diff ---

func (s *AppConfigDiffSuite) TestRunnerSectionTOMLDiff() {
	old := baseConfig()
	new := baseConfig()
	new.Runner.RunnerType = "gpu"

	d := new.Diff(old)
	runner := findChild(d, "runner")
	s.Require().NotNil(runner)
	s.Require().NotNil(runner.Diff)
	s.Equal(diff.OpChange, runner.Diff.Op)
	s.Contains(runner.Diff.Before, "aws")
	s.Contains(runner.Diff.After, "gpu")
	// field children are retained for change counting
	s.Require().NotNil(findChild(runner, "runner_type"))
}

func (s *AppConfigDiffSuite) TestInputsSectionTOMLDiff() {
	old := baseConfig()
	new := baseConfig()
	new.Inputs = &AppInputConfig{
		Inputs: []AppInput{
			{Name: "cluster_name", Type: "string", Required: true},
		},
	}

	d := new.Diff(old)
	inputs := findChild(d, "inputs")
	s.Require().NotNil(inputs)
	s.Require().NotNil(inputs.Diff)
	s.Equal(diff.OpAdd, inputs.Diff.Op)
	s.Contains(inputs.Diff.After, "cluster_name")
}

func (s *AppConfigDiffSuite) TestActionInlineScriptContentDiff() {
	old := baseConfig()
	old.Actions = []*ActionConfig{
		{
			Name:  "migrate",
			Steps: []*ActionStepConfig{{Name: "run", InlineContents: "#!/bin/bash\necho v1\n"}},
		},
	}
	new := baseConfig()
	new.Actions = []*ActionConfig{
		{
			Name:  "migrate",
			Steps: []*ActionStepConfig{{Name: "run", InlineContents: "#!/bin/bash\necho v2\n"}},
		},
	}

	d := new.Diff(old)
	step := findChild(findChild(findChild(d, "actions"), "action.migrate"), "step.run")
	s.Require().NotNil(step)
	script := findChild(step, "inline_contents")
	s.Require().NotNil(script)
	s.Equal(diff.OpChange, script.Diff.Op)
	s.Contains(script.Diff.Before, "echo v1")
	s.Contains(script.Diff.After, "echo v2")
}

// --- Test: Helm values file content diff ---

func (s *AppConfigDiffSuite) TestHelmValuesFilesContentDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "chart",
			Type: HelmChartComponentType,
			HelmChart: &HelmChartComponentConfig{
				ChartName: "prometheus",
				ValuesFiles: []HelmValuesFile{
					{Path: "./values/prod.yaml", Contents: "replicas: 1\n"},
					{Path: "./values/legacy.yaml", Contents: "enabled: true\n"},
				},
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "chart",
			Type: HelmChartComponentType,
			HelmChart: &HelmChartComponentConfig{
				ChartName: "prometheus",
				ValuesFiles: []HelmValuesFile{
					{Path: "./values/prod.yaml", Contents: "replicas: 3\n"},
					{Path: "./values/new.yaml", Contents: "log_level: info\n"},
				},
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	chart := findChild(comps, "component.chart")
	s.Require().NotNil(chart)

	files := findChild(chart, "values_files")
	s.Require().NotNil(files)

	changed := findChild(files, "./values/prod.yaml")
	s.Require().NotNil(changed)
	s.Equal(diff.OpChange, changed.Diff.Op)
	s.Equal("modified", changed.Diff.Diff)
	s.Equal("replicas: 1\n", changed.Diff.Before)
	s.Equal("replicas: 3\n", changed.Diff.After)

	added := findChild(files, "./values/new.yaml")
	s.Require().NotNil(added)
	s.Equal(diff.OpAdd, added.Diff.Op)
	s.Equal("", added.Diff.Before)
	s.Equal("log_level: info\n", added.Diff.After)

	removed := findChild(files, "./values/legacy.yaml")
	s.Require().NotNil(removed)
	s.Equal(diff.OpRemove, removed.Diff.Op)
	s.Equal("enabled: true\n", removed.Diff.Before)
	s.Equal("", removed.Diff.After)
}

// --- Test: Terraform variables file content diff ---

func (s *AppConfigDiffSuite) TestTerraformVariablesFilesContentDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "vpc",
			Type: TerraformModuleComponentType,
			TerraformModule: &TerraformModuleComponentConfig{
				TerraformVersion: "1.5.0",
				VariablesFiles: []TerraformVariablesFile{
					{Contents: "region = \"us-east-1\"\n"},
				},
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "vpc",
			Type: TerraformModuleComponentType,
			TerraformModule: &TerraformModuleComponentConfig{
				TerraformVersion: "1.5.0",
				VariablesFiles: []TerraformVariablesFile{
					{Contents: "region = \"us-west-2\"\n"},
				},
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	vpc := findChild(comps, "component.vpc")
	s.Require().NotNil(vpc)

	files := findChild(vpc, "var_file")
	s.Require().NotNil(files)

	changed := findChild(files, "var_file[0]")
	s.Require().NotNil(changed)
	s.Equal(diff.OpChange, changed.Diff.Op)
	s.Equal("region = \"us-east-1\"\n", changed.Diff.Before)
	s.Equal("region = \"us-west-2\"\n", changed.Diff.After)
}

// --- Test: Kubernetes manifest content diff ---

func (s *AppConfigDiffSuite) TestKubernetesManifestContentDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "manifests",
			Type: KubernetesManifestComponentType,
			KubernetesManifest: &KubernetesManifestComponentConfig{
				Namespace: "default",
				Manifest:  "kind: Service\nspec:\n  replicas: 1\n",
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "manifests",
			Type: KubernetesManifestComponentType,
			KubernetesManifest: &KubernetesManifestComponentConfig{
				Namespace: "default",
				Manifest:  "kind: Service\nspec:\n  replicas: 3\n",
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	manifests := findChild(comps, "component.manifests")
	s.Require().NotNil(manifests)

	manifest := findChild(manifests, "manifest")
	s.Require().NotNil(manifest)
	s.Equal(diff.OpChange, manifest.Diff.Op)
	s.Equal("modified", manifest.Diff.Diff)
	s.Equal("kind: Service\nspec:\n  replicas: 1\n", manifest.Diff.Before)
	s.Equal("kind: Service\nspec:\n  replicas: 3\n", manifest.Diff.After)
}

// --- Test: Terraform module diff ---

func (s *AppConfigDiffSuite) TestTerraformModuleDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "vpc",
			Type: TerraformModuleComponentType,
			TerraformModule: &TerraformModuleComponentConfig{
				TerraformVersion: "1.5.0",
				VarsMap:          map[string]string{"cidr": "10.0.0.0/16"},
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "vpc",
			Type: TerraformModuleComponentType,
			TerraformModule: &TerraformModuleComponentConfig{
				TerraformVersion: "1.5.0",
				VarsMap:          map[string]string{"cidr": "10.1.0.0/16"},
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	vpc := findChild(comps, "component.vpc")
	s.Require().NotNil(vpc)

	vars := findChild(vpc, "vars")
	cidr := findChild(vars, "cidr")
	s.Require().NotNil(cidr)
	s.Equal(diff.OpChange, cidr.Diff.Op)

	tfVer := findChild(vpc, "terraform_version")
	s.Equal(diff.OpNoop, tfVer.Diff.Op)
}

// --- Test: Docker build diff ---

func (s *AppConfigDiffSuite) TestDockerBuildDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "api",
			Type: DockerBuildComponentType,
			DockerBuild: &DockerBuildComponentConfig{
				Dockerfile: "Dockerfile",
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "api",
			Type: DockerBuildComponentType,
			DockerBuild: &DockerBuildComponentConfig{
				Dockerfile: "Dockerfile.prod",
				EnvVarMap:  map[string]string{"GO_ENV": "production"},
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	api := findChild(comps, "component.api")
	s.Require().NotNil(api)

	df := findChild(api, "dockerfile")
	s.Require().NotNil(df)
	s.Equal(diff.OpChange, df.Diff.Op)

	ev := findChild(api, "env_vars")
	s.Require().NotNil(ev)
}

// --- Test: External image diff ---

func (s *AppConfigDiffSuite) TestExternalImageDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "worker",
			Type: ExternalImageComponentType,
			ExternalImage: &ExternalImageComponentConfig{
				PublicImageConfig: &PublicImageConfig{
					ImageURL: "nginx",
					Tag:      "1.24",
				},
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "worker",
			Type: ExternalImageComponentType,
			ExternalImage: &ExternalImageComponentConfig{
				PublicImageConfig: &PublicImageConfig{
					ImageURL: "nginx",
					Tag:      "1.25",
				},
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	worker := findChild(comps, "component.worker")
	s.Require().NotNil(worker)

	pub := findChild(worker, "public_image")
	s.Require().NotNil(pub)

	tag := findChild(pub, "tag")
	s.Require().NotNil(tag)
	s.Equal(diff.OpChange, tag.Diff.Op)
	s.Contains(tag.Diff.Diff, "'1.24' -> '1.25'")
}

// --- Test: Kubernetes manifest diff ---

func (s *AppConfigDiffSuite) TestKubernetesManifestDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "deploy",
			Type: KubernetesManifestComponentType,
			KubernetesManifest: &KubernetesManifestComponentConfig{
				Manifest:  "apiVersion: v1\nkind: ConfigMap",
				Namespace: "default",
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "deploy",
			Type: KubernetesManifestComponentType,
			KubernetesManifest: &KubernetesManifestComponentConfig{
				Manifest:  "apiVersion: v1\nkind: Deployment",
				Namespace: "production",
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	deploy := findChild(comps, "component.deploy")
	s.Require().NotNil(deploy)

	manifest := findChild(deploy, "manifest")
	s.Equal(diff.OpChange, manifest.Diff.Op)

	ns := findChild(deploy, "namespace")
	s.Equal(diff.OpChange, ns.Diff.Op)
}

// --- Test: Job diff ---

func (s *AppConfigDiffSuite) TestJobDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "migrate",
			Type: JobComponentType,
			Job: &JobComponentConfig{
				ImageURL: "python",
				Tag:      "3.11",
				Cmd:      []string{"python", "migrate.py"},
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "migrate",
			Type: JobComponentType,
			Job: &JobComponentConfig{
				ImageURL: "python",
				Tag:      "3.12",
				Cmd:      []string{"python", "migrate.py"},
				Args:     []string{"--verbose"},
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	migrate := findChild(comps, "component.migrate")
	s.Require().NotNil(migrate)

	tag := findChild(migrate, "tag")
	s.Equal(diff.OpChange, tag.Diff.Op)

	args := findChild(migrate, "args")
	s.Equal(diff.OpAdd, args.Diff.Op)
}

// --- Test: Pulumi diff ---

func (s *AppConfigDiffSuite) TestPulumiDiff() {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "pulumi-infra",
			Type: PulumiComponentType,
			Pulumi: &PulumiComponentConfig{
				Runtime:   "go",
				ConfigMap: map[string]string{"aws:region": "us-east-1"},
			},
		},
	}
	new := baseConfig()
	new.Components = ComponentList{
		{
			Name: "pulumi-infra",
			Type: PulumiComponentType,
			Pulumi: &PulumiComponentConfig{
				Runtime:   "nodejs",
				ConfigMap: map[string]string{"aws:region": "us-west-2"},
			},
		},
	}

	d := new.Diff(old)
	comps := findChild(d, "components")
	pi := findChild(comps, "component.pulumi-infra")
	s.Require().NotNil(pi)

	rt := findChild(pi, "runtime")
	s.Equal(diff.OpChange, rt.Diff.Op)

	cfg := findChild(pi, "config")
	region := findChild(cfg, "aws:region")
	s.Equal(diff.OpChange, region.Diff.Op)
}

// --- Test: Input changes ---

func (s *AppConfigDiffSuite) TestInputChanges() {
	old := baseConfig()
	old.Inputs = &AppInputConfig{
		Groups: []AppInputGroup{
			{Name: "database", Description: "DB config"},
		},
		Inputs: []AppInput{
			{Name: "db_host", DisplayName: "DB Host", Group: "database", Required: true},
			{Name: "db_port", DisplayName: "DB Port", Group: "database"},
		},
	}
	new := baseConfig()
	new.Inputs = &AppInputConfig{
		Groups: []AppInputGroup{
			{Name: "database", Description: "Database configuration"},
			{Name: "cache", Description: "Cache config"},
		},
		Inputs: []AppInput{
			{Name: "db_host", DisplayName: "Database Host", Group: "database", Required: true},
			{Name: "redis_url", DisplayName: "Redis URL", Group: "cache"},
		},
	}

	d := new.Diff(old)
	inputs := findChild(d, "inputs")
	s.Require().NotNil(inputs)

	// Group changed
	dbGroup := findChild(inputs, "group.database")
	s.Require().NotNil(dbGroup)

	// Group added — now expanded with children
	cacheGroup := findChild(inputs, "group.cache")
	s.Require().NotNil(cacheGroup)
	s.Require().NotNil(cacheGroup.Children)
	cacheDesc := findChild(cacheGroup, "description")
	s.Require().NotNil(cacheDesc)
	s.Equal(diff.OpAdd, cacheDesc.Diff.Op)

	// Input changed
	dbHost := findChild(inputs, "input.db_host")
	s.Require().NotNil(dbHost)
	dn := findChild(dbHost, "display_name")
	s.Equal(diff.OpChange, dn.Diff.Op)

	// Input removed — now expanded with children
	dbPort := findChild(inputs, "input.db_port")
	s.Require().NotNil(dbPort)
	s.Require().NotNil(dbPort.Children)
	dbPortDN := findChild(dbPort, "display_name")
	s.Require().NotNil(dbPortDN)
	s.Equal(diff.OpRemove, dbPortDN.Diff.Op)

	// Input added — now expanded with children
	redis := findChild(inputs, "input.redis_url")
	s.Require().NotNil(redis)
	s.Require().NotNil(redis.Children)
	redisDN := findChild(redis, "display_name")
	s.Require().NotNil(redisDN)
	s.Equal(diff.OpAdd, redisDN.Diff.Op)
}

// --- Test: Install changes ---

func (s *AppConfigDiffSuite) TestInstallDiff() {
	old := baseConfig()
	old.Installs = []*Install{
		{
			Name: "prod",
			AWSAccount: &AWSAccount{
				Region: "us-east-1",
			},
		},
		{
			Name: "staging",
			AWSAccount: &AWSAccount{
				Region: "us-west-2",
			},
		},
	}
	new := baseConfig()
	new.Installs = []*Install{
		{
			Name: "prod",
			AWSAccount: &AWSAccount{
				Region: "eu-west-1", // changed
			},
		},
		{
			Name: "dev", // added
			AWSAccount: &AWSAccount{
				Region: "us-east-1",
			},
		},
	}

	d := new.Diff(old)
	installs := findChild(d, "installs")
	s.Require().NotNil(installs)

	// Prod changed (delegated to Install.Diff)
	prod := findChild(installs, "prod")
	s.Require().NotNil(prod)
	summary := prod.Summary()
	s.True(summary.HasChanged)

	// Staging removed — uses Install.Diff to show details
	staging := findChild(installs, "staging")
	s.Require().NotNil(staging)
	s.Require().NotNil(staging.Children)

	// Dev added
	dev := findChild(installs, "dev")
	s.Require().NotNil(dev)
}

// --- Test: Permission changes ---

func (s *AppConfigDiffSuite) TestPermissionChanges() {
	old := baseConfig()
	old.Permissions = &PermissionsConfig{
		ProvisionRole: &AppAWSIAMRole{
			Name:        "provision",
			Description: "Provision role",
		},
	}
	new := baseConfig()
	new.Permissions = &PermissionsConfig{
		ProvisionRole: &AppAWSIAMRole{
			Name:        "provision",
			Description: "Updated provision role",
		},
		CustomRoles: []*AppAWSIAMRole{
			{Name: "custom-ops", Description: "Custom operations"},
		},
	}

	d := new.Diff(old)
	perms := findChild(d, "permissions")
	s.Require().NotNil(perms)

	provRole := findChild(perms, "provision_role")
	s.Require().NotNil(provRole)
	desc := findChild(provRole, "description")
	s.Equal(diff.OpChange, desc.Diff.Op)

	customOps := findChild(perms, "custom_role.custom-ops")
	s.Require().NotNil(customOps)
	// Added role now expanded with children
	s.Require().NotNil(customOps.Children)
	customDesc := findChild(customOps, "description")
	s.Require().NotNil(customDesc)
	s.Equal(diff.OpAdd, customDesc.Diff.Op)
}

// --- Test: Secret changes ---

func (s *AppConfigDiffSuite) TestSecretChanges() {
	old := baseConfig()
	old.Secrets = &SecretsConfig{
		Secrets: []*AppSecret{
			{Name: "api_key", Description: "API key", Required: true},
			{Name: "db_pass", Description: "DB password", AutoGenerate: true},
		},
	}
	new := baseConfig()
	new.Secrets = &SecretsConfig{
		Secrets: []*AppSecret{
			{Name: "api_key", Description: "API key", Required: false, Format: "base64"},
			{Name: "new_secret", Description: "New one"},
		},
	}

	d := new.Diff(old)
	secrets := findChild(d, "secrets")
	s.Require().NotNil(secrets)

	apiKey := findChild(secrets, "secret.api_key")
	s.Require().NotNil(apiKey)
	req := findChild(apiKey, "required")
	s.Equal(diff.OpChange, req.Diff.Op)
	format := findChild(apiKey, "format")
	s.Equal(diff.OpAdd, format.Diff.Op)

	dbPass := findChild(secrets, "secret.db_pass")
	s.Require().NotNil(dbPass)
	// Removed secret now expanded with children
	s.Require().NotNil(dbPass.Children)
	dbPassDesc := findChild(dbPass, "description")
	s.Require().NotNil(dbPassDesc)
	s.Equal(diff.OpRemove, dbPassDesc.Diff.Op)

	newSecret := findChild(secrets, "secret.new_secret")
	s.Require().NotNil(newSecret)
	// Added secret now expanded with children
	s.Require().NotNil(newSecret.Children)
	newSecretDesc := findChild(newSecret, "description")
	s.Require().NotNil(newSecretDesc)
	s.Equal(diff.OpAdd, newSecretDesc.Diff.Op)
}

// --- Test: Policy changes ---

func (s *AppConfigDiffSuite) TestPolicyChanges() {
	old := baseConfig()
	old.Policies = &PoliciesConfig{
		Policies: []AppPolicy{
			{Name: "block-tags", Type: AppPolicyTypeKubernetesCluster, Engine: AppPolicyEngineKyverno, Contents: "old-content"},
		},
	}
	new := baseConfig()
	new.Policies = &PoliciesConfig{
		Policies: []AppPolicy{
			{Name: "block-tags", Type: AppPolicyTypeKubernetesCluster, Engine: AppPolicyEngineOPA, Contents: "new-content"},
		},
	}

	d := new.Diff(old)
	policies := findChild(d, "policies")
	s.Require().NotNil(policies)

	bt := findChild(policies, "policy.block-tags")
	s.Require().NotNil(bt)

	engine := findChild(bt, "engine")
	s.Equal(diff.OpChange, engine.Diff.Op)

	contents := findChild(bt, "contents")
	s.Equal(diff.OpChange, contents.Diff.Op)
}

func (s *AppConfigDiffSuite) TestPolicyApplyOnTeardownChange() {
	no := false
	old := baseConfig()
	old.Policies = &PoliciesConfig{
		Policies: []AppPolicy{
			{Name: "protect-rds", Type: AppPolicyTypeTerraformModule, Engine: AppPolicyEngineOPA, Contents: "c", Components: []string{"rds"}},
		},
	}
	new := baseConfig()
	new.Policies = &PoliciesConfig{
		Policies: []AppPolicy{
			{Name: "protect-rds", Type: AppPolicyTypeTerraformModule, Engine: AppPolicyEngineOPA, Contents: "c", Components: []string{"rds"}, ApplyOnTeardown: &no},
		},
	}

	d := new.Diff(old)
	policies := findChild(d, "policies")
	s.Require().NotNil(policies)

	pr := findChild(policies, "policy.protect-rds")
	s.Require().NotNil(pr)

	teardown := findChild(pr, "apply_on_teardown")
	s.Require().NotNil(teardown)
	s.NotEqual(diff.OpNoop, teardown.Diff.Op)
}

// --- Test: Action changes ---

func (s *AppConfigDiffSuite) TestActionChanges() {
	old := baseConfig()
	old.Actions = []*ActionConfig{
		{
			Name:    "healthcheck",
			Timeout: "30s",
			Steps: []*ActionStepConfig{
				{Name: "check", Command: "curl http://localhost"},
			},
			Triggers: []*ActionTriggerConfig{
				{Type: "manual"},
			},
		},
	}
	new := baseConfig()
	new.Actions = []*ActionConfig{
		{
			Name:    "healthcheck",
			Timeout: "60s",
			Steps: []*ActionStepConfig{
				{Name: "check", Command: "curl http://localhost/health"},
				{Name: "notify", Command: "echo done"},
			},
			Triggers: []*ActionTriggerConfig{
				{Type: "cron", CronSchedule: "*/5 * * * *"},
			},
		},
	}

	d := new.Diff(old)
	actions := findChild(d, "actions")
	s.Require().NotNil(actions)

	hc := findChild(actions, "action.healthcheck")
	s.Require().NotNil(hc)

	timeout := findChild(hc, "timeout")
	s.Equal(diff.OpChange, timeout.Diff.Op)

	check := findChild(hc, "step.check")
	s.Require().NotNil(check)
	cmd := findChild(check, "command")
	s.Equal(diff.OpChange, cmd.Diff.Op)

	notify := findChild(hc, "step.notify")
	s.Require().NotNil(notify)
	// Added step now expanded with children
	s.Require().NotNil(notify.Children)
	notifyCmd := findChild(notify, "command")
	s.Require().NotNil(notifyCmd)
	s.Equal(diff.OpAdd, notifyCmd.Diff.Op)

	trigger := findChild(hc, "trigger.0")
	s.Require().NotNil(trigger)
	trigType := findChild(trigger, "type")
	s.Equal(diff.OpChange, trigType.Diff.Op)
}

func TestAppConfigDiff_ActionDependencyProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		names := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 0, 12)
		oldDependencies := drawPermutation(t, names)
		newDependencies := append([]string{}, drawPermutation(t, names)...)
		script := hegel.Draw(t, hegel.Text())
		makeConfig := func(dependencies []string) *AppConfig {
			cfg := baseConfig()
			cfg.Actions = []*ActionConfig{{
				Name:         "inspect",
				Dependencies: slices.Clone(dependencies),
				Steps: []*ActionStepConfig{
					{Name: "inspect", InlineContents: script},
					{Name: "notify", Command: "echo done"},
				},
			}}
			return cfg
		}
		old, new := makeConfig(oldDependencies), makeConfig(newDependencies)
		oldSnapshot, err := json.Marshal(old)
		require.NoError(t, err)
		newSnapshot, err := json.Marshal(new)
		require.NoError(t, err)

		for range 2 {
			require.False(t, new.Diff(old).Summary().HasChanged, "dependency permutations must not change actions")
			require.Empty(t, new.Diff(old).FormatChanged(""))
			require.False(t, old.Diff(new).Summary().HasChanged)
		}
		oldAfter, err := json.Marshal(old)
		require.NoError(t, err)
		newAfter, err := json.Marshal(new)
		require.NoError(t, err)
		require.Equal(t, oldSnapshot, oldAfter, "comparison must not mutate the old config")
		require.Equal(t, newSnapshot, newAfter, "comparison must not mutate the new config")

		for _, edit := range []struct {
			name   string
			mutate func(*ActionConfig)
		}{
			{"dependency membership", func(a *ActionConfig) { a.Dependencies = append(a.Dependencies, "extra_dependency") }},
			{"script contents", func(a *ActionConfig) { a.Steps[0].InlineContents += "\necho changed\n" }},
			{"step order", func(a *ActionConfig) { slices.Reverse(a.Steps) }},
		} {
			changed := makeConfig(newDependencies)
			edit.mutate(changed.Actions[0])
			require.True(t, changed.Diff(old).Summary().HasChanged, "%s must remain visible", edit.name)
			require.True(t, old.Diff(changed).Summary().HasChanged, "%s must remain visible in reverse", edit.name)
		}
		if len(newDependencies) > 0 {
			changed := makeConfig(newDependencies)
			changed.Actions[0].Dependencies[0] = "replacement_dependency"
			require.True(t, changed.Diff(old).Summary().HasChanged, "same-length dependency changes must remain visible")
		}
	})
}

func TestAppConfigDiff_IdentityAndCollectionProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		old := generatedDiffConfig(t)
		before, err := json.Marshal(old)
		require.NoError(t, err)
		var new AppConfig
		require.NoError(t, json.Unmarshal(before, &new))
		require.False(t, old.Diff(old).Summary().HasChanged)
		require.False(t, new.Diff(old).Summary().HasChanged)

		new.Components = drawPermutation(t, new.Components)
		new.Actions = drawPermutation(t, new.Actions)
		new.Runbooks = drawPermutation(t, new.Runbooks)
		newBefore, err := json.Marshal(&new)
		require.NoError(t, err)
		forward := new.Diff(old)
		require.False(t, forward.Summary().HasChanged)
		require.Empty(t, forward.FormatChanged(""))
		require.False(t, old.Diff(&new).Summary().HasChanged)
		require.Equal(t, forward, new.Diff(old), "repeated comparison must be deterministic")
		after, err := json.Marshal(old)
		require.NoError(t, err)
		require.Equal(t, before, after)
		newAfter, err := json.Marshal(&new)
		require.NoError(t, err)
		require.Equal(t, newBefore, newAfter)
	})
}

func TestAppConfigDiff_MapOrderProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		values := hegel.Draw(t, hegel.Maps(hegel.FromRegex(`[a-z]{1,8}`, true), hegel.Text()).MaxSize(10))
		keys := slices.Sorted(maps.Keys(values))
		makeConfig := func(order []string) *AppConfig {
			m := make(map[string]string, len(values))
			for _, key := range order {
				m[key] = values[key]
			}
			cfg := baseConfig()
			cfg.Runner.EnvVarMap = maps.Clone(m)
			cfg.Sandbox.VarsMap = maps.Clone(m)
			cfg.Components = ComponentList{
				{Name: "terraform", Type: TerraformModuleComponentType, Labels: maps.Clone(m), TerraformModule: &TerraformModuleComponentConfig{VarsMap: maps.Clone(m), EnvVarMap: maps.Clone(m)}},
				{Name: "helm", Type: HelmChartComponentType, HelmChart: &HelmChartComponentConfig{ValuesMap: maps.Clone(m)}},
				{Name: "pulumi", Type: PulumiComponentType, Pulumi: &PulumiComponentConfig{ConfigMap: maps.Clone(m)}},
			}
			cfg.Actions = []*ActionConfig{{Name: "inspect", Labels: maps.Clone(m), Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo done", EnvVarMap: maps.Clone(m)}}}}
			return cfg
		}
		old := makeConfig(drawPermutation(t, keys))
		new := makeConfig(drawPermutation(t, keys))
		require.False(t, new.Diff(old).Summary().HasChanged)
		require.Equal(t, new.Diff(old), new.Diff(old))
		old.Runner.EnvVarMap = nil
		new.Runner.EnvVarMap = map[string]string{}
		require.False(t, new.Diff(old).Summary().HasChanged, "nil and empty optional maps are equivalent")
	})
}

func TestAppConfigDiff_ResourcePresenceProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		names := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 1, 10)
		oldNames := drawDistinctStrings(t, hegel.SampledFrom(names), 0, len(names))
		newNames := drawDistinctStrings(t, hegel.SampledFrom(names), 0, len(names))
		makeConfig := func(names []string) *AppConfig {
			cfg := baseConfig()
			for _, name := range names {
				cfg.Components = append(cfg.Components, &Component{Name: name, Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}})
				cfg.Actions = append(cfg.Actions, &ActionConfig{Name: name, Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo done"}}})
			}
			return cfg
		}
		old, new := makeConfig(oldNames), makeConfig(newNames)
		forward, reverse := new.Diff(old), old.Diff(new)
		for _, name := range names {
			wasPresent, isPresent := slices.Contains(oldNames, name), slices.Contains(newNames, name)
			for _, id := range []diff.NodeID{ComponentResourceID(name), ActionResourceID(name)} {
				if !wasPresent && !isPresent {
					require.Nil(t, forward.FindResource(id))
					continue
				}
				require.NotNil(t, forward.FindResource(id))
				require.NotNil(t, reverse.FindResource(id))
				require.Equal(t, wasPresent != isPresent, forward.FindResource(id).DirectSummary().HasChanged, "%s presence", id)
				require.Equal(t, wasPresent != isPresent, reverse.FindResource(id).DirectSummary().HasChanged, "%s reverse presence", id)
			}
		}
	})
}

func TestAppConfigDiff_MeaningfulEditProperties(t *testing.T) {
	t.Parallel()
	edits := []struct {
		name, section string
		mutate        func(*AppConfig)
	}{
		{"metadata", "description", func(c *AppConfig) { c.Description += " changed" }},
		{"runner", "runner", func(c *AppConfig) { c.Runner.ImageID = "ami-0123456789abcdef0" }},
		{"sandbox", "sandbox", func(c *AppConfig) { c.Sandbox.TerraformVersion = "1.7.0" }},
		{"stack", "stack", func(c *AppConfig) { c.Stack.Description += " changed" }},
		{"input", "inputs", func(c *AppConfig) { c.Inputs.Inputs[0].Description += " changed" }},
		{"secret", "secrets", func(c *AppConfig) { c.Secrets.Secrets[0].Required = !c.Secrets.Secrets[0].Required }},
		{"permission", "permissions", func(c *AppConfig) { c.Permissions.MaintenanceRole.Description += " changed" }},
		{"policy", "policies", func(c *AppConfig) { c.Policies.Policies[0].Contents += " changed" }},
		{"terraform", "components", func(c *AppConfig) { c.Components[0].TerraformModule.TerraformVersion = "1.7.0" }},
		{"helm", "components", func(c *AppConfig) { c.Components[1].HelmChart.Namespace += "-changed" }},
		{"manifest", "components", func(c *AppConfig) { c.Components[2].KubernetesManifest.Manifest += "\n# changed\n" }},
		{"pulumi", "components", func(c *AppConfig) { c.Components[3].Pulumi.Runtime = "python" }},
		{"image", "components", func(c *AppConfig) { c.Components[4].ExternalImage.PublicImageConfig.Tag += "-changed" }},
		{"job", "components", func(c *AppConfig) { c.Components[5].Job.Tag += "-changed" }},
		{"docker", "components", func(c *AppConfig) { c.Components[6].DockerBuild.Dockerfile += "-changed" }},
		{"action", "actions", func(c *AppConfig) { c.Actions[0].Steps[0].InlineContents += "\necho changed\n" }},
		{"runbook", "runbooks", func(c *AppConfig) { c.Runbooks[0].Steps[0].Command += " changed" }},
	}
	hegel.Test(t, func(t *hegel.T) {
		old := generatedDiffConfig(t)
		before, err := json.Marshal(old)
		require.NoError(t, err)
		for _, edit := range edits {
			var new AppConfig
			require.NoError(t, json.Unmarshal(before, &new))
			edit.mutate(&new)
			forward, reverse := new.Diff(old), old.Diff(&new)
			require.True(t, forward.DirectSummary().HasChanged, "%s edit must not disappear", edit.name)
			require.NotEmpty(t, forward.FormatChanged(""), edit.name)
			require.True(t, reverse.DirectSummary().HasChanged, "%s reverse edit must not disappear", edit.name)
			require.False(t, new.Diff(&new).Summary().HasChanged, "%s self-diff", edit.name)
			var changedSections []string
			for _, section := range forward.Children {
				if section.DirectSummary().HasChanged {
					changedSections = append(changedSections, section.Key)
				}
			}
			require.Equal(t, []string{edit.section}, changedSections, "%s must not create unrelated changes", edit.name)
			for _, component := range old.Components {
				got := forward.FindResource(ComponentResourceID(component.Name))
				require.Equal(t, edit.section == "components" && component.Name == edit.name, got.DirectSummary().HasChanged, "%s must not change sibling component %s", edit.name, component.Name)
			}
		}
		after, err := json.Marshal(old)
		require.NoError(t, err)
		require.Equal(t, before, after)
	})
}

func TestAppConfigDiff_OrderedSequenceProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		values := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 2, 8)
		for _, configure := range []struct {
			name string
			set  func(*AppConfig, []string)
		}{
			{"post-deploy runbooks", func(c *AppConfig, v []string) { c.Branch = &AppBranchConfig{PostDeployRunbooks: slices.Clone(v)} }},
			{"Helm values precedence", func(c *AppConfig, v []string) {
				c.Components = ComponentList{{Name: "chart", Type: HelmChartComponentType, HelmChart: &HelmChartComponentConfig{}}}
				for _, value := range v {
					c.Components[0].HelmChart.ValuesFiles = append(c.Components[0].HelmChart.ValuesFiles, HelmValuesFile{Path: value + ".yaml", Contents: "key: " + value})
				}
			}},
			{"Terraform variable precedence", func(c *AppConfig, v []string) {
				c.Components = ComponentList{{Name: "module", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{}}}
				for _, value := range v {
					c.Components[0].TerraformModule.VariablesFiles = append(c.Components[0].TerraformModule.VariablesFiles, TerraformVariablesFile{Contents: "key = " + strconv.Quote(value)})
				}
			}},
			{"job arguments", func(c *AppConfig, v []string) {
				c.Components = ComponentList{{Name: "job", Type: JobComponentType, Job: &JobComponentConfig{Args: slices.Clone(v)}}}
			}},
			{"action steps", func(c *AppConfig, v []string) {
				c.Actions = []*ActionConfig{{Name: "inspect"}}
				for _, value := range v {
					c.Actions[0].Steps = append(c.Actions[0].Steps, &ActionStepConfig{Name: value, Command: "echo " + value})
				}
			}},
			{"runbook steps", func(c *AppConfig, v []string) {
				c.Runbooks = []*RunbookConfig{{Name: "inspect"}}
				for _, value := range v {
					c.Runbooks[0].Steps = append(c.Runbooks[0].Steps, &RunbookStepConfig{Name: value, Type: RunbookStepTypeAction, Command: "echo " + value})
				}
			}},
		} {
			old, new := baseConfig(), baseConfig()
			configure.set(old, values)
			reversed := slices.Clone(values)
			slices.Reverse(reversed)
			configure.set(new, reversed)
			require.True(t, new.Diff(old).Summary().HasChanged, "%s order is meaningful", configure.name)
			require.True(t, old.Diff(new).Summary().HasChanged, configure.name)
			require.NotEmpty(t, new.Diff(old).FormatChanged(""), configure.name)
			permuted := drawPermutation(t, values)
			new = baseConfig()
			configure.set(new, permuted)
			require.Equal(t, !slices.Equal(values, permuted), new.Diff(old).Summary().HasChanged, configure.name)
			require.Equal(t, !slices.Equal(values, permuted), old.Diff(new).Summary().HasChanged, configure.name)
		}
	})
}

func TestAppConfigDiff_ComponentLifecycle(t *testing.T) {
	t.Parallel()
	name := "acme"
	value := "example"
	fixtures := []struct {
		kind ComponentType
		make func() *Component
		key  string
	}{
		{TerraformModuleComponentType, func() *Component {
			return &Component{TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}}
		}, "terraform_version"},
		{HelmChartComponentType, func() *Component {
			return &Component{HelmChart: &HelmChartComponentConfig{ChartName: "chart_" + value}}
		}, "chart_name"},
		{DockerBuildComponentType, func() *Component {
			return &Component{DockerBuild: &DockerBuildComponentConfig{Dockerfile: "Dockerfile_" + value}}
		}, "dockerfile"},
		{ContainerImageComponentType, func() *Component {
			return &Component{ExternalImage: &ExternalImageComponentConfig{PublicImageConfig: &PublicImageConfig{ImageURL: "example.com/" + value, Tag: "v1"}}}
		}, "public_image"},
		{KubernetesManifestComponentType, func() *Component {
			return &Component{KubernetesManifest: &KubernetesManifestComponentConfig{Manifest: "kind: ConfigMap\nmetadata:\n  name: " + value}}
		}, "manifest"},
		{JobComponentType, func() *Component { return &Component{Job: &JobComponentConfig{Cmd: []string{"echo", value}}} }, "cmd"},
		{PulumiComponentType, func() *Component { return &Component{Pulumi: &PulumiComponentConfig{Runtime: "python"}} }, "runtime"},
	}
	makeConfig := func(index int, resourceName string) *AppConfig {
		component := fixtures[index].make()
		component.Name, component.Type = resourceName, fixtures[index].kind
		return &AppConfig{Components: ComponentList{component, {Name: "control_" + name, Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.5.0"}}}, Actions: []*ActionConfig{{Name: name, Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo ready"}}}}}
	}
	for i := range fixtures {
		old := makeConfig(i, name)
		for j := range fixtures {
			next := makeConfig(j, name)
			result := next.Diff(old)
			require.Equal(t, i != j, result.FindResource(ComponentResourceID(name)).DirectSummary().HasChanged)
			require.False(t, result.FindResource(ComponentResourceID("control_"+name)).Summary().HasChanged)
			require.False(t, result.FindResource(ActionResourceID(name)).Summary().HasChanged, "resource namespaces must not alias")
			if i != j {
				require.Equal(t, diff.OpChange, findChild(result.FindResource(ComponentResourceID(name)), "type").Diff.Op)
				require.True(t, findChild(result.FindResource(ComponentResourceID(name)), fixtures[i].key).DirectSummary().HasChanged, "old %s payload removal", fixtures[i].kind)
				require.True(t, findChild(result.FindResource(ComponentResourceID(name)), fixtures[j].key).DirectSummary().HasChanged, "new %s payload addition", fixtures[j].kind)
			}
		}
		renamed := makeConfig(i, "renamed_"+name).Diff(old)
		require.True(t, renamed.FindResource(ComponentResourceID(name)).DirectSummary().HasChanged)
		require.True(t, renamed.FindResource(ComponentResourceID("renamed_"+name)).DirectSummary().HasChanged)
		for _, pair := range [][2]*AppConfig{{{}, old}, {old, {}}} {
			result := pair[1].Diff(pair[0])
			require.True(t, result.FindResource(ComponentResourceID(name)).DirectSummary().HasChanged)
			require.NotEmpty(t, result.FormatChanged(""))
		}
	}
}

func TestAppConfigDiff_ResourceLifecycle(t *testing.T) {
	name := "acme"
	for _, resource := range []struct {
		section string
		set     func(*AppConfig, string)
	}{
		{"inputs", func(c *AppConfig, name string) {
			c.Inputs = &AppInputConfig{Inputs: []AppInput{{Name: name, Type: "string", Description: "region", DisplayName: "Region", Group: "deployment"}}}
		}},
		{"secrets", func(c *AppConfig, name string) {
			c.Secrets = &SecretsConfig{Secrets: []*AppSecret{{Name: name, Description: "token"}}}
		}},
		{"actions", func(c *AppConfig, name string) {
			c.Actions = []*ActionConfig{{Name: name, Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo ready"}}}}
		}},
		{"runbooks", func(c *AppConfig, name string) {
			c.Runbooks = []*RunbookConfig{{Name: name, Steps: []*RunbookStepConfig{{Name: "inspect", Type: RunbookStepTypeAction, Command: "echo ready"}}}}
		}},
		{"installs", func(c *AppConfig, name string) { c.Installs = []*Install{{Name: name}} }},
		{"branches", func(c *AppConfig, name string) { c.Branches = []*AppBranchConfig{{Name: name}} }},
		{"policies", func(c *AppConfig, name string) {
			c.Policies = &PoliciesConfig{Policies: []AppPolicy{{Name: name, Type: AppPolicyTypeKubernetesCluster, Engine: AppPolicyEngineOPA, Contents: "package acme\n"}}}
		}},
		{"kubernetes_contexts", func(c *AppConfig, name string) {
			c.KubernetesContexts = &KubernetesContextsConfig{Contexts: []*KubernetesContext{{Name: name, Component: "cluster"}}}
		}},
	} {
		old, renamed := &AppConfig{}, &AppConfig{}
		resource.set(old, name)
		resource.set(renamed, "renamed_"+name)
		for _, pair := range [][2]*AppConfig{{{}, old}, {old, {}}, {old, renamed}} {
			result := pair[1].Diff(pair[0])
			require.True(t, findChild(result, resource.section).DirectSummary().HasChanged, "%s membership must be observable, including minimally populated members", resource.section)
			require.NotEmpty(t, result.FormatChanged(""))
		}
		require.False(t, old.Diff(old).Summary().HasChanged)
	}
}

func TestAppConfigDiff_SharedStorage(t *testing.T) {
	value := "acme"
	for _, changed := range []bool{false, true} {
		dependencies := []string{"zulu", "alpha", "unused"}
		labels := map[string]string{"region": value}
		role := &AppAWSIAMRole{Name: "maintenance", Description: value}
		references := []refs.Ref{{Type: refs.RefTypeInputs, Name: "region", Input: "nuon.inputs.inputs.region"}}
		old := &AppConfig{
			Components:  ComponentList{{Name: "api", Type: TerraformModuleComponentType, Dependencies: dependencies[:2], Labels: labels, SourceFile: value + ".toml", Checksum: value, References: references, TerraformModule: &TerraformModuleComponentConfig{VarsMap: labels}}},
			Permissions: &PermissionsConfig{MaintenanceRole: role, Roles: []*AppAWSIAMRole{role}},
			Actions:     []*ActionConfig{{Name: "inspect", Labels: labels, Dependencies: dependencies[:2], References: references, Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo ready", EnvVarMap: labels, References: references}}}},
		}
		next := *old
		component := *old.Components[0]
		next.Components = ComponentList{&component}
		component.Dependencies = dependencies[1:2]
		if changed {
			component.VarName = "changed_" + value
		}
		beforeOld, err := copystructure.Copy(old)
		require.NoError(t, err)
		beforeNext, err := copystructure.Copy(&next)
		require.NoError(t, err)
		result := next.Diff(old)
		require.Equal(t, beforeNext.(*AppConfig).Diff(beforeOld.(*AppConfig)), result, "alias layout must not affect observations")
		require.Equal(t, beforeOld, old, "transient as well as serialized state must be unchanged")
		require.Equal(t, beforeNext, &next)
		require.Equal(t, []string{"zulu", "alpha", "unused"}, dependencies, "overlapping backing array must remain untouched")
		require.Equal(t, result, next.Diff(old))
		require.False(t, old.Diff(old).Summary().HasChanged)
		snapshot, err := copystructure.Copy(result)
		require.NoError(t, err)
		result.FindResource(ComponentResourceID("api")).Diff.Before = "corrupted result"
		require.Equal(t, snapshot, next.Diff(old), "returned trees must not share storage")
		require.Equal(t, beforeOld, old)
		require.Equal(t, beforeNext, &next)
	}
}

func drawPermutation[T any](t hegel.TestCase, values []T) []T {
	result := slices.Clone(values)
	for i := len(result) - 1; i > 0; i-- {
		j := hegel.Draw(t, hegel.Integers(0, i))
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func drawDistinctStrings(t hegel.TestCase, values hegel.Generator[string], min, max int) []string {
	set := hegel.Draw(t, hegel.Maps(values, hegel.Just(true)).MinSize(min).MaxSize(max))
	return slices.Sorted(maps.Keys(set))
}

func generatedDiffConfig(t *hegel.T) *AppConfig {
	text := hegel.Draw(t, hegel.Text())
	labels := hegel.Draw(t, hegel.Maps(hegel.FromRegex(`[a-z]{1,8}`, true), hegel.Text()).MaxSize(4))
	cfg := baseConfig()
	cfg.Description = text
	cfg.Runner.EnvVarMap = maps.Clone(labels)
	cfg.Sandbox.VarsMap = maps.Clone(labels)
	cfg.Stack = &StackConfig{Name: "acme", Description: text}
	cfg.Inputs = &AppInputConfig{Inputs: []AppInput{{Name: "region", Type: "string", Default: text, Description: text}}}
	cfg.Secrets = &SecretsConfig{Secrets: []*AppSecret{{Name: "api_key", Required: hegel.Draw(t, hegel.Booleans())}}}
	cfg.Permissions = &PermissionsConfig{MaintenanceRole: &AppAWSIAMRole{Name: "maintenance", Description: text}}
	cfg.Policies = &PoliciesConfig{Policies: []AppPolicy{{Name: "acme", Type: AppPolicyTypeKubernetesCluster, Engine: AppPolicyEngineOPA, Contents: text}}}
	cfg.Components = ComponentList{
		{Name: "terraform", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0", VarsMap: maps.Clone(labels)}},
		{Name: "helm", Type: HelmChartComponentType, HelmChart: &HelmChartComponentConfig{ChartName: "acme", Namespace: "default", ValuesMap: maps.Clone(labels)}},
		{Name: "manifest", Type: KubernetesManifestComponentType, KubernetesManifest: &KubernetesManifestComponentConfig{Manifest: text}},
		{Name: "pulumi", Type: PulumiComponentType, Pulumi: &PulumiComponentConfig{Runtime: "go", ConfigMap: maps.Clone(labels)}},
		{Name: "image", Type: ContainerImageComponentType, ExternalImage: &ExternalImageComponentConfig{PublicImageConfig: &PublicImageConfig{ImageURL: "example.com/acme", Tag: "v1"}}},
		{Name: "job", Type: JobComponentType, Job: &JobComponentConfig{ImageURL: "example.com/acme", Tag: "v1", Cmd: []string{"sh", "-c"}, Args: []string{text}}},
		{Name: "docker", Type: DockerBuildComponentType, DockerBuild: &DockerBuildComponentConfig{Dockerfile: "Dockerfile", EnvVarMap: maps.Clone(labels)}},
	}
	for _, name := range drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 1, 5) {
		cfg.Actions = append(cfg.Actions, &ActionConfig{Name: name, Labels: maps.Clone(labels), Steps: []*ActionStepConfig{{Name: "inspect", InlineContents: text, EnvVarMap: maps.Clone(labels)}}})
		cfg.Runbooks = append(cfg.Runbooks, &RunbookConfig{Name: name, Description: text, Steps: []*RunbookStepConfig{{Name: "inspect", Type: RunbookStepTypeAction, Command: "echo " + text}}})
	}
	return cfg
}

// --- Test: Nil handling ---

func (s *AppConfigDiffSuite) TestNilHandling() {
	// Diff against nil old
	a := baseConfig()
	d := a.Diff(nil)
	s.Require().NotNil(d)
	summary := d.Summary()
	s.True(summary.HasChanged)

	// Config with nil sub-configs
	minimal := &AppConfig{Version: "1"}
	d = minimal.Diff(&AppConfig{Version: "1"})
	s.Require().NotNil(d)
	s.False(d.Summary().HasChanged)
}

// --- Test: Combined changes ---

func (s *AppConfigDiffSuite) TestCombinedChanges() {
	old := baseConfig()
	old.Components = ComponentList{
		{Name: "db", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.5.0"}},
	}
	old.Installs = []*Install{
		{Name: "prod", AWSAccount: &AWSAccount{Region: "us-east-1"}},
	}

	new := baseConfig()
	new.Description = "Changed"
	new.Runner.EnvVarMap = map[string]string{"NEW": "val"}
	new.Components = ComponentList{
		{Name: "db", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}},
		{Name: "api", Type: HelmChartComponentType, HelmChart: &HelmChartComponentConfig{ChartName: "api"}},
	}
	new.Installs = []*Install{
		{Name: "prod", AWSAccount: &AWSAccount{Region: "eu-west-1"}},
	}

	d := new.Diff(old)
	summary := d.Summary()
	s.True(summary.HasChanged)
	s.Greater(summary.Changed, 0)
	s.Greater(summary.Added, 0)
}

// --- Test: String output ---

func (s *AppConfigDiffSuite) TestDiffStringOutput() {
	old := baseConfig()
	new := baseConfig()
	new.Version = "2"

	d := new.Diff(old)
	output := d.String("")
	s.Contains(output, "app_config:")
	s.Contains(output, "version:")
	s.Contains(output, "'1' -> '2'")

	// FormatChanged omits unchanged sections and adds prefix markers
	changed := d.FormatChanged("")
	s.Contains(changed, "version:")
	s.Contains(changed, "'1' -> '2'")
	s.Contains(changed, "~ ")
	s.NotContains(changed, "(unchanged)")
}

// --- Test: Break glass ---

func (s *AppConfigDiffSuite) TestBreakGlassChanges() {
	old := baseConfig()
	old.BreakGlass = &BreakGlass{
		Roles: []*AppAWSIAMRole{
			{Name: "emergency", Description: "Emergency access"},
		},
	}
	new := baseConfig()
	new.BreakGlass = &BreakGlass{
		Roles: []*AppAWSIAMRole{
			{Name: "emergency", Description: "Updated emergency access"},
		},
	}

	d := new.Diff(old)
	bg := findChild(d, "break_glass")
	s.Require().NotNil(bg)

	role := findChild(bg, "role.emergency")
	s.Require().NotNil(role)
	desc := findChild(role, "description")
	s.Equal(diff.OpChange, desc.Diff.Op)
}

// --- Test: Stack changes ---

func (s *AppConfigDiffSuite) TestStackChanges() {
	old := baseConfig()
	old.Stack = &StackConfig{
		Name:        "my-stack",
		Description: "Original",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "vpc", TemplateURL: "https://s3.amazonaws.com/bucket/vpc.yaml"},
		},
	}
	new := baseConfig()
	new.Stack = &StackConfig{
		Name:        "my-stack",
		Description: "Updated",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "vpc", TemplateURL: "https://s3.amazonaws.com/bucket/vpc-v2.yaml"},
			{Name: "eks", TemplateURL: "https://s3.amazonaws.com/bucket/eks.yaml"},
		},
	}

	d := new.Diff(old)
	stack := findChild(d, "stack")
	s.Require().NotNil(stack)

	desc := findChild(stack, "description")
	s.Equal(diff.OpChange, desc.Diff.Op)

	vpc := findChild(stack, "custom_stack.vpc")
	s.Require().NotNil(vpc)
	tplURL := findChild(vpc, "template_url")
	s.Equal(diff.OpChange, tplURL.Diff.Op)

	eks := findChild(stack, "custom_stack.eks")
	s.Require().NotNil(eks)
	s.Equal(diff.OpAdd, eks.Diff.Op)
}

// --- Helpers ---

// findChild searches for a child with the given key in a diff tree (one level deep).
func findChild(d *diff.Diff, key string) *diff.Diff {
	if d == nil {
		return nil
	}
	for _, c := range d.Children {
		if c.Key == key {
			return c
		}
	}
	return nil
}
