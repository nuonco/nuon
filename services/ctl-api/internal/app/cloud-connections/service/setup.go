package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const azureTokenExchangeAudience = "api://AzureADTokenExchange"

type SetupOptions struct {
	Repositories []string
	Registry     string
}

type SetupResponse struct {
	IssuerURL      string                          `json:"issuer_url"`
	Subject        string                          `json:"subject"`
	Audience       string                          `json:"audience"`
	TrustPolicy    map[string]any                  `json:"trust_policy"`
	Terraform      string                          `json:"terraform"`
	CLI            string                          `json:"cli"`
	PortalJSON     string                          `json:"portal_json,omitempty"`
	CloudFormation string                          `json:"cloudformation"`
	Capabilities   []app.CloudConnectionCapability `json:"capabilities"`
	Repositories   []string                        `json:"repositories,omitempty"`
	Registry       string                          `json:"registry,omitempty"`
}

func subject(connection *app.CloudConnection) string {
	return fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID)
}

func (s *service) setup(connection *app.CloudConnection, options SetupOptions) SetupResponse {
	switch connection.Platform {
	case app.CloudPlatformAzure:
		return s.azureSetup(connection, options)
	case app.CloudPlatformGCP:
		return s.gcpSetup(connection, options)
	default:
		return s.awsSetup(connection, options)
	}
}

func (s *service) awsSetup(connection *app.CloudConnection, options SetupOptions) SetupResponse {
	response := SetupResponse{Subject: subject(connection), Audience: "sts.amazonaws.com", Capabilities: connection.RequestedCapabilities, Repositories: options.Repositories}
	if s.issuer == nil {
		return response
	}
	issuerURL := s.issuer.Issuer()
	issuer, _ := url.Parse(issuerURL)
	conditionPrefix := strings.TrimPrefix(issuer.Host+issuer.Path, "/")
	trustPolicy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []any{map[string]any{
			"Effect":    "Allow",
			"Principal": map[string]string{"Federated": fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", connection.TargetID, conditionPrefix)},
			"Action":    "sts:AssumeRoleWithWebIdentity",
			"Condition": map[string]any{"StringEquals": map[string]string{
				conditionPrefix + ":aud": "sts.amazonaws.com",
				conditionPrefix + ":sub": subject(connection),
			}},
		}},
	}
	policyJSON, _ := json.Marshal(trustPolicy)
	capabilitiesJSON, _ := json.Marshal(connection.RequestedCapabilities)
	repositoriesJSON, _ := json.Marshal(options.Repositories)
	response.IssuerURL = issuerURL
	response.TrustPolicy = trustPolicy
	response.Terraform = fmt.Sprintf(`module "nuon_cloud_connection" {
  source  = "nuonco/ecr-access/aws"
  nuon_issuer = %q
  nuon_subject = %q
  capabilities = %s
  repositories = %s
}`, issuerURL, subject(connection), capabilitiesJSON, repositoriesJSON)
	response.CLI = fmt.Sprintf("aws iam create-open-id-connect-provider --url %s --client-id-list sts.amazonaws.com --thumbprint-list <thumbprint>\naws iam create-role --role-name <role-name> --assume-role-policy-document '%s'", issuerURL, policyJSON)
	response.CloudFormation = fmt.Sprintf("NuonConnectionRole:\n  Type: AWS::IAM::Role\n  Properties:\n    AssumeRolePolicyDocument: %s\n", string(policyJSON))
	return response
}

func (s *service) azureSetup(connection *app.CloudConnection, options SetupOptions) SetupResponse {
	if options.Repositories == nil {
		options.Repositories = []string{}
	}
	response := SetupResponse{Subject: subject(connection), Audience: azureTokenExchangeAudience, Capabilities: connection.RequestedCapabilities, Repositories: options.Repositories, Registry: options.Registry}
	if s.issuer == nil {
		return response
	}
	issuerURL := s.issuer.Issuer()
	federatedCredential := map[string]any{
		"name":      "nuon-cloud-connection",
		"issuer":    issuerURL,
		"subject":   subject(connection),
		"audiences": []string{azureTokenExchangeAudience},
	}
	portalJSON, _ := json.MarshalIndent(federatedCredential, "", "  ")
	capabilitiesJSON, _ := json.Marshal(connection.RequestedCapabilities)
	repositoriesJSON, _ := json.Marshal(options.Repositories)
	response.IssuerURL = issuerURL
	response.TrustPolicy = federatedCredential
	response.PortalJSON = string(portalJSON)
	response.Terraform = fmt.Sprintf(`module "nuon_cloud_connection" {
  source          = "nuonco/acr-access/azure"
  nuon_issuer     = %q
  nuon_subject    = %q
  capabilities    = %s
  repositories    = %s
  subscription_id = %q
  registry        = %q
}`, issuerURL, subject(connection), capabilitiesJSON, repositoriesJSON, connection.TargetID, options.Registry)
	lines := []string{
		"APP_ID=$(az ad app create --display-name nuon-cloud-connection --query appId -o tsv)",
		"cat > nuon-federated-credential.json <<'JSON'",
		string(portalJSON),
		"JSON",
		"az ad app federated-credential create --id \"$APP_ID\" --parameters nuon-federated-credential.json",
	}
	if connection.HasRequestedCapability(app.CloudConnectionCapabilityStacks) {
		lines = append(lines, fmt.Sprintf("az role assignment create --assignee \"$APP_ID\" --role Owner --scope /subscriptions/%s", connection.TargetID))
	}
	if connection.HasRequestedCapability(app.CloudConnectionCapabilityImages) {
		lines = append(lines, "REGISTRY_ID=$(az acr show --name "+options.Registry+" --query id -o tsv)", "az role assignment create --assignee \"$APP_ID\" --role AcrPull --scope \"$REGISTRY_ID\"")
	}
	response.CLI = strings.Join(lines, "\n")
	return response
}

