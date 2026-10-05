package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var accountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)

func validateConnection(connection *app.CloudConnection) error {
	if strings.TrimSpace(connection.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if connection.Platform != app.CloudPlatformAWS {
		return fmt.Errorf("only aws cloud connections are supported")
	}
	if connection.Preset != app.CloudConnectionPresetStacks && connection.Preset != app.CloudConnectionPresetCustom {
		return fmt.Errorf("preset must be stacks or custom")
	}
	if !accountIDPattern.MatchString(connection.TargetID) {
		return fmt.Errorf("target_id must be exactly 12 digits for AWS")
	}
	parsed, err := arn.Parse(connection.Principal)
	if err != nil || parsed.Partition != "aws" || parsed.Service != "iam" || parsed.Region != "" || parsed.AccountID != connection.TargetID || !strings.HasPrefix(parsed.Resource, "role/") || strings.Trim(strings.TrimPrefix(parsed.Resource, "role/"), "/") == "" || strings.ContainsAny(parsed.Resource, "*?") {
		return fmt.Errorf("principal must be an IAM role ARN in account %s", connection.TargetID)
	}
	if connection.DefaultRegion == "" {
		connection.DefaultRegion = "us-east-1"
	}
	return nil
}
