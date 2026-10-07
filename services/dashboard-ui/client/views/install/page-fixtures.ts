import type { TFixture, TFixtureReply } from './install-fixture'
import { VIEW_APP_ID, VIEW_INSTALL_ID } from './install-fixture'
import { ACTIVE_DEPLOYMENT_STATUSES } from '@/components/installs/DeploymentDetail/deployment-progress'
import {
  APP_CONFIG_ID,
  pendingReply,
  viewInstall,
  withChrome,
} from './install-fixtures'

const ok = (body: unknown, paginated = false): TFixtureReply => ({
  body,
  paginated,
})

const fail = (error: string): TFixtureReply => ({
  body: { error },
  status: 500,
})

const isoDay = (daysAgo: number) => {
  const date = new Date()
  date.setUTCDate(date.getUTCDate() - daysAgo)
  return date.toISOString().slice(0, 10)
}

const readme = [
  '# Payments',
  '',
  'Open the app at https://payments.example.com.',
  '',
  '## Access',
  '',
  'The API is available on port 443. Workers process the checkout queue.',
].join('\n')

const incompleteReadme = [
  '# Payments',
  '',
  'Open the app at {{ .nuon.install_stack.quick_link_url }}.',
  '',
  'Region: {{ .nuon.install_stack.outputs.region }}',
].join('\n')

export const overviewFixture = (
  state: 'empty' | 'warnings' | 'rendered'
): TFixture =>
  withChrome(viewInstall(), (url) => {
    if (!url.pathname.endsWith('/readme')) return undefined
    if (state === 'empty') return ok({})
    if (state === 'warnings') {
      return ok({
        readme: incompleteReadme,
        warnings: [
          'unable to execute template: {{ .nuon.install_stack.quick_link_url }}',
          'unable to execute template: {{ .nuon.install_stack.outputs.region }}',
        ],
      })
    }
    return ok({ readme })
  })

const deployment = (
  id: string,
  type: string,
  status: string,
  createdAt: string,
  title: string,
  summary: string,
  resources: { components: string[]; images: string[]; stack?: boolean; sandbox?: boolean },
  extra: { component_name?: string; image?: Record<string, string> } = {}
) => ({
  id,
  type,
  status,
  created_at: createdAt,
  title,
  summary,
  app_branch: { id: 'br-1', name: 'main', sha: 'abc123def456' },
  workflow: { id: `wf-${id}`, name: title, type },
  affected_resources: resources,
  change_groups: [],
  ...extra,
})

const deployments = [
  deployment(
    'dep-api-image',
    'image_update',
    'in-progress',
    '2026-10-05T14:00:00Z',
    'Deploy api image',
    'Roll ghcr.io/acme/api from v1.2.3 to v1.3.0.',
    { components: ['api'], images: ['ghcr.io/acme/api'] },
    {
      component_name: 'api',
      image: { repository: 'ghcr.io/acme/api', previous_tag: 'v1.2.3', next_tag: 'v1.3.0' },
    }
  ),
  deployment(
    'dep-worker-image',
    'image_update',
    'success',
    '2026-10-05T09:30:00Z',
    'Deploy worker image',
    'Roll ghcr.io/acme/worker from v2.3.1 to v2.4.0.',
    { components: ['worker'], images: ['ghcr.io/acme/worker'] },
    {
      component_name: 'worker',
      image: { repository: 'ghcr.io/acme/worker', previous_tag: 'v2.3.1', next_tag: 'v2.4.0' },
    }
  ),
  deployment(
    'dep-worker-scale',
    'component_deploy',
    'in-progress',
    '2026-10-05T11:10:00Z',
    'Scale checkout workers',
    'Raise the worker replica count from 3 to 6.',
    { components: ['worker'], images: [] },
    { component_name: 'worker' }
  ),
  deployment(
    'dep-ledger',
    'component_deploy',
    'success',
    '2026-10-03T16:40:00Z',
    'Update ledger module',
    'Apply the ledger terraform module from main.',
    { components: ['ledger'], images: [] },
    { component_name: 'ledger' }
  ),
  deployment(
    'dep-webhooks',
    'component_deploy',
    'error',
    '2026-10-04T18:20:00Z',
    'Update webhooks manifest',
    'The webhooks deployment failed its readiness check.',
    { components: ['webhooks'], images: [] },
    { component_name: 'webhooks' }
  ),
  deployment(
    'dep-provision',
    'provision',
    'success',
    '2026-09-28T16:00:00Z',
    'Initial provision',
    'Provision the stack, sandbox, and every component.',
    { stack: true, sandbox: true, components: ['api', 'worker', 'ledger'], images: [] }
  ),
  deployment(
    'dep-sandbox',
    'sandbox_reprovision',
    'success',
    '2026-10-01T13:00:00Z',
    'Reprovision sandbox',
    'Replace the sandbox network after the VPC change.',
    { sandbox: true, components: [], images: [] }
  ),
  deployment(
    'dep-config',
    'install_config_update',
    'error',
    '2026-10-02T11:00:00Z',
    'Config update',
    'Install inputs failed to apply.',
    { components: [], images: [] }
  ),
]

