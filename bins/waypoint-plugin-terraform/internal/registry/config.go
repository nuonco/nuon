package registry

import "github.com/powertoolsdev/mono/pkg/aws/credentials"

type Config struct {
	// settings to assume a role and access ECR
	Region string             `hcl:"region,optional"`
	Auth   credentials.Config `hcl:"auth,block"`

	// repository
	Repository string `hcl:"repository,optional"`
	Tag        string `hcl:"tag,attr"`
}

func (r *Registry) Config() (interface{}, error) {
	return &r.config, nil
}
