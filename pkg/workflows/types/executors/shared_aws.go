package executors

import (
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type AWSSettings struct {
	IAMRoleARN string `json:"iam_role_arn"`
	Region     string `json:"region"`

	AWSRoleDelegationSettings *AWSRoleDelegationSettings `json:"aws_role_delegation_settings"`
}

type AWSRoleDelegationSettings struct {
	IAMRoleARN string `json:"iam_role_arn"`

	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

func (a *AWSSettings) Validate() error {
	availableRegions := []string{
		"us-east-2",
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"af-south-1",
		"ap-east-1",
		"ap-south-2",
		"ap-southeast-3",
		"ap-southeast-4",
		"ap-south-1",
		"us-gov-west-1",
		"us-gov-east-1",
		"ap-northeast-3",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ca-central-1",
		"ca-west-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-south-1",
		"eu-west-3",
		"eu-south-2",
		"eu-north-1",
		"eu-central-2",
		"il-central-1",
		"me-south-1",
		"me-central-1",
		"sa-east-1",
	}

	if !generics.SliceContains(a.Region, availableRegions) {
		return fmt.Errorf("unsupported region %s must be one of %s", a.Region, strings.Join(availableRegions, ", "))
	}

	return nil
}
