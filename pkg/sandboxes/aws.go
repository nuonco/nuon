package sandboxes

type ECROutputs struct {
	RepositoryURL  string `mapstructure:"repository_url" validate:"required"`
	RepositoryARN  string `mapstructure:"repository_arn" validate:"required"`
	RepositoryName string `mapstructure:"repository_name" validate:"required"`
	RegistryID     string `mapstructure:"registry_id" validate:"required"`
	RegistryURL    string `mapstructure:"repository_url" validate:"required"`
}

type AWSAccountOutputs struct {
	ID     string `mapstructure:"id" validate:"required"`
	Region string `mapstructure:"region" validate:"required"`
}

type VPCOutputs struct {
	Name                    string        `mapstructure:"name" validate:"required"`
	ID                      string        `mapstructure:"id" validate:"required"`
	CIDR                    string        `mapstructure:"cidr" validate:"required"`
	AZs                     []interface{} `mapstructure:"azs" validate:"required" faker:"stringSliceAsInt"`
	PrivateSubnetCidrBlocks []interface{} `mapstructure:"private_subnet_cidr_blocks" validate:"required" faker:"stringSliceAsInt"`
	PrivateSubnetIDs        []interface{} `mapstructure:"private_subnet_ids" validate:"required" faker:"stringSliceAsInt"`
	PublicSubnetIDs         []interface{} `mapstructure:"public_subnet_ids" validate:"required" faker:"stringSliceAsInt"`
	PublicSubnetCidrBlocks  []interface{} `mapstructure:"public_subnet_cidr_blocks" validate:"required" faker:"stringSliceAsInt"`
	DefaultSecurityGroupID  string        `mapstructure:"default_security_group_id" validate:"required"`
}

type DomainOutputs struct {
	Nameservers []interface{} `mapstructure:"nameservers" faker:"stringSliceAsInt"`
	Name        string        `mapstructure:"name" faker:"domain"`

	ZoneID string `mapstructure:"zone_id"`
	ID     string `mapstructure:"id"`
}
