package registry

type Config struct {
	// settings to assume a role and access ECR
	Region  string `hcl:"region,optional"`
	RoleARN string `hcl:"role_arn"`

	// repository
	Repository string `hcl:"repository,optional"`
	Tag        string `hcl:"tag,attr"`
}

func (r *Registry) Config() (interface{}, error) {
	return &r.config, nil
}
