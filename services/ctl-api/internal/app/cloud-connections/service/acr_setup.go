package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func (s *service) acrSetup(connection *app.CloudConnection, existingPrincipal bool, repositories []string) SetupResponse {
	response := SetupResponse{Subject: subject(connection), Audience: cloudconnections.ACRAudience, Preset: connection.Preset}
	if s.issuer == nil {
		return response
	}
	response.IssuerURL = s.issuer.Issuer()
	credentialName := "nuon-" + connection.ID
	response.TrustPolicy = map[string]any{
		"name": credentialName, "issuer": response.IssuerURL,
		"subject": response.Subject, "audiences": []string{response.Audience},
	}
	trustJSON, _ := json.MarshalIndent(response.TrustPolicy, "", "  ")
	preflight := fmt.Sprintf(`set -eu
command -v jq >/dev/null || { printf 'Install jq before setting up this connection.\n' >&2; exit 1; }
subscription=%[1]s
tenant=%[2]s
client_id=%[3]s
registry_id=%[4]s
login_server=%[5]s
az account set --subscription "$subscription"
actual_tenant=$(az account show --query tenantId --output tsv | tr '[:upper:]' '[:lower:]')
if [ "$actual_tenant" != "$tenant" ]; then
  printf 'The selected subscription must belong to the configured Entra tenant.\n' >&2
  exit 1
fi
service_principal_object_id=$(az ad sp show --id "$client_id" --query id --output tsv)
test -n "$service_principal_object_id"
actual_login_server=$(az resource show --ids "$registry_id" --api-version 2025-11-01 --subscription "$subscription" --query properties.loginServer --output tsv)
if [ "$actual_login_server" != "$login_server" ]; then
  printf 'The registry ARM ID does not match the configured login server. Review step 3.\n' >&2
  exit 1
fi
permission_mode=$(az resource show --ids "$registry_id" --api-version 2025-11-01 --subscription "$subscription" --query properties.roleAssignmentMode --output tsv)
if [ "$permission_mode" != 'AbacRepositoryPermissions' ]; then
  printf 'The registry must already use RBAC Registry + ABAC Repository Permissions. Review step 3. Setup will not change this mode: switching can break existing AcrPull clients.\n' >&2
  exit 1
fi
`, acrShellQuote(connection.TargetID), acrShellQuote(strings.ToLower(connection.AzureTenantID)), acrShellQuote(connection.Principal), acrShellQuote(connection.AzureRegistryID), acrShellQuote(connection.AzureRegistryLoginServer))
	response.CLI = preflight + fmt.Sprintf(`application_object_id=$(az ad app show --id "$client_id" --query id --output tsv)
test -n "$application_object_id"
printf 'The Nuon issuer discovery and JWKS endpoints must be publicly reachable by Microsoft Entra. Azure does not accept uploaded JWKS.\n'
curl --fail --silent --show-error %[1]s --output /dev/null
curl --fail --silent --show-error %[2]s --output /dev/null
credential=$(cat <<'NUON_FEDERATED_CREDENTIAL'
%[3]s
NUON_FEDERATED_CREDENTIAL
)
credentials=$(az ad app federated-credential list --id "$application_object_id" --output json)
existing=$(printf '%%s' "$credentials" | jq -ce --argjson credential "$credential" 'if type == "array" then [.[] | select(.name == $credential.name)] else error("Invalid federated credential list") end')
if [ "$(printf '%%s' "$existing" | jq 'length')" = 0 ]; then
  az ad app federated-credential create --id "$application_object_id" --parameters "$credential"
else
  printf '%%s' "$existing" | jq -e --argjson credential "$credential" 'length == 1 and .[0].issuer == $credential.issuer and .[0].subject == $credential.subject and .[0].audiences == $credential.audiences' >/dev/null
fi
`, acrShellQuote(response.IssuerURL+"/.well-known/openid-configuration"), acrShellQuote(response.IssuerURL+oidcissuer.JWKSPath), trustJSON)
	response.Terraform = fmt.Sprintf(`terraform {
  required_version = ">= 1.5.0"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 3.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
    azapi = {
      source  = "Azure/azapi"
      version = "~> 2.0"
    }
  }
}

provider "azuread" {
  tenant_id = %[1]q
}

provider "azurerm" {
  subscription_id = %[2]q
  tenant_id       = %[1]q
  features {}
}

provider "azapi" {
  subscription_id = %[2]q
  tenant_id       = %[1]q
}

data "azuread_application" "nuon" {
  client_id = %[3]q
}

data "azuread_service_principal" "nuon" {
  client_id = %[3]q
}

data "azapi_resource" "nuon_registry" {
  type                   = "Microsoft.ContainerRegistry/registries@2025-11-01"
  resource_id            = %[4]q
  response_export_values = ["properties.loginServer", "properties.roleAssignmentMode"]

  lifecycle {
    postcondition {
      condition     = self.output.properties.loginServer == %[5]q
      error_message = "The registry ARM ID must match the configured login server. Review step 3."
    }
    postcondition {
      condition     = try(self.output.properties.roleAssignmentMode, "") == "AbacRepositoryPermissions"
      error_message = "The registry must already use RBAC Registry + ABAC Repository Permissions. Setup will not change mode because doing so can break existing AcrPull clients. Review step 3."
    }
  }
}

resource "azuread_application_federated_identity_credential" "nuon" {
  application_id = data.azuread_application.nuon.id
  display_name   = %[6]q
  audiences      = [%[7]q]
  issuer         = %[8]q
  subject        = %[9]q
}
`, connection.AzureTenantID, connection.TargetID, connection.Principal, connection.AzureRegistryID, connection.AzureRegistryLoginServer, credentialName, response.Audience, response.IssuerURL, response.Subject)
	if len(repositories) == 0 {
		return response
	}
	matches := make([]string, 0, len(repositories))
	for _, repository := range repositories {
		matches = append(matches, "@Request[Microsoft.ContainerRegistry/registries/repositories:name] StringEqualsIgnoreCase '"+repository+"'")
	}
	condition := "((!(ActionMatches{'Microsoft.ContainerRegistry/registries/repositories/content/read'}) AND !(ActionMatches{'Microsoft.ContainerRegistry/registries/repositories/metadata/read'})) OR (" + strings.Join(matches, " OR ") + "))"
	response.PermissionsPolicy = map[string]any{
		"scope": connection.AzureRegistryID, "role": "Container Registry Repository Reader",
		"condition": condition, "condition_version": "2.0",
	}
	permissionsCLI := fmt.Sprintf(`condition=$(cat <<'NUON_ABAC_CONDITION'
%[1]s
NUON_ABAC_CONDITION
)
assignments=$(az role assignment list --subscription "$subscription" --scope "$registry_id" --assignee-object-id "$service_principal_object_id" --role 'Container Registry Repository Reader' --fill-principal-name false --fill-role-definition-name false --output json)
assignments=$(printf '%%s\n' "$assignments" | jq -e --arg scope "$registry_id" --arg principal "$service_principal_object_id" 'if type == "array" then [.[] | select((.scope | ascii_downcase) == ($scope | ascii_downcase) and (.principalId | ascii_downcase) == ($principal | ascii_downcase))] else error("Invalid role assignment list") end')
assignment_count=$(printf '%%s\n' "$assignments" | jq 'length')
case "$assignment_count" in
  0)
    az role assignment create --subscription "$subscription" --assignee-object-id "$service_principal_object_id" --assignee-principal-type ServicePrincipal --role 'Container Registry Repository Reader' --scope "$registry_id" --condition "$condition" --condition-version '2.0'
    ;;
  1)
    assignment=$(printf '%%s\n' "$assignments" | jq -c '.[0]')
    existing_condition=$(printf '%%s\n' "$assignment" | jq -r '.condition // ""')
    if [ -z "$existing_condition" ]; then
      printf 'The existing Repository Reader assignment is unconditional and already grants access; no permissions changed.\n'
    elif [ "$(printf '%%s\n' "$assignment" | jq -r '.conditionVersion')" != '2.0' ]; then
      printf 'The existing assignment has an unsupported condition version. Review it manually; no permissions changed.\n' >&2
      exit 1
    elif printf '%%s\n' "$assignment" | jq -e --arg condition "$condition" '.condition == $condition or (.condition | contains("(" + $condition + ")"))' >/dev/null; then
      printf 'The requested repository grant is already present; no permissions changed.\n'
    else
      assignment=$(printf '%%s\n' "$assignment" | jq -c --arg condition "$condition" '.condition = "(" + .condition + ") OR (" + $condition + ")"')
      az role assignment update --subscription "$subscription" --role-assignment "$assignment"
    fi
    ;;
  *)
    printf 'Multiple exact-scope Repository Reader assignments match this identity. Review them manually; no permissions changed.\n' >&2
    exit 1
    ;;
esac
`, condition)
	response.PermissionsCLI = preflight + permissionsCLI
	response.CLI += permissionsCLI
	if existingPrincipal {
		response.Terraform += fmt.Sprintf(`
output "nuon_repository_condition_to_merge" {
  description = "Merge this exact repository condition with OR into the Repository Reader assignment in its original owning state, preserving existing grants, or use the additive permissions CLI. Azure does not support a second assignment for the same principal, role, and scope in independent state. This configuration only manages this connection's federation, not the shared role assignment."
  value = %q
}
`, condition)
		return response
	}
	response.Terraform += fmt.Sprintf(`
resource "azurerm_role_assignment" "nuon" {
  scope                = data.azapi_resource.nuon_registry.id
  role_definition_name = "Container Registry Repository Reader"
  principal_id         = data.azuread_service_principal.nuon.object_id
  principal_type       = "ServicePrincipal"
  condition_version    = "2.0"
  condition            = %q
  depends_on           = [azuread_application_federated_identity_credential.nuon]
}
`, condition)
	return response
}

func acrShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
