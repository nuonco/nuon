package runner

import (
	"fmt"
	"strings"
)

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

type nodeClassRefValues struct {
	Group string `mapstructure:"group"`
	Kind  string `mapstructure:"kind"`
	Name  string `mapstructure:"name"`
}

type nodePoolValues struct {
	Enabled      bool               `mapstructure:"enabled"`
	InstanceType instanceType       `mapstructure:"instance_type"`
	NodeClassRef nodeClassRefValues `mapstructure:"node_class_ref"`
}

type helmValues struct {
	Image helmValuesImage `mapstructure:"image"`
	Env   helmValuesEnv   `mapstructure:"env"`

	ServiceAccount serviceAccountValues `mapstructure:"serviceAccount"`
	PodLabels      map[string]string    `mapstructure:"podLabels"`
	NodePool       nodePoolValues       `mapstructure:"node_pool"`
}

func (a *Activities) getValues(req *InstallOrUpgradeRequest) helmValues {
	annotations := map[string]string{}
	podLabels := map[string]string{}
	enableNodePool := true
	nodeClassRef := nodeClassRefValues{
		Group: "karpenter.k8s.aws",
		Kind:  "EC2NodeClass",
		Name:  "default",
	}
	switch req.CloudProvider {
	case "gcp":
		if req.RunnerIAMRole != "" {
			annotations["iam.gke.io/gcp-service-account"] = req.RunnerIAMRole
		}
		// GKE uses Autopilot or node auto-provisioning, not Karpenter
		enableNodePool = false
	case "azure":
		if req.RunnerIAMRole != "" {
			annotations["azure.workload.identity/client-id"] = req.RunnerIAMRole
		}
		podLabels["azure.workload.identity/use"] = "true"
		// AKS uses Node Auto-Provisioning (NAP) with Karpenter
		nodeClassRef = nodeClassRefValues{
			Group: "karpenter.azure.com",
			Kind:  "AKSNodeClass",
			Name:  "default",
		}
	default:
		orgID := strings.TrimPrefix(req.RunnerServiceAccountName, "runner-")
		if orgID != "" && a.config.ManagementAccountID != "" {
			annotations["eks.amazonaws.com/role-arn"] = fmt.Sprintf(
				"arn:aws:iam::%s:role/orgs/%s/runner-%s",
				a.config.ManagementAccountID, orgID, orgID,
			)
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
		PodLabels: podLabels,
		NodePool: nodePoolValues{
			Enabled: enableNodePool,
			InstanceType: instanceType{
				Name: req.InstanceTypeName,
			},
			NodeClassRef: nodeClassRef,
		},
	}
}