const deploymentSummary = (record: {
  id: string
  type: string
  status: string
  created_at: string
  title: string
  summary: string
  affected_resources: { components: string[]; stack?: boolean; sandbox?: boolean }
}) => {
  const resources = record.affected_resources
  const targets: { name: string; step_target_type: string; component_name?: string }[] = [
    ...(resources.stack ? [{ name: 'await install stack', step_target_type: 'install_stack_versions' }] : []),
    ...(resources.sandbox ? [{ name: 'provision sandbox apply plan', step_target_type: 'install_sandbox_runs' }] : []),
    ...resources.components.map((component) => ({
      name: `apply ${component}`,
      step_target_type: 'install_deploys',
      component_name: component,
    })),
  ]
  return {
    id: record.id,
    type: record.type === 'image_update' ? 'component_deploy' : record.type,
    status: record.status,
    created_at: record.created_at,
    title: record.title,
    activity: record.summary,
    finished: record.status === 'success' || record.status === 'error',
    steps: targets.map((target, index) => ({
      ...target,
      id: `${record.id}-step-${index}`,
      group_idx: index + 1,
      execution_type: 'system',
      status: record.status === 'success' || index < targets.length - 1 ? 'success' : record.status,
    })),
  }
}

const deploymentPage = (rows: unknown[], total?: number) =>
  ok({
    deployments: rows,
    page: 0,
    offset: 0,
    limit: 20,
    has_more: false,
    total,
  })

export const deploymentsFixture = (
  state: 'loading' | 'empty' | 'results'
): TFixture =>
  withChrome(viewInstall(), (url) => {
    const detail = deployments.find((item) =>
      url.pathname.endsWith(`/deployments/${item.id}`)
    )
    if (detail) return ok(detail)
    const workflow = deployments.find(
      (item) => url.pathname === `/v1/workflows/${item.id}`
    )
    if (workflow) {
      const summary = deploymentSummary(workflow)
      return ok({
        id: summary.id,
        name: summary.title,
        type: summary.type,
        created_at: summary.created_at,
        finished: summary.finished,
        status: {
          status: summary.status,
          status_human_description: summary.activity,
        },
        steps: summary.steps.map(({ status, component_name, ...step }) => ({
          ...step,
          status: { status },
          metadata: component_name ? { component_name } : undefined,
        })),
      })
    }
    if (url.pathname.endsWith('/deployments')) return deploymentPage(deployments)
    if (!url.pathname.endsWith('/deployment-summaries')) return undefined
    if (state === 'loading') return pendingReply()
    if (state === 'empty') return deploymentPage([])
    const search = (url.searchParams.get('search') ?? '').toLowerCase()
    const status = url.searchParams.get('status') ?? ''
    const type = url.searchParams.get('type') ?? ''
    const lifecycle = url.searchParams.get('state')
    const rows = deployments.map(deploymentSummary).filter((item) => {
      const active = ACTIVE_DEPLOYMENT_STATUSES.has(item.status)
      if (lifecycle === 'active' && !active) return false
      if (lifecycle === 'finished' && active) return false
      const haystack = [item.id, item.title].join(' ').toLowerCase()
      if (search && !haystack.includes(search)) return false
      if (status && !status.split(',').includes(item.status)) return false
      if (type && !type.split(',').includes(item.type)) return false
      return true
    })
    return deploymentPage(
      rows,
      lifecycle === 'active' ? rows.length : undefined
    )
  })

const healthComponent = (
  name: string,
  health: string,
  uptime = 99.9,
  observed = 30 * 86400
) => ({
  install_component_id: `instcmp-${name}`,
  component_id: `cmp-${name}`,
  component_name: name,
  current_health: health,
  uptime_percent: uptime,
  observed_seconds: observed,
})

