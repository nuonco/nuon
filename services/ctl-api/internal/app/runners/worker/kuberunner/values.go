package runner

// NOTE(jm): this struct must match the yaml expected in the helm chart to install the runner.
type helmValuesImage struct {
	Tag        string `mapstructure:"tag"`
	Repository string `mapstructure:"repository"`
}

type serviceAccountValues struct {
	Name        string            `mapstructure:"name"`
	Annotations map[string]string `mapstructure:"annotations"`
}

type helmValuesEnv struct {
	RunnerID               string `mapstructure:"RUNNER_ID"`
	RunnerAPIToken         string `mapstructure:"RUNNER_API_TOKEN"`
	APIURL                 string `mapstructure:"RUNNER_API_URL"`
	SettingsRefreshTimeout string `mapstructure:"SETTINGS_REFRESH_TIMEOUT"`
}

type instanceType struct {
	Name string `mapstructure:"name"`
}

type nodePoolValues struct {
	Enabled      bool         `mapstructure:"enabled"`
	InstanceType instanceType `mapstructure:"instance_type"`
}

type helmValues struct {
	Image helmValuesImage `mapstructure:"image"`
	Env   helmValuesEnv   `mapstructure:"env"`

	ServiceAccount serviceAccountValues `mapstructure:"serviceAccount"`
	NodePool       nodePoolValues       `mapstructure:"node_pool"`
}

func (a *Activities) getValues(req *InstallOrUpgradeRequest) helmValues {
	annotations := map[string]string{}
	enableNodePool := true
	switch req.CloudProvider {
	case "gcp":
		if req.RunnerIAMRole != "" {
			annotations["iam.gke.io/gcp-service-account"] = req.RunnerIAMRole
		}
		// GKE uses Autopilot or node auto-provisioning, not Karpenter
		enableNodePool = false
	default:
		if req.RunnerIAMRole != "" {
			annotations["eks.amazonaws.com/role-arn"] = req.RunnerIAMRole
		}
	}

	return helmValues{
		Image: helmValuesImage{
			Tag:        req.Image.Tag,
			Repository: req.Image.URL,
		},
		Env: helmValuesEnv{
			RunnerID:               req.RunnerID,
			RunnerAPIToken:         req.APIToken,
			APIURL:                 req.APIURL,
			SettingsRefreshTimeout: req.SettingsRefreshTimeout.String(),
		},
		ServiceAccount: serviceAccountValues{
			Name:        req.RunnerServiceAccountName,
			Annotations: annotations,
		},
		NodePool: nodePoolValues{
			Enabled: enableNodePool,
			InstanceType: instanceType{
				Name: req.InstanceTypeName,
			},
		},
	}
}
