package protos

import (
	"fmt"
	"time"

	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	variablesv1 "github.com/powertoolsdev/mono/pkg/types/components/variables/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	defaultTerraformModuleDeployTimeout time.Duration = time.Minute * 30
)

func (c *Adapter) toTerraformVariables(inputVals map[string]*string) *variablesv1.TerraformVariables {
	vals := make([]*variablesv1.TerraformVariable, 0)
	for k, v := range inputVals {
		if v == nil {
			continue
		}

		vals = append(vals, &variablesv1.TerraformVariable{
			Name:      k,
			Value:     *v,
			Sensitive: true,
		})
	}

	return &variablesv1.TerraformVariables{
		Variables: vals,
	}
}

func (c *Adapter) awsAccess(install app.Install) *deployv1.AwsAccess {
	if install.AWSAccount == nil {
		return nil
	}

	return &deployv1.AwsAccess{
		Region: install.AWSAccount.Region,
	}
}

func (c *Adapter) azureAccess(install app.Install) *deployv1.AzureAccess {
	if install.AzureAccount == nil {
		return nil
	}

	return &deployv1.AzureAccess{
		Location:       install.AzureAccount.Location,
		SubscriptionId: install.AzureAccount.SubscriptionID,
		TenantId:       install.AzureAccount.SubscriptionTenantID,
		ClientId:       install.AzureAccount.ServicePrincipalAppID,
		ClientSecret:   install.AzureAccount.ServicePrincipalPassword,
	}
}

func (c *Adapter) ToTerraformModuleComponentConfig(cfg *app.TerraformModuleComponentConfig, connections []app.InstallDeploy, gitRef string, installDeploy *app.InstallDeploy) (*componentv1.Component, error) {
	vcsCfg, err := c.ToVCSConfig(gitRef, cfg.PublicGitVCSConfig, cfg.ConnectedGithubVCSConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get vcs config: %w", err)
	}

	var (
		awsAccess   *deployv1.AwsAccess
		azureAccess *deployv1.AzureAccess
	)
	if installDeploy != nil {
		awsAccess = c.awsAccess(installDeploy.InstallComponent.Install)
		azureAccess = c.azureAccess(installDeploy.InstallComponent.Install)
	}

	return &componentv1.Component{
		Id: cfg.ComponentConfigConnection.ComponentID,
		BuildCfg: &buildv1.Config{
			Timeout: durationpb.New(defaultBuildTimeout),
			Cfg: &buildv1.Config_TerraformModuleCfg{
				TerraformModuleCfg: &buildv1.TerraformModuleConfig{
					VcsCfg: vcsCfg,
				},
			},
		},
		DeployCfg: &deployv1.Config{
			Timeout: durationpb.New(defaultTerraformModuleDeployTimeout),
			Cfg: &deployv1.Config_TerraformModuleConfig{
				TerraformModuleConfig: &deployv1.TerraformModuleConfig{
					TerraformVersion: cfg.Version,
					Vars:             c.toTerraformVariables(cfg.Variables),
					EnvVars:          c.toEnvVars(cfg.EnvVars),
					AzureAccess:      azureAccess,
					AwsAccess:        awsAccess,
				},
			},
		},
		Connections: c.toConnections(connections),
	}, nil
}