const healthDay = (health: string, index: number) => ({
  date: isoDay(29 - index),
  health,
  unhealthy_seconds: health === 'unhealthy' ? 3600 : 0,
  degraded_seconds: health === 'degraded' ? 1800 : 0,
  unknown_seconds: health === 'unknown' ? 86400 : 0,
  observed_seconds: health === 'unknown' ? 0 : 86400,
})

const healthResource = (
  name: string,
  kind: string,
  health: string,
  source: 'component' | 'sandbox' = 'component'
) => ({
  install_component_id: source === 'sandbox' ? undefined : `instcmp-${name}`,
  component_id: source === 'sandbox' ? undefined : `cmp-${name}`,
  source,
  owner_name: source === 'sandbox' ? name : undefined,
  kind,
  namespace: source === 'sandbox' ? 'cert-manager' : 'payments',
  name,
  health,
  provider: 'kubernetes',
  observed_at: new Date(Date.now() - 60_000).toISOString(),
  ...(health === 'unhealthy'
    ? { message: 'Available replicas are below the desired count.' }
    : health === 'degraded'
      ? { message: 'One pod is not ready.' }
      : {}),
})

type THealthState =
  | 'healthy'
  | 'degraded'
  | 'unhealthy'
  | 'access-error'
  | 'no-observations'
  | 'empty'
  | 'loading'

export const healthFixture = (state: THealthState): TFixture =>
  withChrome(viewInstall(), (url) => {
    const path = url.pathname
    const resourcesPath = path.endsWith('/resources')
    const timelinePath = path.endsWith('/health/timeline')
    const componentsPath = path.endsWith('/components')
    if (!resourcesPath && !timelinePath && !componentsPath) return undefined
    if (state === 'loading') return pendingReply()

    if (componentsPath) {
      return ok(
        ['api', 'worker', 'ledger', 'webhooks'].map((name) => ({
          id: `instcmp-${name}`,
          component: { id: `cmp-${name}`, name },
        })),
        true
      )
    }

    if (resourcesPath) {
      const rows =
        state === 'healthy'
          ? [
              healthResource('api', 'Deployment', 'healthy'),
              healthResource('api', 'Service', 'healthy'),
              healthResource('worker', 'Deployment', 'healthy'),
              healthResource('worker', 'Service', 'healthy'),
              healthResource('ledger', 'Deployment', 'healthy'),
              healthResource('webhooks', 'Ingress', 'healthy'),
              healthResource('cert-manager', 'Deployment', 'healthy', 'sandbox'),
              healthResource('cert-manager', 'Certificate', 'healthy', 'sandbox'),
            ]
          : state === 'degraded'
            ? [
                healthResource('api', 'Deployment', 'degraded'),
                healthResource('worker', 'Deployment', 'healthy'),
              ]
            : state === 'unhealthy'
              ? [healthResource('api', 'Deployment', 'unhealthy')]
              : []
      const health = url.searchParams.get('health')
      return ok(health ? rows.filter((row) => row.health === health) : rows)
    }

    const current =
      state === 'degraded'
        ? 'degraded'
        : state === 'unhealthy'
          ? 'unhealthy'
          : state === 'healthy' || state === 'empty'
            ? 'healthy'
            : 'unknown'
    const components =
      state === 'access-error'
        ? []
        : state === 'no-observations'
          ? [healthComponent('api', 'unknown', 0, 0)]
          : state === 'degraded'
            ? [healthComponent('api', 'degraded', 98.7), healthComponent('worker', 'healthy', 99.91)]
            : state === 'unhealthy'
              ? [healthComponent('api', 'unhealthy', 97.4)]
              : state === 'empty'
                ? [healthComponent('api', 'healthy'), healthComponent('worker', 'healthy', 99.91)]
                : [healthComponent('api', 'healthy', 99.98), healthComponent('worker', 'healthy', 99.91)]

    return ok({
      days: 30,
      uptime_percent: current === 'unknown' ? 0 : 99.9,
      observed_seconds: components.reduce(
        (total, component) => total + component.observed_seconds,
        0
      ),
      current_health: current,
      cluster_access_error:
        state === 'access-error'
          ? 'The runner cannot list pods in the payments namespace.'
          : undefined,
      components,
      daily: Array.from({ length: 30 }, (_, index) => healthDay(current, index)),
    })
  })

