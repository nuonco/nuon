package build

import (
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
)

// VCS carries git sources the caller already resolved, which needs a database.
type VCS struct {
	Github *app.ConnectedGithubVCSConfig
	Public *app.PublicGitVCSConfig
}

func DockerBuildComponentConfig(obj *config.DockerBuildComponentConfig, vcs VCS) (*app.DockerBuildComponentConfig, error) {
	if obj == nil {
		return nil, errors.New("docker_build config is required")
	}
	if vcs.Github == nil && vcs.Public == nil {
		return nil, errors.New("docker_build requires either a connected repo or public repo configuration")
	}

	dockerfile := obj.Dockerfile
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	return &app.DockerBuildComponentConfig{
		Dockerfile:               dockerfile,
		EnvVars:                  Hstore(obj.EnvVarMap),
		PublicGitVCSConfig:       vcs.Public,
		ConnectedGithubVCSConfig: vcs.Github,
	}, nil
}

func HelmComponentConfig(obj *config.HelmChartComponentConfig, vcs VCS) (*app.HelmComponentConfig, error) {
	if obj == nil {
		return nil, errors.New("helm_chart config is required")
	}
	if vcs.Github == nil && vcs.Public == nil && obj.HelmRepo == nil {
		return nil, errors.New("helm_chart requires either a VCS configuration or Helm repository configuration")
	}
	if err := validation.ValidateDNSName(obj.ChartName, 5, 62); err != nil {
		return nil, err
	}

	var helmRepoConfig *app.HelmRepoConfig
	if obj.HelmRepo != nil {
		helmRepoConfig = &app.HelmRepoConfig{
			RepoURL: obj.HelmRepo.RepoURL,
			Chart:   obj.HelmRepo.Chart,
			Version: obj.HelmRepo.Version,
		}
	}

	valuesFiles := make([]string, 0, len(obj.ValuesFiles))
	for _, vf := range obj.ValuesFiles {
		if vf.Path != "" {
			valuesFiles = append(valuesFiles, vf.Path)
		} else if vf.Contents != "" {
			valuesFiles = append(valuesFiles, vf.Contents)
		}
	}

	return &app.HelmComponentConfig{
		PublicGitVCSConfig:       vcs.Public,
		ConnectedGithubVCSConfig: vcs.Github,
		HelmConfig: &app.HelmConfig{
			ChartName:      obj.ChartName,
			Namespace:      obj.Namespace,
			Values:         Hstore(obj.ValuesMap),
			ValuesFiles:    valuesFiles,
			StorageDriver:  obj.StorageDriver,
			TakeOwnership:  obj.TakeOwnership,
			SkipCRDs:       obj.SkipCRDs,
			HelmRepoConfig: helmRepoConfig,
		},
	}, nil
}

func TerraformModuleComponentConfig(obj *config.TerraformModuleComponentConfig, vcs VCS, version string) (*app.TerraformModuleComponentConfig, error) {
	if obj == nil {
		return nil, errors.New("terraform_module config is required")
	}
	if vcs.Github == nil && vcs.Public == nil {
		return nil, errors.New("terraform_module requires either a connected repo or public repo configuration")
	}
	if obj.TerraformVersion != "" {
		if err := validation.ValidateTerraformMinVersion(obj.TerraformVersion); err != nil {
			return nil, err
		}
	}

	variablesFiles := make([]string, 0, len(obj.VariablesFiles))
	for _, vf := range obj.VariablesFiles {
		if vf.Contents != "" {
			variablesFiles = append(variablesFiles, vf.Contents)
		}
	}

	return &app.TerraformModuleComponentConfig{
		PublicGitVCSConfig:       vcs.Public,
		ConnectedGithubVCSConfig: vcs.Github,
		Version:                  version,
		Variables:                Hstore(obj.VarsMap),
		EnvVars:                  Hstore(obj.EnvVarMap),
		VariablesFiles:           pq.StringArray(variablesFiles),
	}, nil
}

func PulumiComponentConfig(obj *config.PulumiComponentConfig, vcs VCS) (*app.PulumiComponentConfig, error) {
	if obj == nil {
		return nil, errors.New("pulumi config is required")
	}
	if vcs.Github == nil && vcs.Public == nil {
		return nil, errors.New("pulumi requires either a connected repo or public repo configuration")
	}

	return &app.PulumiComponentConfig{
		Runtime:                  obj.Runtime,
		Version:                  obj.PulumiVersion,
		PublicGitVCSConfig:       vcs.Public,
		ConnectedGithubVCSConfig: vcs.Github,
		Config:                   Hstore(obj.ConfigMap),
		EnvVars:                  Hstore(obj.EnvVarMap),
	}, nil
}

func KubernetesManifestComponentConfig(obj *config.KubernetesManifestComponentConfig, vcs VCS) (*app.KubernetesManifestComponentConfig, error) {
	if obj == nil {
		return nil, errors.New("kubernetes_manifest config is required")
	}

	hasManifest := obj.Manifest != ""
	hasKustomize := obj.Kustomize != nil && obj.Kustomize.Path != ""
	if !hasManifest && !hasKustomize {
		return nil, errors.New("one of 'manifest' or 'kustomize' must be specified")
	}
	if hasManifest && hasKustomize {
		return nil, errors.New("only one of 'manifest' or 'kustomize' can be specified")
	}
	if hasKustomize && vcs.Github == nil && vcs.Public == nil {
		return nil, errors.New("kustomize requires a git source")
	}
	if hasManifest && (vcs.Github != nil || vcs.Public != nil) {
		return nil, errors.New("VCS config is only valid for kustomize sources, not inline manifests")
	}

	cfg := &app.KubernetesManifestComponentConfig{
		Manifest:                 obj.Manifest,
		Namespace:                obj.Namespace,
		PublicGitVCSConfig:       vcs.Public,
		ConnectedGithubVCSConfig: vcs.Github,
	}

	if hasKustomize {
		cfg.Kustomize = &app.KustomizeConfig{
			Path:           obj.Kustomize.Path,
			Patches:        obj.Kustomize.Patches,
			EnableHelm:     obj.Kustomize.EnableHelm,
			LoadRestrictor: obj.Kustomize.LoadRestrictor,
		}
	}

	return cfg, nil
}

