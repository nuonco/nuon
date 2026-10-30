package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

var accountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)
var awsImageRegionPattern = regexp.MustCompile(`^[a-z]{2}(?:-[a-z]+)+-[0-9]+$`)
var garLocationPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func validateConnection(connection *app.CloudConnection) error {
	if strings.TrimSpace(connection.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if connection.Platform == app.CloudPlatformAzure {
		return cloudconnections.ValidateACRConnection(connection)
	}
	if connection.AzureTenantID != "" || connection.AzureRegistryID != "" || connection.AzureRegistryLoginServer != "" {
		return fmt.Errorf("azure_tenant_id, azure_registry_id, and azure_registry_login_server require the azure platform and acr preset")
	}
	if connection.Platform == app.CloudPlatformGCP {
		return validateGARConnection(connection)
	}
	if connection.Platform != app.CloudPlatformAWS {
		return fmt.Errorf("only aws, gcp, and azure cloud connections are supported")
	}
	if connection.GCPProjectNumber != "" {
		return fmt.Errorf("gcp_project_number requires the gcp platform and gar preset")
	}
	if connection.Preset != app.CloudConnectionPresetStacks && connection.Preset != app.CloudConnectionPresetECR && connection.Preset != app.CloudConnectionPresetCustom {
		return fmt.Errorf("preset must be stacks, ecr, or custom")
	}
	if !accountIDPattern.MatchString(connection.TargetID) {
		return fmt.Errorf("target_id must be exactly 12 digits for AWS")
	}
	parsed, err := arn.Parse(connection.Principal)
	if err != nil || parsed.Partition != "aws" || parsed.Service != "iam" || parsed.Region != "" || parsed.AccountID != connection.TargetID || !strings.HasPrefix(parsed.Resource, "role/") || strings.Trim(strings.TrimPrefix(parsed.Resource, "role/"), "/") == "" || strings.ContainsAny(parsed.Resource, "*?") {
		return fmt.Errorf("principal must be an IAM role ARN in account %s", connection.TargetID)
	}
	if connection.DefaultRegion == "" {
		if connection.Preset == app.CloudConnectionPresetECR {
			return fmt.Errorf("default_region is required for ECR image connection matching")
		}
		connection.DefaultRegion = "us-east-1"
	}
	if connection.Preset == app.CloudConnectionPresetECR && (!awsImageRegionPattern.MatchString(connection.DefaultRegion) || strings.HasPrefix(connection.DefaultRegion, "us-gov-") || strings.HasPrefix(connection.DefaultRegion, "cn-")) {
		return fmt.Errorf("default_region must be a supported commercial AWS region for image access")
	}
	return nil
}

func validateGARConnection(connection *app.CloudConnection) error {
	if connection.Preset != app.CloudConnectionPresetGAR {
		return fmt.Errorf("only aws cloud connections support stacks, ecr, and custom; GCP requires the gar preset")
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`).MatchString(connection.TargetID) {
		return fmt.Errorf("target_id must be a GCP project ID")
	}
	if !regexp.MustCompile(`^[1-9][0-9]{5,19}$`).MatchString(connection.GCPProjectNumber) {
		return fmt.Errorf("gcp_project_number must be the numeric GCP project number")
	}
	suffix := "@" + connection.TargetID + ".iam.gserviceaccount.com"
	if !strings.HasSuffix(connection.Principal, suffix) || !regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`).MatchString(strings.TrimSuffix(connection.Principal, suffix)) {
		return fmt.Errorf("principal must be a service account email in project %s", connection.TargetID)
	}
	if !garLocationPattern.MatchString(connection.DefaultRegion) {
		return fmt.Errorf("default_region must be a GAR region or multi-region location")
	}
	return nil
}