const componentCatalog = [
  ['cmp-api', 'api', 'helm_chart', 'ghcr.io/acme/api:1.3.0'],
  ['cmp-worker', 'worker', 'helm_chart', 'ghcr.io/acme/worker:2.4.0'],
  ['cmp-ledger', 'ledger', 'terraform_module', ''],
  ['cmp-webhooks', 'webhooks', 'kubernetes_manifest', ''],
  ['cmp-scheduler', 'scheduler', 'pulumi', ''],
  ['cmp-migrate', 'migrate', 'job', ''],
  ['cmp-api-image', 'api', 'external_image', 'ghcr.io/acme/api:1.3.0'],
  ['cmp-worker-image', 'worker', 'docker_build', 'ghcr.io/acme/worker:2.4.0'],
  ['cmp-ledger-image', 'ledger', 'external_image', 'ghcr.io/acme/ledger:0.9.1'],
  ['cmp-webhooks-image', 'webhooks', 'external_image', 'ghcr.io/acme/webhooks:1.1.0'],
] as const

const degradedIds = new Set([
  'cmp-api',
  'cmp-ledger',
  'cmp-api-image',
  'cmp-ledger-image',
])

const installComponent = (id: string, name: string, type: string, status: string, image: string) => ({
  id: `ic-${id}`,
  component_id: id,
  enabled: true,
  status,
  status_v2: { status },
  component: { id, name, type, app_id: VIEW_APP_ID },
  install_deploys: [
    {
      id: `dep-${id}`,
      build_id: `bld-${id}`,
      created_at: '2026-10-04T12:00:00Z',
      status,
      status_v2: { status },
      component_build: image
        ? {
            id: `bld-${id}`,
            source_ref: image,
            status,
            created_at: '2026-10-04T12:00:00Z',
          }
        : undefined,
    },
  ],
})

const resourceConfig = (state: 'empty' | 'active' | 'degraded') => ({
  id: APP_CONFIG_ID,
  version: state === 'empty' ? undefined : '12',
  stack:
    state === 'empty'
      ? undefined
      : {
          name: 'payments',
          type: 'terraform',
          runner_nested_template_url: 'https://example.com/acme/runner.tar.gz',
          vpc_nested_template_url: 'https://example.com/acme/vpc.tar.gz',
          custom_nested_stacks: [
            {
              name: 'observability',
              template_url: 'https://example.com/acme/observability.tar.gz',
            },
          ],
        },
  sandbox:
    state === 'empty'
      ? undefined
      : {
          type: 'terraform',
          terraform_version: '1.9.8',
          public_repo: { repo: 'acme/payments-sandbox', directory: 'sandbox' },
        },
  action_workflow_configs: [],
  component_config_connections:
    state === 'empty'
      ? []
      : componentCatalog
          .filter(([, , type, image]) => image && (type === 'external_image' || type === 'docker_build'))
          .map(([id, , , image]) => ({
            component_id: id,
            external_image: { image_url: image },
          })),
})

const resourceComponents = (state: 'empty' | 'active' | 'degraded') =>
  state === 'empty'
    ? []
    : componentCatalog.map(([id, name, type, image]) =>
        installComponent(
          id,
          name,
          type,
          state === 'degraded' && degradedIds.has(id) ? 'error' : 'active',
          image
        )
      )

