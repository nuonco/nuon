package awseks

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"
)

// NOTE: structpb does not support []string type, so we have to use interface{} here
func ToStringSlice(vals []interface{}) []string {
	strVals := make([]string, len(vals))
	for idx, val := range vals {
		v := val
		strVals[idx] = v.(string)
	}

	return strVals
}

type ECSClusterOutputs struct {
	ARN  string `mapstructure:"arn"`
	Name string `mapstructure:"name"`
	ID   string `mapstructure:"id"`
}

type ClusterOutputs struct {
	ARN                      string `mapstructure:"arn"`
	CertificateAuthorityData string `mapstructure:"certificate_authority_data"`
	Endpoint                 string `mapstructure:"endpoint"`
	Name                     string `mapstructure:"name"`
	PlatformVersion          string `mapstructure:"platform_version"`
	Status                   string `mapstructure:"status"`
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

type AccountOutputs struct {
	ID     string `mapstructure:"id" validate:"required"`
	Region string `mapstructure:"region" validate:"required"`
}

type ECROutputs struct {
	RepositoryURL  string `mapstructure:"repository_url" validate:"required"`
	RepositoryARN  string `mapstructure:"repository_arn" validate:"required"`
	RepositoryName string `mapstructure:"repository_name" validate:"required"`
	RegistryID     string `mapstructure:"registry_id" validate:"required"`
	RegistryURL    string `mapstructure:"repository_url" validate:"required"`
}

type RunnerOutputs struct {
	// eks outputs
	DefaultIAMRoleARN string `mapstructure:"default_iam_role_arn"`

	// ecs runner outputs
	Type              string `mapstructure:"type"`
	RunnerIAMRoleARN  string `mapstructure:"runner_iam_role_arn"`
	ODRIAMRoleARN     string `mapstructure:"odr_iam_role_arn"`
	InstallIAMRoleARN string `mapstructure:"install_iam_role_arn"`
}

type DomainOutputs struct {
	Nameservers []interface{} `mapstructure:"nameservers" validate:"required" faker:"stringSliceAsInt"`
	Name        string        `mapstructure:"name" validate:"required" faker:"domain"`
	ZoneID      string        `mapstructure:"zone_id" validate:"required"`
}

type TerraformOutputs struct {
	// domain outputs
	PublicDomain   DomainOutputs `mapstructure:"public_domain"`
	InternalDomain DomainOutputs `mapstructure:"internal_domain"`
	// TODO(jm): rename this to EKSCluster
	Cluster    ClusterOutputs    `mapstructure:"cluster"`
	ECSCluster ECSClusterOutputs `mapstructure:"ecs_cluster"`
	ECR        ECROutputs        `mapstructure:"ecr"`
	VPC        VPCOutputs        `mapstructure:"vpc"`
	Runner     RunnerOutputs     `mapstructure:"runner"`
}

func (t *TerraformOutputs) Validate() error {
	validate := validator.New()
	return validate.Struct(t)
}

func ParseTerraformOutputs(outputs *structpb.Struct) (TerraformOutputs, error) {
	m := outputs.AsMap()

	var tfOutputs TerraformOutputs
	if err := mapstructure.Decode(m, &tfOutputs); err != nil {
		return tfOutputs, fmt.Errorf("invalid terraform outputs: %w", err)
	}

	err := tfOutputs.Validate()
	if err != nil {
		return tfOutputs, fmt.Errorf("terraform output error: %w", err)
	}

	return tfOutputs, nil
}
