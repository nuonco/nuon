import type { TIconVariant } from '@/components/common/Icon'
import { AWS_REGIONS, AZURE_REGIONS, GCP_REGIONS } from '@/configs/cloud-regions'
import { getFlagEmoji } from '@/utils/string-utils'

export type TCloud = 'aws' | 'gcp' | 'azure'
export type TPath = 'example' | 'own'

export const TEST_CLOUDS: TCloud[] = ['aws', 'gcp', 'azure']
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

export const CLOUD_CONNECT: Record<
  TCloud,
  {
    accountNoun: string
    stackLabel: string
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

type TRegionCatalogEntry = { text: string; value: string; iconVariant?: string }

const REGION_CATALOG: Record<TCloud, { label: string; regions: readonly TRegionCatalogEntry[]; withCode: boolean }> = {
  aws: { label: 'AWS region', regions: AWS_REGIONS, withCode: true },
  gcp: { label: 'GCP region', regions: GCP_REGIONS, withCode: true },
  azure: { label: 'Azure location', regions: AZURE_REGIONS, withCode: false },
}

const regionLabel = (region: TRegionCatalogEntry, withCode: boolean) => {
  if (!region.iconVariant) return region.text
  const name = `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text}`
  return withCode ? `${name} [${region.value}]` : name
}

export const regionFieldLabel = (cloud: TCloud) => REGION_CATALOG[cloud].label

export const regionOptions = (cloud: TCloud) => {
  const { regions, withCode } = REGION_CATALOG[cloud]
  return regions.map((region) => ({ value: region.value, label: regionLabel(region, withCode) }))
}

export const defaultRegion = (cloud: TCloud) => REGION_CATALOG[cloud].regions[0].value

export const isKnownRegion = (cloud: TCloud, region: string) =>
  REGION_CATALOG[cloud].regions.some((item) => item.value === region)

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
export const KITCHEN_SINK_VARIANTS: Record<
  TExampleCloud,
  { appName: string; directory: string; sandbox: string }
> = {
  aws: { appName: KITCHEN_SINK_APP, directory: 'kitchen-sink-aws', sandbox: 'nuonco/aws-eks-sandbox' },
  gcp: { appName: `${KITCHEN_SINK_APP}-gcp`, directory: 'kitchen-sink-gcp', sandbox: 'nuonco/gcp-gke-sandbox' },
}
export const KITCHEN_SINK_REPO = 'nuonco/kitchen-sink'
export const KITCHEN_SINK_URL = `https://github.com/${KITCHEN_SINK_REPO}`
export const KITCHEN_SINK_LABEL = 'Kitchen Sink'

export const DEFAULT_BRANCH_NAME = 'default'
export const TRACKED_GIT_BRANCH = 'main'
export const CONFIG_DIRECTORY = '.'

export const APP_NAME_PATTERN = /^[a-z0-9_-]+$/
export const APP_NAME_RULE = 'Lowercase letters, numbers, underscores, and hyphens only.'

export const installNameFor = (appName: string, attempt: number) =>
  attempt === 0 ? `${appName}-test` : `${appName}-test-${attempt + 1}`
export const INSTALL_NAME_ATTEMPTS = 3

export const PROMPT_URL = 'https://nuon.co/loop.md'

export const DEMO_REQUEST = 'https://nuon.co/demo-request'
export const CONTACT_MESSAGE = 'I would like help writing the app template for my first install.'

export const DOCS_URL = 'https://docs.nuon.co'
export const DOCS_MCP = 'https://docs.nuon.co/guides/agents/setup'
export const DOCS_APPS = 'https://docs.nuon.co/concepts/apps'
export const DOCS_APP_INPUTS = 'https://docs.nuon.co/concepts/app-inputs'
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
export const GIT_PUSH = 'git add .\ngit commit -m "Add Nuon app template"\ngit push origin main'
export const watchCommand = (installId: string) => `nuon installs workflows watch -i ${installId}`
export const watchPrompt = (appName: string) =>
  `Watch my ${appName} test install on Nuon while it provisions. Tell me as each step finishes, and if a step fails, pull its logs and explain what went wrong.`

export interface IAppFileStub {
  name: string
  purpose: string
  badge: string
  required: boolean
  snippet: string
}

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

export const testAccountLabel = (cloud: TCloud) => `your test ${CLOUD_CONNECT[cloud].accountNoun}`

export type TStageId = 'runner' | 'sandbox' | 'components'

export interface IBuildStage {
  id: TStageId
  label: string
  blurb: string
}

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

export const AWESOME_BYOC_URL = 'https://www.awesomebyoc.com'

// From github.com/nuonco/awesome-byoc. BYOC platform vendors are left out on purpose.
export const BYOC_REFERENCES: { name: string; category: string; blurb: string; href: string }[] = [
  {
    name: 'Datadog',
    category: 'Observability',
    blurb: 'Log management with indexing and search running in your own Kubernetes cluster.',
    href: 'https://docs.datadoghq.com/cloudprem/introduction/',
  },
  {
    name: 'Snowflake Openflow',
    category: 'Data integration',
    blurb: 'Moves data into Snowflake while keeping compute in your environment.',
    href: 'https://www.snowflake.com/en/blog/openflow-byoc-data-integration/',
  },
  {
    name: 'Vercel',
    category: 'Dev platforms',
    blurb: 'Deploy enterprise apps and AI agents in your own cloud, with Vercel running the control plane.',
    href: 'https://vercel.com/enterprise',
  },
  {
    name: 'ClickHouse',
    category: 'Databases',
    blurb: 'Real-time OLAP database for analytics, deployed in your own cloud account.',
    href: 'https://clickhouse.com/cloud/bring-your-own-cloud',
  },
  {
    name: 'Redpanda',
    category: 'Streaming',
    blurb: 'Kafka-compatible streaming platform with a BYOC deployment option.',
    href: 'https://www.redpanda.com/product/bring-your-own-cloud-byoc',
  },
  {
    name: 'LangChain',
    category: 'AI',
    blurb: 'LangSmith traces, datasets and prompts stored in your own AWS account.',
    href: 'https://www.langchain.com/blog/langsmith-byoc-is-now-generally-available-on-aws',
  },
  {
    name: 'Grafana',
    category: 'Observability',
    blurb: 'Metrics, logs and traces with a bring your own cloud option.',
    href: 'https://grafana.com/products/bring-your-own-cloud-byoc/',
  },
  {
    name: 'Pinecone',
    category: 'Databases',
    blurb: 'Vector database for AI applications, deployable in your own cloud.',
    href: 'https://docs.pinecone.io/guides/production/bring-your-own-cloud',
  },
  {
    name: 'Fivetran',
    category: 'Data integration',
    blurb: 'Managed ELT pipelines with a data-plane agent running in your own VPC.',
    href: 'https://fivetran.com/docs/deployment-models/hybrid-deployment',
  },
  {
    name: 'Retool',
    category: 'Internal tools',
    blurb: 'Internal tool builder with vendor-managed deployment in your AWS VPC.',
    href: 'https://docs.retool.com/self-hosted/retool-managed/concepts/architecture',
  },
  {
    name: 'Glean',
    category: 'Apps',
    blurb: 'Enterprise AI search and assistant platform managed in your own cloud.',
    href: 'https://docs.glean.com/get-started/build/about-self-hosted',
  },
  {
    name: 'Cohere',
    category: 'AI',
    blurb: 'Enterprise LLM platform with model containers deployed in your own cloud.',
    href: 'https://docs.cohere.com/docs/private-deployment-overview',
  },
  {
    name: 'Braintrust',
    category: 'AI',
    blurb: 'LLM evals platform with its data plane running in your own cloud VPC.',
    href: 'https://www.braintrust.dev/docs/guides/self-hosting',
  },
  {
    name: 'CockroachDB',
    category: 'Databases',
    blurb: 'Distributed SQL database with a bring your own cloud deployment option.',
    href: 'https://www.cockroachlabs.com/product/cloud/bring-your-own-cloud/',
  },
]

export const STACK_RENDER_LABEL: Record<TCloud, string> = {
  aws: 'Build the CloudFormation template',
  gcp: 'Build the Terraform stack',
  azure: 'Build the Bicep template',
}

// Normal waits on app.nuon.co are about 2 minutes for builds and under a minute
// for the stack link, so these leave room before calling a wait slow.
export const DEPLOY_SLOW_MS = 10 * 60_000
export const STACK_SLOW_MS = 5 * 60_000
export const DEPLOY_HELP_MESSAGE = 'My first install is stuck while Nuon gets the app template ready.'
export const STACK_HELP_MESSAGE = 'My first install is stuck while Nuon generates the install stack.'