export const resourcesFixture = (
  state: 'empty' | 'provisioning' | 'active' | 'degraded'
): TFixture => {
  const install = viewInstall(
    state === 'active' || state === 'degraded'
      ? {
          sandbox: {
            id: 'sbx-1',
            status: state === 'degraded' ? 'error' : 'active',
            status_v2: { status: state === 'degraded' ? 'error' : 'active' },
          },
          install_sandbox_runs: [
            {
              id: 'srun-1',
              created_at: '2026-10-04T12:00:00Z',
              run_type: 'provision',
            },
          ],
        }
      : { sandbox_status: undefined }
  )

  return withChrome(install, (url) => {
    const path = url.pathname
    if (path.includes('/configs/') && path.includes('/apps/')) {
      if (state === 'provisioning') return pendingReply()
      if (state === 'degraded') {
        return fail('The stack config could not be read from the app branch.')
      }
      return ok(resourceConfig(state === 'empty' ? 'empty' : 'active'))
    }
    if (path.endsWith('/stack')) {
      if (state === 'provisioning') return pendingReply()
      if (state === 'empty') return ok({})
      return ok({
        id: 'stack-1',
        install_stack_outputs: {
          data: {
            region: 'us-west-2',
            cluster_name: 'payments',
            vpc_id: 'vpc-0acme00001',
            api_url: 'https://payments.example.com',
          },
        },
      })
    }
    if (path.includes('/sandbox/builds')) {
      if (state === 'provisioning') return pendingReply()
      return ok([], true)
    }
    if (path.endsWith('/components') || /\/components\/[^/]+$/.test(path)) {
      if (state === 'provisioning') return pendingReply()
      const rows = resourceComponents(state === 'empty' ? 'empty' : state)
      const id = path.split('/components/')[1]
      if (id) return ok(rows.find((row) => row.component_id === id) ?? {})
      const types = url.searchParams.get('types') ?? ''
      const query = (url.searchParams.get('q') ?? '').toLowerCase()
      return ok(
        rows.filter((row) => {
          if (types && !types.split(',').includes(row.component.type)) return false
          if (query && !row.component.name.includes(query)) return false
          return true
        }),
        true
      )
    }
    if (path === '/v1/builds') {
      const id = url.searchParams.get('component_id')
      const row = resourceComponents(state === 'degraded' ? 'degraded' : 'active').find(
        (item) => item.component_id === id
      )
      const build = row?.install_deploys[0].component_build
      return ok(build ? [build] : [], true)
    }
    return undefined
  })
}

const branch = {
  id: 'br-1',
  name: 'main',
  latest_run: {
    id: 'run-1',
    status: 'success',
    created_at: '2026-10-03T12:00:00Z',
    app_branch: { id: 'br-1', name: 'main' },
    head_sha: 'abc123def4567890',
    vcs_connection_commit: {
      sha: 'abc123def4567890',
      message: 'Raise the checkout worker retry budget',
      author_name: 'Example Developer',
    },
    app_branch_config: {
      connected_github_vcs_config: { repo: 'acme/payments', branch: 'main' },
    },
  },
}

const hex = (value: string) =>
  Array.from(new TextEncoder().encode(value))
    .map((byte) => byte.toString(16).padStart(2, '0'))
    .join('')

const overrideName = (kind: string, component: string) =>
  `nuon_component_override_v1_${kind}_${hex(component)}`

const inputGroups = [
  {
    id: 'group-cloud',
    name: 'cloud',
    display_name: 'Cloud',
    description: 'Where this install runs.',
  },
  {
    id: 'group-platform',
    name: 'platform',
    display_name: 'Platform',
    description: 'How the payments service is exposed.',
  },
  {
    id: 'group-secrets',
    name: 'secrets',
    display_name: 'Secrets',
    description: 'Values are redacted after they are written.',
  },
  {
    id: 'group-overrides',
    name: 'nuon_component_overrides',
    display_name: 'Component overrides',
  },
]

const configInputs = [
  {
    id: 'in-region',
    group_id: 'group-cloud',
    name: 'region',
    display_name: 'Region',
    default: 'us-east-1',
  },
  {
    id: 'in-account',
    group_id: 'group-cloud',
    name: 'account_id',
    display_name: 'AWS account',
    default: '000000000000',
  },
  {
    id: 'in-domain',
    group_id: 'group-platform',
    name: 'domain',
    display_name: 'Domain',
    default: 'example.com',
  },
  {
    id: 'in-size',
    group_id: 'group-platform',
    name: 'cluster_size',
    display_name: 'Cluster size',
    default: '3',
  },
  {
    id: 'in-token',
    group_id: 'group-secrets',
    name: 'api_token',
    display_name: 'API token',
    sensitive: true,
  },
  {
    id: 'ov-api-enabled',
    group_id: 'group-overrides',
    name: overrideName('enabled', 'api'),
    index: 0,
  },
  {
    id: 'ov-api-helm',
    group_id: 'group-overrides',
    name: overrideName('helm_values', 'api'),
    index: 1,
  },
  {
    id: 'ov-worker-enabled',
    group_id: 'group-overrides',
    name: overrideName('enabled', 'worker'),
    index: 2,
  },
  {
    id: 'ov-worker-helm',
    group_id: 'group-overrides',
    name: overrideName('helm_values', 'worker'),
    index: 3,
  },
  {
    id: 'ov-ledger-enabled',
    group_id: 'group-overrides',
    name: overrideName('enabled', 'ledger'),
    index: 4,
  },
  {
    id: 'ov-ledger-tf',
    group_id: 'group-overrides',
    name: overrideName('tf_vars', 'ledger'),
    index: 5,
  },
  {
    id: 'ov-webhooks-enabled',
    group_id: 'group-overrides',
    name: overrideName('enabled', 'webhooks'),
    index: 6,
  },
]