func (s *service) gcpSetup(connection *app.CloudConnection, options SetupOptions) SetupResponse {
	if options.Repositories == nil {
		options.Repositories = []string{}
	}
	provider := connection.IdentityProvider
	if provider == "" {
		provider = "projects/<project-number>/locations/global/workloadIdentityPools/<pool-id>/providers/<provider-id>"
	}
	audience := "https://iam.googleapis.com/" + strings.TrimPrefix(strings.TrimPrefix(provider, "https://iam.googleapis.com/"), "//iam.googleapis.com/")
	response := SetupResponse{Subject: subject(connection), Audience: audience, Capabilities: connection.RequestedCapabilities, Repositories: options.Repositories}
	if s.issuer == nil {
		return response
	}
	issuerURL := s.issuer.Issuer()
	serviceAccountID := strings.SplitN(connection.Principal, "@", 2)[0]
	capabilitiesJSON, _ := json.Marshal(connection.RequestedCapabilities)
	repositoriesJSON, _ := json.Marshal(options.Repositories)
	response.IssuerURL = issuerURL
	response.Terraform = fmt.Sprintf(`module "nuon_cloud_connection" {
  source       = "nuonco/gar-access/google"
  nuon_issuer  = %q
  nuon_subject = %q
  capabilities = %s
  repositories = %s
  project_id   = %q
}`, issuerURL, subject(connection), capabilitiesJSON, repositoriesJSON, connection.TargetID)
	lines := []string{
		"PROJECT_ID=" + connection.TargetID,
		"PROJECT_NUMBER=$(gcloud projects describe \"$PROJECT_ID\" --format='value(projectNumber)')",
		"POOL_ID=<pool-id>",
		"PROVIDER_ID=<provider-id>",
		"SERVICE_ACCOUNT_EMAIL=" + connection.Principal,
		"gcloud iam workload-identity-pools create \"$POOL_ID\" --project \"$PROJECT_ID\" --location global --display-name \"Nuon cloud connection\"",
		fmt.Sprintf("gcloud iam workload-identity-pools providers create-oidc \"$PROVIDER_ID\" --project \"$PROJECT_ID\" --location global --workload-identity-pool \"$POOL_ID\" --issuer-uri %q --allowed-audiences %q --attribute-mapping=\"google.subject=assertion.sub\"", issuerURL, audience),
		"gcloud iam service-accounts create " + serviceAccountID + " --project \"$PROJECT_ID\"",
		fmt.Sprintf("gcloud iam service-accounts add-iam-policy-binding \"$SERVICE_ACCOUNT_EMAIL\" --project \"$PROJECT_ID\" --role roles/iam.workloadIdentityUser --member \"principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL_ID/subject/%s\"", subject(connection)),
	}
	if len(options.Repositories) == 0 {
		lines = append(lines, "gcloud projects add-iam-policy-binding \"$PROJECT_ID\" --role roles/artifactregistry.reader --member \"serviceAccount:$SERVICE_ACCOUNT_EMAIL\"")
	} else {
		for _, repository := range options.Repositories {
			parts := strings.SplitN(strings.Trim(repository, "/"), "/", 2)
			location, name := "<location>", repository
			if len(parts) == 2 {
				location, name = parts[0], parts[1]
			}
			lines = append(lines, fmt.Sprintf("gcloud artifacts repositories add-iam-policy-binding %q --project \"$PROJECT_ID\" --location %q --role roles/artifactregistry.reader --member \"serviceAccount:$SERVICE_ACCOUNT_EMAIL\"", name, location))
		}
	}
	lines = append(lines, "gcloud iam workload-identity-pools providers describe \"$PROVIDER_ID\" --project \"$PROJECT_ID\" --location global --workload-identity-pool \"$POOL_ID\" --format='value(name)'", "Paste the provider resource name returned above into identity_provider.")
	response.CLI = strings.Join(lines, "\n")
	return response
}
