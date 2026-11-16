package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"

	"github.com/nuonco/nuon/pkg/azure/acr"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var (
	accountIDPattern        = regexp.MustCompile(`^[0-9]{12}$`)
	awsImageRegionPattern   = regexp.MustCompile(`^[a-z]{2}(?:-[a-z]+)+-[0-9]+$`)
	gcpIDPattern            = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)
	gcpProjectNumberPattern = regexp.MustCompile(`^[1-9][0-9]{5,19}$`)
	garLocationPattern      = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	azureUUIDPattern        = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	azureRegistryIDPattern  = regexp.MustCompile(`(?i)^/subscriptions/([0-9a-f-]+)/resourceGroups/([a-z0-9_.()-]{1,90})/providers/Microsoft\.ContainerRegistry/registries/([a-z0-9]{5,50})$`)
)

func validateConnection(connection *app.CloudConnection) error {
	if strings.TrimSpace(connection.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if err := rejectOtherPlatformFields(connection); err != nil {
		return err
	}
	switch connection.Platform {
	case app.CloudPlatformAWS:
		return validateAWS(connection)
	case app.CloudPlatformGCP:
		return validateGCP(connection)
	case app.CloudPlatformAzure:
		return validateAzure(connection)
	}
	return fmt.Errorf("only aws, gcp, and azure cloud connections are supported")
}

func rejectOtherPlatformFields(connection *app.CloudConnection) error {
	if connection.Platform != app.CloudPlatformGCP && connection.GCPProjectNumber != "" {
		return fmt.Errorf("gcp_project_number requires the gcp platform and gar preset")
	}
	if connection.Platform != app.CloudPlatformAzure && (connection.AzureTenantID != "" || connection.AzureRegistryID != "" || connection.AzureRegistryLoginServer != "") {
		return fmt.Errorf("azure_tenant_id, azure_registry_id, and azure_registry_login_server require the azure platform and acr preset")
	}
	return nil
}

func validateAWS(connection *app.CloudConnection) error {
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

func validateGCP(connection *app.CloudConnection) error {
	if connection.Preset != app.CloudConnectionPresetGAR {
		return fmt.Errorf("only aws cloud connections support stacks, ecr, and custom; GCP requires the gar preset")
	}
	if !gcpIDPattern.MatchString(connection.TargetID) {
		return fmt.Errorf("target_id must be a GCP project ID")
	}
	if !gcpProjectNumberPattern.MatchString(connection.GCPProjectNumber) {
		return fmt.Errorf("gcp_project_number must be the numeric GCP project number")
	}
	suffix := "@" + connection.TargetID + ".iam.gserviceaccount.com"
	if !strings.HasSuffix(connection.Principal, suffix) || !gcpIDPattern.MatchString(strings.TrimSuffix(connection.Principal, suffix)) {
		return fmt.Errorf("principal must be a service account email in project %s", connection.TargetID)
	}
	if !garLocationPattern.MatchString(connection.DefaultRegion) {
		return fmt.Errorf("default_region must be a GAR region or multi-region location")
	}
	return nil
}

func validateAzure(connection *app.CloudConnection) error {
	if connection.Preset != app.CloudConnectionPresetACR {
		return fmt.Errorf("only aws cloud connections support stacks, ecr, and custom; Azure requires the acr preset")
	}
	if !azureUUIDPattern.MatchString(connection.TargetID) {
		return fmt.Errorf("target_id must be an Azure subscription UUID")
	}
	if !azureUUIDPattern.MatchString(connection.Principal) {
		return fmt.Errorf("principal must be the client ID UUID of an existing Microsoft Entra application with an existing service principal")
	}
	if !azureUUIDPattern.MatchString(connection.AzureTenantID) {
		return fmt.Errorf("azure_tenant_id must be a Microsoft Entra tenant UUID")
	}
	parts := azureRegistryIDPattern.FindStringSubmatch(connection.AzureRegistryID)
	if parts == nil || !strings.EqualFold(parts[1], connection.TargetID) || strings.HasSuffix(parts[2], ".") {
		return fmt.Errorf("azure_registry_id must be an exact Azure Container Registry ARM ID in subscription %s", connection.TargetID)
	}
	if err := acr.ValidateLoginServer(connection.AzureRegistryLoginServer); err != nil {
		return err
	}
	if connection.AzureRegistryLoginServer != strings.ToLower(connection.AzureRegistryLoginServer) {
		return fmt.Errorf("azure_registry_login_server must be a lowercase bare <registry>.azurecr.io hostname")
	}
	return nil
}