const inputValues = {
  region: 'us-west-2',
  account_id: '000000000000',
  domain: 'payments.example.com',
  cluster_size: '6',
  api_token: '••••••••',
  [overrideName('enabled', 'api')]: 'true',
  [overrideName('helm_values', 'api')]:
    'replicaCount: 3\nimage:\n  tag: "1.3.0"\ningress:\n  host: payments.example.com',
  [overrideName('enabled', 'worker')]: 'true',
  [overrideName('helm_values', 'worker')]:
    'replicaCount: 6\nresources:\n  requests:\n    cpu: 500m',
  [overrideName('enabled', 'ledger')]: 'true',
  [overrideName('tf_vars', 'ledger')]:
    'instance_type = "m6i.large"\nbackup_retention_days = 14',
  [overrideName('enabled', 'webhooks')]: 'false',
}

const inputConfig = {
  id: APP_CONFIG_ID,
  version: '4',
  action_workflow_configs: [],
  component_config_connections: [],
  input: {
    inputs: configInputs,
    input_groups: inputGroups,
  },
}

const configToml = `[install]
name = "payments-prod"

[inputs]
region = "us-west-2"
account_id = "000000000000"
domain = "payments.example.com"
cluster_size = "6"
`

const configuredComponents = [
  { component: { name: 'api', type: 'helm_chart' } },
  { component: { name: 'worker', type: 'helm_chart' } },
  { component: { name: 'ledger', type: 'terraform_module' } },
  { component: { name: 'webhooks', type: 'kubernetes_manifest' } },
]

export const configurationFixture = (
  state: 'unset' | 'manual' | 'managed'
): TFixture => {
  const install = viewInstall(
    state === 'unset'
      ? { app_branch: undefined, app_branch_id: undefined, app_config_id: undefined }
      : {
          install_components: configuredComponents,
          ...(state === 'managed'
            ? { metadata: { managed_by: 'nuon/cli/install-config' } }
            : {}),
        }
  )

  return withChrome(install, (url) => {
    const path = url.pathname
    if (path.includes(`/branches/br-1`)) return ok(branch)
    if (path.includes('/configs/') && path.includes('/apps/')) {
      return ok(state === 'unset' ? { id: APP_CONFIG_ID } : inputConfig)
    }
    if (path.endsWith('/inputs/current')) {
      return ok({
        redacted_values: state === 'unset' ? {} : inputValues,
      })
    }
    if (path.endsWith('/state')) {
      if (state === 'unset') return ok({})
      return ok({
        region: 'us-west-2',
        account_id: '000000000000',
        domain: 'payments.example.com',
        cluster_size: '6',
        cluster: 'payments',
      })
    }
    if (path.endsWith('/generate-cli-install-config')) {
      return ok({ content: configToml, filename: 'install.toml' })
    }
    if (path.endsWith('/config-versions')) {
      return ok([
        {
          id: 'cfg-9',
          file_path: 'install.toml',
          created_at: '2026-10-04T18:00:00Z',
        },
      ])
    }
    return undefined
  })
}

const planStep = (
  id: string,
  name: string,
  groupIdx: number,
  status: string,
  planType: string,
  componentName: string,
  counts: { create: number; update: number; delete: number; replace: number; noop: number },
  countsState = 'ok'
) => ({
  id,
  name,
  group_idx: groupIdx,
  execution_type: planType === 'noop' ? 'system' : 'approval',
  finished: status === 'success' || status === 'error',
  execution_time: status === 'success' ? 8_200_000_000 : undefined,
  status: {
    status,
    status_human_description:
      status === 'error' ? 'Helm upgrade failed for api' : undefined,
  },
  metadata: { component_name: componentName },
  approval:
    planType === 'noop'
      ? undefined
      : {
          id: `approval-${id}`,
          type: planType,
          changes_state: countsState,
          changes_create: counts.create,
          changes_update: counts.update,
          changes_delete: counts.delete,
          changes_replace: counts.replace,
          changes_noop: counts.noop,
        },
})

const syncStep = {
  id: 'step-sync',
  name: 'Sync install configuration',
  group_idx: 0,
  execution_type: 'system',
  finished: true,
  execution_time: 1_800_000_000,
  status: { status: 'success' },
}

