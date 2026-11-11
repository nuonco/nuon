package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

func (s *service) garSetup(connection *app.CloudConnection, existingPrincipal bool, repositories []string) SetupResponse {
	response := SetupResponse{Subject: cloudconnections.Subject(connection), Audience: "//iam.googleapis.com/" + cloudconnections.GARProvider(connection), Preset: connection.Preset}
	if s.issuer == nil {
		return response
	}
	response.IssuerURL = s.issuer.Issuer()
	poolID := "nuon-" + connection.ID
	pool := fmt.Sprintf("projects/%s/locations/global/workloadIdentityPools/%s", connection.GCPProjectNumber, poolID)
	member := "principal://iam.googleapis.com/" + pool + "/subject/" + response.Subject
	condition := "assertion.sub == '" + response.Subject + "'"
	response.TrustPolicy = map[string]any{
		"issuer_uri": response.IssuerURL, "allowed_audiences": []string{response.Audience},
		"attribute_mapping":   map[string]string{"google.subject": "assertion.sub"},
		"attribute_condition": condition,
		"service_account":     connection.Principal, "role": "roles/iam.workloadIdentityUser", "member": member,
	}
	bindings := make([]map[string]string, 0, len(repositories))
	for _, repository := range repositories {
		bindings = append(bindings, map[string]string{"resource": repository, "role": "roles/artifactregistry.reader", "member": "serviceAccount:" + connection.Principal})
	}
	if len(bindings) > 0 {
		response.PermissionsPolicy = map[string]any{"bindings": bindings}
	}
	jwks, _ := json.MarshalIndent(s.jwks, "", "  ")
	saName := strings.Split(connection.Principal, "@")[0]
	response.CLI = fmt.Sprintf(`set -eu
command -v jq >/dev/null
project=%[1]s
service_account=%[8]s
pool_id=%[3]s
pool_name=%[10]s
project_number=$(gcloud projects describe "$project" --format='value(projectNumber)')
if [ "$project_number" != %[11]s ]; then
  printf 'The GCP project number must match the configured project.\n' >&2
  exit 1
fi
gcloud services enable iam.googleapis.com iamcredentials.googleapis.com sts.googleapis.com artifactregistry.googleapis.com --project="$project"
accounts=$(gcloud iam service-accounts list --project="$project" --format=json)
account_count=$(printf '%%s' "$accounts" | jq -er --arg email "$service_account" 'if type == "array" then [.[] | select(.email == $email)] | length else error("Invalid service account list") end')
if [ "$account_count" = 0 ]; then
  gcloud iam service-accounts create %[2]s --project="$project"
fi
pools=$(gcloud iam workload-identity-pools list --project="$project" --location=global --format=json)
pool_count=$(printf '%%s' "$pools" | jq -er --arg name "$pool_name" 'if type == "array" then [.[] | select(.name == $name)] | length else error("Invalid pool list") end')
if [ "$pool_count" = 0 ]; then
  gcloud iam workload-identity-pools create "$pool_id" --project="$project" --location=global --display-name="Nuon image repository access"
fi
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT HUP INT TERM
cat > "$workdir/jwks.json" <<'NUON_JWKS'
%[4]s
NUON_JWKS
providers=$(gcloud iam workload-identity-pools providers list --project="$project" --location=global --workload-identity-pool="$pool_id" --format=json)
provider=$(printf '%%s' "$providers" | jq -ce --arg name "$pool_name/providers/nuon" 'if type == "array" then [.[] | select(.name == $name)] else error("Invalid provider list") end')
if [ "$(printf '%%s' "$provider" | jq 'length')" = 0 ]; then
  provider_operation=create-oidc
else
  provider_operation=update-oidc
fi
gcloud iam workload-identity-pools providers "$provider_operation" nuon --project="$project" --location=global --workload-identity-pool="$pool_id" --issuer-uri=%[5]s --allowed-audiences=%[6]s --attribute-mapping="google.subject=assertion.sub" --attribute-condition=%[7]s --jwk-json-path="$workdir/jwks.json"
gcloud iam service-accounts add-iam-policy-binding "$service_account" --project="$project" --role=roles/iam.workloadIdentityUser --member=%[9]s --condition=None
`, shellQuote(connection.TargetID), shellQuote(saName), shellQuote(poolID), jwks, shellQuote(response.IssuerURL), shellQuote(response.Audience), shellQuote(condition), shellQuote(connection.Principal), shellQuote(member), shellQuote(pool), shellQuote(connection.GCPProjectNumber))
	if repositories == nil {
		repositories = []string{}
	}
	repositoriesJSON, _ := json.Marshal(repositories)
	serviceAccountType := "resource"
	serviceAccountReference := "google_service_account.nuon"
	if existingPrincipal {
		serviceAccountType = "data"
		serviceAccountReference = "data.google_service_account.nuon"
	}
	serviceAccount := fmt.Sprintf(`%s "google_service_account" "nuon" {
  project = %q
  account_id = %q
  depends_on = [google_project_service.nuon]
}
`, serviceAccountType, connection.TargetID, saName)
	response.Terraform = fmt.Sprintf(`resource "google_project_service" "nuon" {
  for_each = toset(["iam.googleapis.com", "iamcredentials.googleapis.com", "sts.googleapis.com", "artifactregistry.googleapis.com"])
  project = %[1]q
  service = each.value
  disable_on_destroy = false
}
%[10]s
resource "google_iam_workload_identity_pool" "nuon" {
  project = %[1]q
  workload_identity_pool_id = %[3]q
  depends_on = [google_project_service.nuon]
}
resource "google_iam_workload_identity_pool_provider" "nuon" {
  project = %[1]q
  workload_identity_pool_id = google_iam_workload_identity_pool.nuon.workload_identity_pool_id
  workload_identity_pool_provider_id = "nuon"
  attribute_mapping = { "google.subject" = "assertion.sub" }
  attribute_condition = %[4]q
  oidc {
    issuer_uri = %[5]q
    allowed_audiences = [%[6]q]
    jwks_json = jsonencode(%[7]s)
  }
}
resource "google_service_account_iam_member" "nuon" {
  service_account_id = %[11]s.name
  role = "roles/iam.workloadIdentityUser"
  member = %[8]q
  depends_on = [google_iam_workload_identity_pool_provider.nuon]
}
resource "google_artifact_registry_repository_iam_member" "nuon" {
  for_each = toset(%[9]s)
  project = split("/", each.value)[1]
  location = split("/", each.value)[3]
  repository = split("/", each.value)[5]
  role = "roles/artifactregistry.reader"
  member = "serviceAccount:${%[11]s.email}"
}
`, connection.TargetID, saName, poolID, condition, response.IssuerURL, response.Audience, jwks, member, repositoriesJSON, serviceAccount, serviceAccountReference)
	for _, repository := range repositories {
		parts := strings.Split(repository, "/")
		response.PermissionsCLI += fmt.Sprintf("gcloud artifacts repositories add-iam-policy-binding %s --project=%s --location=%s --role=roles/artifactregistry.reader --member=serviceAccount:%s --condition=None\n", parts[5], connection.TargetID, parts[3], connection.Principal)
	}
	if response.PermissionsCLI != "" {
		response.PermissionsCLI = "set -eu\n" + response.PermissionsCLI
		response.CLI += response.PermissionsCLI
	}
	return response
}
