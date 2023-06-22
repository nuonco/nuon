package configs

type HelmSet struct {
	Name  string `hcl:"name"`
	Value string `hcl:"value"`
}

type HelmRepoDeploy struct {
	Plugin string `hcl:"plugin,label"`

	Name            string    `hcl:"name,attr"`
	Repository      string    `hcl:"repository"`
	Chart           string    `hcl:"chart"`
	Version         string    `hcl:"version"`
	Namespace       string    `hcl:"namespace"`
	CreateNamespace bool      `hcl:"create_namespace,optional"`
	Values          []string  `hcl:"values,optional"`
	HelmSet         []HelmSet `hcl:"set,block"`
}