const deploymentRecord = (status: string) => ({
  id: 'wf-deploy-1',
  type: 'app_branch_update',
  status,
  created_at: '2026-10-02T16:00:00Z',
  title: 'Update production install',
  summary: 'Apply the latest app template to this install.',
  workflow: {
    id: 'wf-deploy-1',
    name: 'Update production install',
    type: 'app_branch_config_update',
  },
  app_branch: {
    id: 'br-1',
    name: 'main',
    run_id: 'run-main-42',
    sha: 'abc123def456',
  },
  affected_resources: {
    stack: true,
    components: ['api', 'worker'],
    images: ['ghcr.io/acme/api'],
  },
  change_groups: [],
})

const counts = {
  stack: { create: 2, update: 3, delete: 0, replace: 0, noop: 8 },
  api: { create: 0, update: 2, delete: 0, replace: 0, noop: 5 },
}

const deploymentWorkflow = (state: 'in-progress' | 'awaiting' | 'succeeded' | 'failed') => {
  const status =
    state === 'in-progress'
      ? 'in-progress'
      : state === 'awaiting'
        ? 'approval-awaiting'
        : state === 'succeeded'
          ? 'success'
          : 'error'
  const description =
    state === 'in-progress'
      ? 'Deploying components'
      : state === 'awaiting'
        ? 'Waiting for plan approval'
        : state === 'succeeded'
          ? 'Deployment finished'
          : 'Helm upgrade failed for api'
  const steps =
    state === 'awaiting'
      ? [
          syncStep,
          planStep(
            'step-stack',
            'Plan stack',
            1,
            'approval-awaiting',
            'terraform_plan',
            'Install stack',
            counts.stack
          ),
        ]
      : [
          syncStep,
          planStep(
            'step-stack',
            'Plan stack',
            1,
            'success',
            'terraform_plan',
            'Install stack',
            counts.stack
          ),
          state === 'in-progress'
            ? planStep(
                'step-api',
                'Plan api',
                2,
                'in-progress',
                'helm_approval',
                'api',
                counts.api,
                'unknown'
              )
            : state === 'failed'
              ? planStep(
                  'step-deploy',
                  'Deploy components',
                  2,
                  'error',
                  'helm_approval',
                  'api',
                  counts.api,
                  'unknown'
                )
              : planStep(
                  'step-api',
                  'Plan api',
                  2,
                  'success',
                  'helm_approval',
                  'api',
                  counts.api
                ),
        ]

  return {
    id: 'wf-deploy-1',
    name: 'Update production install',
    type: 'app_branch_config_update',
    created_at: '2026-10-02T16:00:00Z',
    finished: state === 'succeeded' || state === 'failed',
    approval_option: state === 'awaiting' ? 'prompt' : undefined,
    status: { status, status_human_description: description },
    steps,
  }
}

export const deploymentDetailFixture = (
  state: 'in-progress' | 'awaiting' | 'succeeded' | 'failed'
): TFixture =>
  withChrome(viewInstall(), (url) => {
    const path = url.pathname
    const record = deploymentRecord(
      state === 'awaiting'
        ? 'pending'
        : state === 'succeeded'
          ? 'success'
          : state === 'failed'
            ? 'error'
            : 'in-progress'
    )
    if (path.endsWith(`/deployments/${record.id}`)) return ok(record)
    if (path.endsWith('/deployments')) return deploymentPage([record])
    if (path.endsWith('/deployment-summaries')) {
      const search = url.searchParams.get('search') ?? ''
      const active = ACTIVE_DEPLOYMENT_STATUSES.has(record.status)
      const lifecycle = url.searchParams.get('state')
      const matchesState =
        !lifecycle || (lifecycle === 'active' ? active : !active)
      const matchesSearch =
        !search ||
        record.id.includes(search) ||
        record.title.toLowerCase().includes(search.toLowerCase())
      const rows =
        matchesState && matchesSearch ? [deploymentSummary(record)] : []
      return deploymentPage(
        rows,
        lifecycle === 'active' ? rows.length : undefined
      )
    }
    if (path === '/v1/workflows/wf-deploy-1')
      return ok(deploymentWorkflow(state))
    if (path.includes('/comparison')) {
      return ok({
        config_diff_content: '',
        source_diff_content: { files: [] },
      })
    }
    return undefined
  })