func JobComponentConfig(obj *config.JobComponentConfig) (*app.JobComponentConfig, error) {
	if obj == nil {
		return nil, errors.New("job config is required")
	}
	if obj.ImageURL == "" {
		return nil, errors.New("image_url is required")
	}

	envVars := make(map[string]string, len(obj.EnvVarMap)+len(obj.EnvVars))
	for _, v := range obj.EnvVars {
		envVars[v.Name] = v.Value
	}
	for k, v := range obj.EnvVarMap {
		envVars[k] = v
	}

	return &app.JobComponentConfig{
		ImageURL: obj.ImageURL,
		Tag:      obj.Tag,
		Cmd:      pq.StringArray(obj.Cmd),
		Args:     pq.StringArray(obj.Args),
		EnvVars:  Hstore(envVars),
	}, nil
}

func ExternalImageComponentConfig(obj *config.ExternalImageComponentConfig) (*app.ExternalImageComponentConfig, error) {
	if obj == nil {
		return nil, errors.New("external_image config is required")
	}

	if err := obj.Verification.Validate(); err != nil {
		return nil, err
	}

	cfg := &app.ExternalImageComponentConfig{Verification: obj.Verification}
	sources := 0

	if src := obj.AWSECRImageConfig; src != nil {
		sources++
		cfg.ImageURL = src.ImageURL
		cfg.Tag = src.Tag
		cfg.UpdatePolicy = src.UpdatePolicy
		cfg.AWSECRImageConfig = &app.AWSECRImageConfig{
			IAMRoleARN: src.IAMRoleARN,
			AWSRegion:  src.AWSRegion,
		}
	}
	if src := obj.GCPGARImageConfig; src != nil {
		sources++
		cfg.ImageURL = src.ImageURL
		cfg.Tag = src.Tag
		cfg.UpdatePolicy = src.UpdatePolicy
		cfg.GCPGARImageConfig = &app.GCPGARImageConfig{
			GCPProjectID:             src.GCPProjectID,
			GCPRegion:                src.GCPRegion,
			ServiceAccountEmail:      src.ServiceAccountEmail,
			WorkloadIdentityProvider: src.WorkloadIdentityProvider,
		}
	}
	if src := obj.AzureACRImageConfig; src != nil {
		sources++
		if err := src.ValidateCredentials(); err != nil {
			return nil, err
		}
		cfg.ImageURL = src.ImageURL
		cfg.Tag = src.Tag
		cfg.UpdatePolicy = src.UpdatePolicy
		cfg.AzureACRImageConfig = &app.AzureACRImageConfig{
			RegistryURL:           src.RegistryURL,
			TenantID:              src.TenantID,
			ClientID:              src.ClientID,
			ClientSecretName:      src.ClientSecretName,
			ClientCertificateName: src.ClientCertificateName,
		}
	}
	if src := obj.PublicImageConfig; src != nil {
		sources++
		cfg.ImageURL = src.ImageURL
		cfg.Tag = src.Tag
		cfg.UpdatePolicy = src.UpdatePolicy
	}

	switch sources {
	case 0:
		return nil, errors.New("external_image requires one of aws_ecr, gcp_gar, azure_acr, or public image source configuration")
	case 1:
	default:
		return nil, errors.New("external_image must specify exactly one of aws_ecr, gcp_gar, azure_acr, or public image source configuration")
	}

	return cfg, nil
}

// AttachTypeConfig wires the per-type config onto the shared connection.
func AttachTypeConfig(ccc *app.ComponentConfigConnection, comp *config.Component, vcs VCS, terraformVersion string) error {
	switch {
	case comp.DockerBuild != nil:
		cfg, err := DockerBuildComponentConfig(comp.DockerBuild, vcs)
		if err != nil {
			return err
		}
		ccc.DockerBuildComponentConfig = cfg
	case comp.HelmChart != nil:
		cfg, err := HelmComponentConfig(comp.HelmChart, vcs)
		if err != nil {
			return err
		}
		ccc.HelmComponentConfig = cfg
	case comp.TerraformModule != nil:
		cfg, err := TerraformModuleComponentConfig(comp.TerraformModule, vcs, terraformVersion)
		if err != nil {
			return err
		}
		ccc.TerraformModuleComponentConfig = cfg
	case comp.Pulumi != nil:
		cfg, err := PulumiComponentConfig(comp.Pulumi, vcs)
		if err != nil {
			return err
		}
		ccc.PulumiComponentConfig = cfg
	case comp.KubernetesManifest != nil:
		cfg, err := KubernetesManifestComponentConfig(comp.KubernetesManifest, vcs)
		if err != nil {
			return err
		}
		ccc.KubernetesManifestComponentConfig = cfg
	case comp.Job != nil:
		cfg, err := JobComponentConfig(comp.Job)
		if err != nil {
			return err
		}
		ccc.JobComponentConfig = cfg
	case comp.ExternalImage != nil:
		cfg, err := ExternalImageComponentConfig(comp.ExternalImage)
		if err != nil {
			return err
		}
		ccc.ExternalImageComponentConfig = cfg
	default:
		return fmt.Errorf("component %q has no type configuration", comp.Name)
	}

	return nil
}
