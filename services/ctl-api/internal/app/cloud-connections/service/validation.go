package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var accountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)
var azureIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func validateConnection(connection *app.CloudConnection) error {
	if strings.TrimSpace(connection.Name) == "" {
		return fmt.Errorf("name is required")
	}
	switch connection.Platform {
	case app.CloudPlatformAWS:
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
	case app.CloudPlatformAzure:
		if !azureIDPattern.MatchString(connection.TargetID) {
			return fmt.Errorf("target_id must be an Azure subscription ID")
		}
		if !azureIDPattern.MatchString(connection.TenantID) {
			return fmt.Errorf("tenant_id must be an Entra tenant ID")
		}
		if !azureIDPattern.MatchString(connection.Principal) {
			return fmt.Errorf("principal must be an Entra application client ID")
		}
	default:
		return fmt.Errorf("unsupported cloud platform %q", connection.Platform)
	}
	seen := make(map[app.CloudConnectionCapability]struct{}, len(connection.Capabilities))
	for _, capability := range connection.Capabilities {
		if capability != app.CloudConnectionCapabilityStacks && capability != app.CloudConnectionCapabilityImages {
			return fmt.Errorf("unsupported capability %q", capability)
		}
		if _, ok := seen[capability]; ok {
			return fmt.Errorf("capability %q is duplicated", capability)
		}
		seen[capability] = struct{}{}
	}
	if len(connection.Capabilities) == 0 {
		return fmt.Errorf("at least one capability is required")
	}
	return nil
}
