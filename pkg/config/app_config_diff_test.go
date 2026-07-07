package config

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/stretchr/testify/suite"
)

type AppConfigDiffSuite struct {
	suite.Suite
}

func TestAppConfigDiffSuite(t *testing.T) {
	suite.Run(t, new(AppConfigDiffSuite))
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

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
