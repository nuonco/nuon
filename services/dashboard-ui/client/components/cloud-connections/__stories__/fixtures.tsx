import type { ReactNode } from 'react'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import type { TCloudConnection } from '@/types'

export const connection: TCloudConnection = {
  id: 'cc-example',
  org_id: 'org-mock-001',
  name: 'acme-production',
  platform: 'aws',
  target_id: '123456789012',
  principal: 'arn:aws:iam::123456789012:role/acme-nuon',
  created_at: '2026-09-28T07:00:00Z',
  updated_at: '2026-09-28T07:00:00Z',
  status: 'pending',
  verification_in_progress: false,
  preset: 'stacks',
  used_by: { installs: 0 },
  setup: {
    issuer_url: 'https://api.example.com',
    subject: 'org:org-mock-001:connection:cc-example',
    audience: 'sts.amazonaws.com',
    preset: 'stacks',
    trust_policy: {
      Version: '2012-10-17',
      Statement: [
        {
          Effect: 'Allow',
          Action: 'sts:AssumeRoleWithWebIdentity',
          Principal: {
            Federated:
              'arn:aws:iam::123456789012:oidc-provider/api.example.com',
          },
          Condition: {
            StringEquals: {
              'api.example.com:aud': 'sts.amazonaws.com',
              'api.example.com:sub': 'org:org-mock-001:connection:cc-example',
            },
          },
        },
      ],
    },
    permissions_policy: {
      Version: '2012-10-17',
      Statement: [
        {
          Effect: 'Allow',
          Action: ['cloudformation:DescribeStacks'],
          Resource: '*',
        },
      ],
    },
    cli: 'aws iam create-role --role-name <role-name> --assume-role-policy-document file://trust.json',
    terraform:
      'resource "aws_iam_role" "nuon" {\n  name = "<role-name>"\n  assume_role_policy = file("trust.json")\n}',
    cloudformation:
      'NuonRole:\n  Type: AWS::IAM::Role\n  Properties:\n    RoleName: <role-name>',
  },
}
export const customConnection: TCloudConnection = {
  ...connection,
  name: 'acme-custom',
  preset: 'custom',
  setup: {
    ...connection.setup,
    preset: 'custom',
    permissions_policy: undefined,
  },
}
const ecrRepositoryARNs = [
  'arn:aws:ecr:us-west-2:123456789012:repository/acme/api',
  'arn:aws:ecr:us-west-2:123456789012:repository/acme/worker',
]
const ecrPermissionsPolicy = {
  Version: '2012-10-17',
  Statement: [
    {
      Effect: 'Allow',
      Action: ['ecr:GetAuthorizationToken'],
      Resource: '*',
    },
    {
      Effect: 'Allow',
      Action: [
        'ecr:DescribeRepositories',
        'ecr:ListImages',
        'ecr:BatchGetImage',
        'ecr:GetDownloadUrlForLayer',
        'ecr:BatchCheckLayerAvailability',
      ],
      Resource: ecrRepositoryARNs,
    },
  ],
}
export const initialECRConnection: TCloudConnection = {
  ...connection,
  name: 'acme-images',
  preset: 'ecr',
  default_region: 'us-west-2',
  setup: {
    ...connection.setup,
    preset: 'ecr',
    permissions_policy: ecrPermissionsPolicy,
    cli: `${connection.setup.cli}
aws iam put-role-policy --role-name <role-name> --policy-name nuon-ecr-access --policy-document '${JSON.stringify(ecrPermissionsPolicy)}'`,
    terraform: `${connection.setup.terraform}
resource "aws_iam_role_policy" "nuon_ecr" {
  name = "nuon-ecr-access"
  role = aws_iam_role.nuon.id
  policy = jsonencode(${JSON.stringify(ecrPermissionsPolicy, null, 2)})
}`,
    cloudformation: `${connection.setup.cloudformation}
    Policies:
      - PolicyName: nuon-ecr-access
        PolicyDocument: ${JSON.stringify(ecrPermissionsPolicy)}`,
  },
}
export const ecrConnection: TCloudConnection = {
  ...initialECRConnection,
  setup: {
    ...initialECRConnection.setup,
    permissions_policy: undefined,
    cli: connection.setup.cli,
    terraform: connection.setup.terraform,
    cloudformation: connection.setup.cloudformation,
  },
}
export const verifiedECRConnection: TCloudConnection = {
  ...ecrConnection,
  status: 'verified',
  status_message:
    'OIDC exchange, foreign-subject rejection, and ECR token retrieval passed. Component builds check repository access.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const failedECRConnection: TCloudConnection = {
  ...ecrConnection,
  status: 'error',
  status_message:
    'ECR token retrieval failed: AWS denied ecr:GetAuthorizationToken for this role.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
const garRepositories = [
  'projects/acme-project/locations/us-central1/repositories/acme-images',
  'projects/acme-project/locations/us-central1/repositories/acme-workers',
]
const garSubject = 'org:org-mock-001:connection:cc-gar-example'
const garPool =
  'projects/123456789012/locations/global/workloadIdentityPools/nuon-acme'
const garProvider = `${garPool}/providers/nuon`
const garAudience = `//iam.googleapis.com/${garProvider}`
const garMember = `principal://iam.googleapis.com/${garPool}/subject/${garSubject}`
const garPrincipal = 'nuon-images@acme-project.iam.gserviceaccount.com'
export const initialGARConnection: TCloudConnection = {
  ...connection,
  id: 'cc-gar-example',
  name: 'acme-gar-images',
  platform: 'gcp',
  target_id: 'acme-project',
  gcp_project_number: '123456789012',
  principal: garPrincipal,
  preset: 'gar',
  default_region: 'us-central1',
  setup: {
    issuer_url: 'https://api.example.com',
    subject: garSubject,
    audience: garAudience,
    preset: 'gar',
    cloudformation: '',
    trust_policy: {
      provider: garProvider,
      issuer_url: 'https://api.example.com',
      allowed_audiences: [garAudience],
      attribute_mapping: { 'google.subject': 'assertion.sub' },
      attribute_condition: `assertion.sub == '${garSubject}'`,
      service_account_binding: {
        service_account: garPrincipal,
        role: 'roles/iam.workloadIdentityUser',
        member: garMember,
      },
    },
    permissions_policy: {
      bindings: garRepositories.map((repository) => ({
        repository,
        role: 'roles/artifactregistry.reader',
        member: `serviceAccount:${garPrincipal}`,
      })),
    },
    cli: `# 1. Create the service account and workload identity pool/provider
gcloud iam service-accounts create nuon-images --project=acme-project
gcloud iam workload-identity-pools create nuon-acme --project=acme-project --location=global
gcloud iam workload-identity-pools providers create-oidc nuon --project=acme-project --location=global --workload-identity-pool=nuon-acme --issuer-uri=https://api.example.com --allowed-audiences='${garAudience}' --attribute-mapping='google.subject=assertion.sub' --attribute-condition="assertion.sub == '${garSubject}'"

# 2. Grant impersonation only to this connection subject
gcloud iam service-accounts add-iam-policy-binding ${garPrincipal} --project=acme-project --role=roles/iam.workloadIdentityUser --member='${garMember}'

# 3. Grant reader access on each selected repository
gcloud artifacts repositories add-iam-policy-binding acme-images --project=acme-project --location=us-central1 --role=roles/artifactregistry.reader --member=serviceAccount:${garPrincipal}
gcloud artifacts repositories add-iam-policy-binding acme-workers --project=acme-project --location=us-central1 --role=roles/artifactregistry.reader --member=serviceAccount:${garPrincipal}`,
    terraform: `resource "google_service_account" "nuon" {
  project    = "acme-project"
  account_id = "nuon-images"
}

resource "google_iam_workload_identity_pool" "nuon" {
  project                   = "acme-project"
  workload_identity_pool_id = "nuon-acme"
}

resource "google_iam_workload_identity_pool_provider" "nuon" {
  project                            = "acme-project"
  workload_identity_pool_id          = google_iam_workload_identity_pool.nuon.workload_identity_pool_id
  workload_identity_pool_provider_id = "nuon"
  attribute_mapping                  = { "google.subject" = "assertion.sub" }
  attribute_condition                = "assertion.sub == '${garSubject}'"
  oidc {
    issuer_uri        = "https://api.example.com"
    allowed_audiences = ["${garAudience}"]
  }
}

resource "google_service_account_iam_member" "nuon" {
  service_account_id = google_service_account.nuon.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "${garMember}"
}

resource "google_artifact_registry_repository_iam_member" "nuon" {
  for_each   = toset(["acme-images", "acme-workers"])
  project    = "acme-project"
  location   = "us-central1"
  repository = each.value
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:\${google_service_account.nuon.email}"
}`,
  },
}
export const garConnection: TCloudConnection = {
  ...initialGARConnection,
  setup: {
    ...initialGARConnection.setup,
    permissions_policy: undefined,
    cli: initialGARConnection.setup.cli!.split('# 3.')[0],
    terraform: initialGARConnection.setup.terraform!.split(
      'resource "google_artifact_registry_repository_iam_member"'
    )[0],
  },
}
export const verifiedGARConnection: TCloudConnection = {
  ...garConnection,
  status: 'verified',
  status_message:
    'GCP STS exchange, service account impersonation, and foreign-subject rejection passed. Component builds check repository access.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const failedGARConnection: TCloudConnection = {
  ...garConnection,
  status: 'error',
  status_message:
    'Federation trust check failed: GCP denied service account impersonation for this connection subject.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const failedGARTrustConnection: TCloudConnection = {
  ...garConnection,
  status: 'error',
  status_message:
    'Federation trust check failed: GCP denied service account impersonation for this connection subject.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
const azureSubscription = '11111111-1111-4111-8111-111111111111'
const azureTenant = '22222222-2222-4222-8222-222222222222'
const azureClient = '33333333-3333-4333-8333-333333333333'
const azureRegistry = `/subscriptions/${azureSubscription}/resourceGroups/acme-images/providers/Microsoft.ContainerRegistry/registries/acmeimages`
const acrSubject = 'org:org-mock-001:connection:cc-acr-example'
const acrRepositories = ['acme/api', 'acme/worker']
const acrCondition = `((!(ActionMatches{'Microsoft.ContainerRegistry/registries/repositories/content/read'}) AND !(ActionMatches{'Microsoft.ContainerRegistry/registries/repositories/metadata/read'})) OR (${acrRepositories.map((repository) => `@Request[Microsoft.ContainerRegistry/registries/repositories:name] StringEqualsIgnoreCase '${repository}'`).join(' OR ')}))`
const acrTrustPolicy = {
  name: 'nuon-cc-acr-example',
  issuer: 'https://api.example.com',
  subject: acrSubject,
  audiences: ['api://AzureADTokenExchange'],
}
export const initialACRConnection: TCloudConnection = {
  ...connection,
  id: 'cc-acr-example',
  name: 'acme-acr-images',
  platform: 'azure',
  target_id: azureSubscription,
  principal: azureClient,
  azure_tenant_id: azureTenant,
  azure_registry_id: azureRegistry,
  azure_registry_login_server: 'acmeimages-abc123.azurecr.io',
  preset: 'acr',
  setup: {
    issuer_url: 'https://api.example.com',
    subject: acrSubject,
    audience: 'api://AzureADTokenExchange',
    preset: 'acr',
    trust_policy: acrTrustPolicy,
    permissions_policy: {
      scope: azureRegistry,
      role: 'Container Registry Repository Reader',
      condition_version: '2.0',
      condition: acrCondition,
    },
    cli: `# 1. Check the existing identity and registry prerequisites
az account show --subscription '${azureSubscription}'
az ad app show --id '${azureClient}'
az ad sp show --id '${azureClient}'
az acr show --ids '${azureRegistry}' --query '{loginServer:loginServer,roleAssignmentMode:roleAssignmentMode}'
# The registry must already use RBAC + ABAC repository permissions.
# Review existing roles before changing registry modes; switching can break access.

# 2. Add federation for this exact connection subject
az ad app federated-credential create --id '${azureClient}' --parameters '${JSON.stringify(acrTrustPolicy)}'

# 3. Grant reader access only to the selected repository names
SERVICE_PRINCIPAL_ID=$(az ad sp show --id '${azureClient}' --query id --output tsv)
az role assignment create --assignee-object-id "$SERVICE_PRINCIPAL_ID" --assignee-principal-type ServicePrincipal --role 'Container Registry Repository Reader' --scope '${azureRegistry}' --condition-version '2.0' --condition "${acrCondition}"`,
    terraform: `# Use the existing application, service principal, and RBAC + ABAC registry.
# Do not switch registry permission modes without reviewing existing role assignments.
data "azuread_application" "nuon" {
  client_id = "${azureClient}"
}

data "azuread_service_principal" "nuon" {
  client_id = "${azureClient}"
}

resource "azuread_application_federated_identity_credential" "nuon" {
  application_id = data.azuread_application.nuon.id
  display_name   = "nuon-cc-acr-example"
  issuer         = "https://api.example.com"
  subject        = "${acrSubject}"
  audiences      = ["api://AzureADTokenExchange"]
}

resource "azurerm_role_assignment" "nuon_acr" {
  scope                = "${azureRegistry}"
  role_definition_name = "Container Registry Repository Reader"
  principal_id         = data.azuread_service_principal.nuon.object_id
  condition_version    = "2.0"
  condition            = "${acrCondition}"
}`,
  },
}
export const acrConnection: TCloudConnection = {
  ...initialACRConnection,
  setup: {
    ...initialACRConnection.setup,
    permissions_policy: undefined,
    cli: initialACRConnection.setup.cli!.split('# 3.')[0],
    terraform: initialACRConnection.setup.terraform!.split(
      'resource "azurerm_role_assignment"'
    )[0],
  },
}
export const verifiedACRConnection: TCloudConnection = {
  ...acrConnection,
  status: 'verified',
  status_message:
    'Entra federation, foreign-subject rejection, and ACR token exchange passed. Component builds check repository access.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const failedACRConnection: TCloudConnection = {
  ...acrConnection,
  status: 'error',
  status_message:
    'ACR token exchange could not be checked. Review the registry login server and Microsoft Entra identity, then retry.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const failedACRTrustConnection: TCloudConnection = {
  ...acrConnection,
  status: 'error',
  status_message:
    'Entra federation trust check failed: no matching federated identity credential was found for this connection subject.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const failedConnection: TCloudConnection = {
  ...connection,
  status: 'error',
  status_message:
    'AWS denied sts:AssumeRoleWithWebIdentity for this connection subject.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const verifiedConnection: TCloudConnection = {
  ...connection,
  status: 'verified',
  status_message: 'OIDC exchange and read-only CloudFormation probe passed.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const noop = () => {}

export const PageStory = ({ children }: { children: ReactNode }) => (
  <NotificationContext.Provider
    value={{
      emitNotification: async () => false,
      permission: 'default',
      requestPermission: async () => 'default',
      isSupported: false,
      settings: { permissionRequested: false },
      hasRequestedPermission: false,
      muted: false,
      toggleMute: noop,
    }}
  >
    <BreadcrumbContext.Provider
      value={{ breadcrumbLinks: [], isLoading: false, updateBreadcrumb: noop }}
    >
      <SidebarContext.Provider
        value={{
          isSidebarOpen: true,
          closeSidebar: noop,
          openSidebar: noop,
          toggleSidebar: noop,
        }}
      >
        {children}
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)
