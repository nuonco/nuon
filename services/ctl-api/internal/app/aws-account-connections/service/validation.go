package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

var accountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)

var supportedRegions = map[string]struct{}{
	"us-east-1": {}, "us-east-2": {}, "us-west-1": {}, "us-west-2": {},
	"eu-central-1": {}, "eu-west-1": {}, "eu-west-2": {}, "eu-west-3": {}, "eu-north-1": {}, "eu-south-1": {}, "eu-south-2": {},
	"ap-east-1": {}, "ap-south-1": {}, "ap-south-2": {}, "ap-northeast-1": {}, "ap-northeast-2": {}, "ap-northeast-3": {}, "ap-southeast-1": {}, "ap-southeast-2": {}, "ap-southeast-3": {}, "ap-southeast-4": {},
	"ca-central-1": {}, "ca-west-1": {}, "me-south-1": {}, "me-central-1": {}, "af-south-1": {}, "sa-east-1": {},
}

func validateAccountID(value string) error {
	if !accountIDPattern.MatchString(value) {
		return fmt.Errorf("account_id must be exactly 12 digits")
	}
	return nil
}

func validateRegion(value string) error {
	if _, ok := supportedRegions[value]; !ok {
		return fmt.Errorf("default_region is not a supported AWS commercial region")
	}
	return nil
}

func validateRoleARN(value, accountID string) error {
	parsed, err := arn.Parse(value)
	if err != nil || parsed.Partition != "aws" || parsed.Service != "iam" || parsed.Region != "" || parsed.AccountID != accountID || !strings.HasPrefix(parsed.Resource, "role/") || strings.Trim(strings.TrimPrefix(parsed.Resource, "role/"), "/") == "" || strings.ContainsAny(parsed.Resource, "*?") {
		return fmt.Errorf("role_arn must be an IAM role ARN in account %s", accountID)
	}
	return nil
}

func validateManagementRoleARN(value string) error {
	parsed, err := arn.Parse(value)
	if err != nil || parsed.Partition != "aws" || parsed.Service != "iam" || parsed.Region != "" || !accountIDPattern.MatchString(parsed.AccountID) || !strings.HasPrefix(parsed.Resource, "role/") || strings.Trim(strings.TrimPrefix(parsed.Resource, "role/"), "/") == "" || strings.ContainsAny(parsed.Resource, "*?") {
		return fmt.Errorf("management IAM role ARN must be an exact IAM role ARN")
	}
	return nil
}
