package configs

import "github.com/powertoolsdev/mono/pkg/aws/credentials"

type OCIRegistry struct {
	Plugin string `hcl:"plugin,label"`

	Repository string             `hcl:"repository"`
	Tag        string             `hcl:"tag"`
	Region     string             `hcl:"region"`
	Auth       credentials.Config `hcl:"auth,block"`
}
