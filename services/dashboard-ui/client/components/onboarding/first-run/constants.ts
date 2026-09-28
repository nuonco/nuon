import type { TIconVariant } from '@/components/common/Icon'

export type TCloud = 'aws' | 'gcp' | 'azure'
export type TPath = 'example' | 'own'

export const TEST_CLOUDS: TCloud[] = ['aws', 'gcp', 'azure']
// Clouds Kitchen Sink can be deployed to from the Start step.
export type TExampleCloud = 'aws' | 'gcp'
export const EXAMPLE_CLOUDS: TExampleCloud[] = ['aws', 'gcp']

export const isCloud = (value: unknown): value is TCloud =>
  typeof value === 'string' && (TEST_CLOUDS as string[]).includes(value)

export const CLOUD_LABEL: Record<TCloud, string> = { aws: 'AWS', gcp: 'GCP', azure: 'Azure' }

export const CLOUD_ICON: Record<TCloud, TIconVariant> = {
  aws: 'AWSColor',
  gcp: 'GCPColor',
  azure: 'AzureColor',
}

// Maps onto the install stack types: aws-cloudformation, gcp-terraform, azure-bicep.
// Only AWS (and Azure at subscription scope) get a console quick link; GCP is applied by hand.
export const CLOUD_CONNECT: Record<
  TCloud,
  {
    accountNoun: string
    stackLabel: string
    // What exists once generation finishes. AWS gets a console link; GCP and Azure (at its
    // default scope) get a template and commands, so the status line must not say "link".
    artifactNoun: string
    generating: string
    launch: string
    helper: string
    waitingHint: string
  }
> = {
  aws: {
    accountNoun: 'AWS account',
    stackLabel: 'CloudFormation stack',
    artifactNoun: 'CloudFormation stack link',
    generating: 'Generating your CloudFormation stack link...',
    launch: 'Open the CloudFormation stack',
    helper:
      'Opens a pre-filled CloudFormation stack in your AWS console. Create it there, then come back. This page updates on its own.',
    waitingHint: 'Create the CloudFormation stack in the AWS console tab, then come back.',
  },
  gcp: {
    accountNoun: 'GCP project',
    stackLabel: 'Terraform stack',
    artifactNoun: 'Terraform stack',
    generating: 'Generating your Terraform stack...',
    launch: 'Get the Terraform stack',
    helper:
      'Nuon generates a Terraform stack for your test GCP project. Apply it from your terminal, then come back. This page updates on its own.',
    waitingHint: 'Apply the Terraform stack from your terminal, then come back.',
  },
  azure: {
    accountNoun: 'Azure subscription',
    stackLabel: 'Bicep stack',
    artifactNoun: 'Bicep template',
    generating: 'Generating your Bicep template...',
    launch: 'Get the Azure commands',
    helper:
      'Nuon generates the Bicep template and the az commands that deploy it. Create the resource group and Key Vault, run the commands, then come back. This page updates on its own.',
    waitingHint: 'Run the az commands from your terminal, then come back.',
  },
}

export const CLOUD_REGIONS: Record<TCloud, { label: string; options: string[] }> = {
  aws: {
    label: 'AWS region',
    options: ['us-east-1', 'us-west-2', 'eu-west-1', 'ap-southeast-1'],
  },
  gcp: {
    label: 'GCP region',
    options: ['us-central1', 'us-east1', 'europe-west1', 'asia-southeast1'],
  },
  azure: {
    label: 'Azure location',
    options: ['eastus', 'westus2', 'westeurope', 'southeastasia'],
  },
}

export const defaultRegion = (cloud: TCloud) => CLOUD_REGIONS[cloud].options[0]

export const regionOptions = (cloud: TCloud) =>
  CLOUD_REGIONS[cloud].options.map((value) => ({ value, label: value }))

// The Nuon sandbox each cloud lands on, and the repo that provisions it.
export const CLOUD_SANDBOX: Record<TCloud, string> = {
  aws: 'nuonco/aws-eks-auto-sandbox',
  gcp: 'nuonco/gcp-gke-sandbox',
  azure: 'nuonco/azure-aks-sandbox',
}

export const SANDBOX_CLUSTER: Record<TCloud, string> = {
  aws: 'An EKS cluster and node group',
  gcp: 'A GKE Autopilot cluster',
  azure: 'An AKS cluster',
}
export const SANDBOX_PARTS = ['cluster', 'registry', 'ingress', 'namespaces']

