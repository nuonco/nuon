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

type nodePoolValues struct {
	Enabled bool `mapstructure:"enabled"`
}

type helmValues struct {
	Image helmValuesImage `mapstructure:"image"`
	Env   helmValuesEnv   `mapstructure:"env"`

	ServiceAccount serviceAccountValues `mapstructure:"serviceAccount"`
	NodePool       nodePoolValues       `mapstructure:"node_pool"`
}

func (a *Activities) getValues(req *InstallOrUpgradeRequest) helmValues {
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
			Name: req.RunnerServiceAccountName,
			Annotations: map[string]string{
				"eks.amazonaws.com/role-arn": req.RunnerIAMRole,
			},
		},
		NodePool: nodePoolValues{
			Enabled: true,
		},
	}
}
