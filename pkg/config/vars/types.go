package vars

type instanceImageRepository struct {
	ID          string `json:"id,omitempty"`
	ARN         string `json:"arn,omitempty"`
	Name        string `json:"name,omitempty"`
	URI         string `json:"uri,omitempty"`
	Image       string `json:"image,omitempty"`
	LoginServer string `json:"login_server,omitempty"`
}

type instanceImageRegistry struct {
	ID string `json:"id"`
}

type instanceImage struct {
	Tag        string                  `json:"tag"`
	Repository instanceImageRepository `json:"repository"`
	Registry   instanceImageRegistry   `json:"registry"`
}

type instanceIntermediate struct {
	Image   instanceImage          `json:"image"`
	Outputs map[string]interface{} `json:"outputs" faker:"-"`
}

type appIntermediate struct {
	ID      string            `json:"id"`
	Secrets map[string]string `json:"secrets" faker:"-"`
}

type orgIntermediate struct {
	ID string `json:"id"`
}

type installSandboxIntermediate struct {
	Type    string                 `json:"type"`
	Version string                 `json:"version"`
	Outputs map[string]interface{} `json:"outputs" faker:"-"`
}

type installIntermediate struct {
	ID string `json:"id"`

	PublicDomain   string                     `json:"public_domain"`
	InternalDomain string                     `json:"internal_domain"`
	Sandbox        installSandboxIntermediate `json:"sandbox"`
	Inputs         map[string]interface{}     `json:"inputs" faker:"-"`
}

type imageIntermediate struct {
	Tag        string `json:"tag"`
	Repository string `json:"repository"`
}

type componentIntermediate struct {
	Outputs map[string]string `json:"outputs"`
	Image   imageIntermediate `json:"image"`
}

type installStackIntermediate struct {
	AccountID             string            `json:"account_id" mapstructure:"account_id"`
	Region                string            `json:"region" mapstructure:"region"`
	VPCID                 string            `json:"vpc_id" mapstructure:"vpc_id"`
	RunnerSubnet          string            `json:"runner_subnet" mapstructure:"runner_subnet"`
	PublicSubnets         []string          `json:"public_subnets" mapstructure:"public_subnets"`
	PrivateSubnets        []string          `json:"private_subnets" mapstructure:"private_subnets"`
	ProvisionIAMRoleARN   string            `json:"provision_iam_role_arn" mapstructure:"provision_iam_role_arn"`
	DeprovisionIAMRoleARN string            `json:"deprovision_iam_role_arn" mapstructure:"deprovision_iam_role_arn"`
	MaintenanceIAMRoleARN string            `json:"maintenance_iam_role_arn" mapstructure:"maintenance_iam_role_arn"`
	RunnerIAMRoleARN      string            `json:"runner_iam_role_arn" mapstructure:"runner_iam_role_arn"`
	BreakGlassRoles       map[string]string `json:"break_glass_roles" mapstructure:"break_glass_roles"`
}

// intermediate represents the intermediate data available to users to interpolate
type intermediate struct {
	Org          orgIntermediate                  `json:"org"`
	App          appIntermediate                  `json:"app"`
	Install      installIntermediate              `json:"install"`
	InstallStack installStackIntermediate         `json:"install_stack"`
	Components   map[string]*instanceIntermediate `json:"components" faker:"-"`
	Sandbox      installSandboxIntermediate       `json:"sandbox"`
}