export const KITCHEN_SINK_APP = 'kitchen-sink'
// Kitchen Sink's root config is AWS-only (runner_type = "aws"); its GCP variant
// lives in gcp/. Each gets its own app so both can exist in one org.
export const KITCHEN_SINK_VARIANTS: Record<TExampleCloud, { appName: string; directory: string }> = {
  aws: { appName: KITCHEN_SINK_APP, directory: '.' },
  gcp: { appName: `${KITCHEN_SINK_APP}-gcp`, directory: 'gcp' },
}
export const KITCHEN_SINK_REPO = 'nuonco/kitchen-sink'
export const KITCHEN_SINK_URL = `https://github.com/${KITCHEN_SINK_REPO}`
export const KITCHEN_SINK_LABEL = 'Kitchen Sink'
export const EXAMPLE_APP_FACTS = [
  'Helm chart: API, UI, and worker pods',
  'Pulumi S3 bucket and CI-built images',
  'Actions, policies, runbooks, app branches',
]

// The branch every first-run app template tracks, and where its config lives:
// the root of the repo, the way every example app config does.
export const DEFAULT_BRANCH_NAME = 'default'
export const TRACKED_GIT_BRANCH = 'main'
export const CONFIG_DIRECTORY = '.'

// Same rule ctl-api's entity_name validator enforces on app names.
export const APP_NAME_PATTERN = /^[a-z0-9_-]+$/
export const APP_NAME_RULE = 'Lowercase letters, numbers, underscores, and hyphens only.'

export const installNameFor = (appName: string, attempt: number) =>
  attempt === 0 ? `${appName}-test` : `${appName}-test-${attempt + 1}`
export const INSTALL_NAME_ATTEMPTS = 3

// The agent prompt. Shown and copied exactly as served.
export const PROMPT_URL = 'https://nuon.co/loop.md'

export const DEMO_REQUEST = 'https://nuon.co/demo-request'
export const CONTACT_MESSAGE = 'I would like help writing the app config for my first install.'

export const DOCS_URL = 'https://docs.nuon.co'
export const DOCS_MCP = 'https://docs.nuon.co/guides/agents/mcp-walkthrough'
export const DOCS_APPS = 'https://docs.nuon.co/concepts/apps'
export const DOCS_RUNNERS = 'https://docs.nuon.co/concepts/runners'
export const DOCS_SANDBOXES = 'https://docs.nuon.co/concepts/sandboxes'
export const DOCS_STACKS = 'https://docs.nuon.co/concepts/stacks'
export const DOCS_CONFIG_FILES = 'https://docs.nuon.co/configuration-files'
export const DOCS_LSP = 'https://docs.nuon.co/configuration-files#language-server-protocol-lsp'
export const AWS_QUICK_CREATE_DOCS =
  'https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-console-create-stacks-quick-create-links.html'
export const GCP_INFRA_MANAGER_DOCS = 'https://cloud.google.com/infrastructure-manager/docs'
export const VSCODE_EXTENSION = 'https://marketplace.visualstudio.com/items?itemName=Nuon.nuon-lsp'
export const LSP_NEOVIM_SETUP = 'https://github.com/nuonco/nuon/blob/main/bins/lsp/README.md#neovim'

export const CLI_INSTALL = 'brew install nuonco/tap/nuon'
export const CLI_LOGIN = 'nuon login'
export const CLI_SETUP = `${CLI_INSTALL}\n${CLI_LOGIN}`
export const MCP_ADD_CLAUDE =
  'claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes'
export const GIT_PUSH = 'git add .\ngit commit -m "Add Nuon app config"\ngit push origin main'
export const watchCommand = (installId: string) => `nuon installs workflows watch -i ${installId}`

// --- Manual steps: the app config files, stubbed per cloud --------------------

export interface IAppFileStub {
  name: string
  purpose: string
  badge: string
  required: boolean
  snippet: string
}

// Policy shapes per cloud, from nuonco/example-app-configs: kitchen-sink (AWS managed
// policy), gke-simple (gcp_predefined_role), aks-simple (azure_built_in_roles).
export const ROLE_POLICY: Record<TCloud, string[]> = {
  aws: ['managed_policy_name = "AdministratorAccess"'],
  gcp: ['name                = "owner"', 'gcp_predefined_role = "roles/owner"'],
  azure: ['name                 = "contributor"', 'azure_built_in_roles = ["Contributor"]'],
}

const ROLES = [
  ['provision', 'Provision the sandbox and components.'],
  ['maintenance', 'Operate and update components.'],
  ['deprovision', 'Tear the install down.'],
] as const

export const permissionsStub = (cloud: TCloud) =>
  ROLES.map(([role, description]) =>
    [
      `[${role}_role]`,
      `name        = "{{.nuon.install.id}}-${role}"`,
      `description = "${description}"`,
      ...(cloud === 'aws' ? [] : [`cloud_platform = "${cloud}"`]),
      `[[${role}_role.policies]]`,
      ...ROLE_POLICY[cloud],
    ].join('\n')
  ).join('\n\n')

