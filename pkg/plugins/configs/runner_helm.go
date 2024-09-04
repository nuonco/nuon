package configs

import (
	"github.com/powertoolsdev/mono/pkg/kube"
)

// RunnerHelm is used to configure the runner to deploy the runner helm chart, which is bundled within it.
type RunnerHelm struct {
	Plugin string `hcl:"plugin,label"`

	Name       string    `hcl:"name,attr"`
	Repository string    `hcl:"repository"`
	Chart      string    `hcl:"chart"`
	Version    string    `hcl:"version"`
	Devel      bool      `hcl:"devel,optional"`
	Values     []string  `hcl:"values,optional"`
	HelmSet    []HelmSet `hcl:"set,block"`
	Driver     string    `hcl:"driver,optional"`
	Namespace  string    `hcl:"namespace,optional"`

	CreateNamespace bool `hcl:"create_namespace,optional"`
	SkipCRDs        bool `hcl:"skip_crds,optional"`

	ClusterInfo *kube.ClusterInfo `hcl:"cluster_info,block"`
}
