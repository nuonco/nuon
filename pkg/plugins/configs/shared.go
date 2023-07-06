package configs

import "github.com/powertoolsdev/mono/pkg/aws/credentials"

type OciArchive struct {
	Image string             `hcl:"image"`
	Tag   string             `hcl:"tag"`
	Auth  credentials.Config `hcl:"auth" validate:"required"`
}