export const appFileStubs = ({
  appName,
  cloud,
  repo,
}: {
  appName: string
  cloud: TCloud
  repo: string
}): IAppFileStub[] => [
  {
    name: 'metadata.toml',
    purpose: 'Names the app and pins the config version.',
    badge: 'Required',
    required: true,
    snippet: `version      = "v1"\ndisplay_name = "${appName}"\ndescription  = "What ${appName} does, in one sentence."`,
  },
  {
    name: 'runner.toml',
    purpose: 'Which cloud the Nuon runner operates in.',
    badge: 'Required',
    required: true,
    snippet: `# aws, azure, or gcp\nrunner_type = "${cloud}"`,
  },
  {
    name: 'sandbox.toml',
    purpose: 'The base infrastructure the app lands on: the Nuon sandbox for your test cloud.',
    badge: 'Required',
    required: true,
    snippet: `terraform_version = "1.11.3"\n\n[public_repo]\nrepo      = "${CLOUD_SANDBOX[cloud]}"\ndirectory = "."\nbranch    = "main"`,
  },
  {
    name: 'permissions.toml',
    purpose: 'The roles Nuon assumes in the customer account: provision, maintenance and deprovision.',
    badge: 'Required',
    required: true,
    snippet: permissionsStub(cloud),
  },
  {
    name: 'branch.toml',
    purpose: 'Tracks the repo you connected. Every push to main syncs the default app branch.',
    badge: 'Created for you',
    required: false,
    snippet: `name = "${DEFAULT_BRANCH_NAME}"\n\n[connected_repo]\nrepo      = "${repo}"\ndirectory = "${CONFIG_DIRECTORY}"\nbranch    = "${TRACKED_GIT_BRANCH}"`,
  },
  {
    name: 'components/api.toml',
    purpose: 'Your app: Helm charts, Terraform, images, manifests.',
    badge: 'One per component',
    required: false,
    snippet: `name       = "api"\ntype       = "helm_chart"\nchart_name = "api"\nnamespace  = "${appName}"\n\n[public_repo]\nrepo      = "${repo}"\ndirectory = "charts/api"\nbranch    = "main"`,
  },
]

// --- Stack step copy ----------------------------------------------------------

// How a customer creates the install stack, per cloud (docs/concepts/stacks.mdx and
// docs/platform-support/*): Terraform plus the platform's native format, except GCP,
// which is Terraform only. The first entry is the one this install uses.
export const STACK_METHODS: Record<TCloud, { name: string; how: string }[]> = {
  aws: [
    {
      name: 'CloudFormation quick-create',
      how: 'One pre-filled link. Your customer creates the stack in their console.',
    },
    { name: 'AWS CLI', how: 'The same CloudFormation template, from a terminal.' },
    {
      name: 'Terraform',
      how: 'Generated tfvars for the install-stacks/aws module, applied with terraform.',
    },
  ],
  gcp: [
    {
      name: 'Terraform',
      how: 'Generated tfvars for the install-stacks/gcp module. gcloud auth, then terraform apply. GCP is Terraform only.',
    },
  ],
  azure: [
    {
      name: 'Azure CLI (Bicep)',
      how: 'Create a resource group and Key Vault, then deploy the template with az. Nuon fills in the commands.',
    },
    {
      name: 'Terraform',
      how: 'Generated tfvars for the install-stacks/azure module, applied with terraform.',
    },
    {
      name: 'Deploy to Azure',
      how: 'A pre-filled portal link. Only when stack.toml sets deployment_scope = "subscription".',
    },
  ],
}

// Onboarding is a proof of concept, so the account is always framed as a test one.
export const testAccountLabel = (cloud: TCloud) => `your test ${CLOUD_CONNECT[cloud].accountNoun}`

export type TStageId = 'runner' | 'sandbox' | 'components'

export interface IBuildStage {
  id: TStageId
  label: string
  blurb: string
}

// What is left after the stack. Durations from docs/get-started: a healthy runner
// about a minute after the stack reports home; the sandbox is the long pole.
export const buildStages = (cloud: TCloud, appName: string): IBuildStage[] => [
  {
    id: 'runner',
    label: 'Nuon runner',
    blurb: `Runs in ${testAccountLabel(cloud)} and builds everything else, using the roles your stack granted. About 1 minute.`,
  },
  {
    id: 'sandbox',
    label: 'Nuon sandbox',
    blurb: `${SANDBOX_CLUSTER[cloud]}, a container registry, ingress and namespaces. About 15 to 20 minutes.`,
  },
  {
    id: 'components',
    label: 'Components',
    blurb: `${appName}'s Terraform, Helm charts and images, deployed into the sandbox. A few minutes.`,
  },
]
