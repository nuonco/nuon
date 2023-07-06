package configs

type HelmSet struct {
	Name  string `hcl:"name"`
	Value string `hcl:"value"`
	Type  string `hcl:"type,optional"`
}

type HelmRepoDeploy struct {
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

	KubeconfigPath  string     `hcl:"kubeconfig,optional"`
	Context         string     `hcl:"context,optional"`
	CreateNamespace bool       `hcl:"create_namespace,optional"`
	SkipCRDs        bool       `hcl:"skip_crds,optional"`
	Archive         OciArchive `hcl:"archive,block"`
}
