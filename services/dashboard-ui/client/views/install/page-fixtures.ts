import type { TFixture, TFixtureReply } from './install-fixture'
import { VIEW_APP_ID, VIEW_INSTALL_ID } from './install-fixture'
import { ACTIVE_DEPLOYMENT_STATUSES } from '@/components/installs/DeploymentDetail/deployment-progress'
import { healthPreviewFixture } from '@/components/installs/InstallHealth/health-preview-fixtures'
import {
  alignmentResource,
  inputUpdateResource,
  installOverview,
  openDeploymentResource,
  stateDriftResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import { mixedHelmPlan } from '@/lib/fixtures/plan-diffs/helm'
import { mockPlan as kubernetesPlan } from '@/lib/fixtures/plan-diffs/kubernetes'
import { securityGroupUpdatePlan } from '@/lib/fixtures/plan-diffs/terraform'
import type { TInstallAlignmentResource, TInstallOverview } from '@/types'
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
  resources: {
    components: string[]
    images: string[]
    stack?: boolean
    sandbox?: boolean
  },
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
      image: {
        repository: 'ghcr.io/acme/api',
        previous_tag: 'v1.2.3',
        next_tag: 'v1.3.0',
      },
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
      image: {
        repository: 'ghcr.io/acme/worker',
        previous_tag: 'v2.3.1',
        next_tag: 'v2.4.0',
      },
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
    'dep-ledger-plan',
    'app_branch_update',
    'approval-awaiting',
    '2026-10-05T13:20:00Z',
    'Release ledger 2.1',
    'Waiting for approval of the ledger plan.',
    { components: ['ledger', 'api'], images: [] }
  ),
  deployment(
    'dep-webhooks-retry',
    'component_deploy',
    'failed-pending-retry',
    '2026-10-05T12:45:00Z',
    'Update webhooks routes',
    'Applying webhooks failed; retry or skip the step.',
    { components: ['webhooks'], images: [] },
    { component_name: 'webhooks' }
  ),
  deployment(
    'dep-api-retrying',
    'component_deploy',
    'retrying',
    '2026-10-05T12:10:00Z',
    'Rotate api certificates',
    'Retrying the api apply after a transient error.',
    { components: ['api'], images: [] },
    { component_name: 'api' }
  ),
  deployment(
    'dep-stack-queued',
    'stack_update',
    'queued',
    '2026-10-05T14:05:00Z',
    'Stack update',
    'Queued behind the running deployments.',
    { stack: true, components: [], images: [] }
  ),
  deployment(
    'dep-secrets-pending',
    'install_config_update',
    'pending',
    '2026-10-05T14:10:00Z',
    'Sync secrets',
    'Waiting for a runner to pick up the job.',
    { components: [], images: [] }
  ),
  deployment(
    'dep-worker-cancelled',
    'component_deploy',
    'cancelled',
    '2026-10-04T08:15:00Z',
    'Rollback worker',
    'Cancelled by an operator before apply.',
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
    {
      stack: true,
      sandbox: true,
      components: ['api', 'worker', 'ledger'],
      images: [],
    }
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
  affected_resources: {
    components: string[]
    stack?: boolean
    sandbox?: boolean
  }
}) => {
  const resources = record.affected_resources
  const targets: {
    name: string
    step_target_type: string
    component_name?: string
  }[] = [
    ...(resources.stack
      ? [
          {
            name: 'await install stack',
            step_target_type: 'install_stack_versions',
          },
        ]
      : []),
    ...(resources.sandbox
      ? [
          {
            name: 'provision sandbox apply plan',
            step_target_type: 'install_sandbox_runs',
          },
        ]
      : []),
    ...resources.components.map((component) => ({
      name: `apply ${component}`,
      step_target_type: 'install_deploys',
      component_name: component,
    })),
  ]
  const finished = ['success', 'error', 'cancelled'].includes(record.status)
  const lastStepStatus: Record<string, string> = {
    'failed-pending-retry': 'error',
    retrying: 'in-progress',
    queued: 'pending',
    pending: 'pending',
  }
  return {
    id: record.id,
    type: record.type === 'image_update' ? 'component_deploy' : record.type,
    status: record.status,
    created_at: record.created_at,
    title: record.title,
    activity: record.summary,
    finished,
    steps: targets.map((target, index) => {
      const last = index === targets.length - 1
      const notStarted = ['pending', 'queued'].includes(record.status)
      const approval = last && record.status === 'approval-awaiting'
      return {
        ...target,
        name: approval ? target.name.replace(/^apply /, 'plan ') : target.name,
        id: `${record.id}-step-${index}`,
        group_idx: index + 1,
        execution_type: approval ? 'approval' : 'system',
        status: notStarted
          ? 'pending'
          : record.status === 'success' || !last
            ? 'success'
            : (lastStepStatus[record.status] ?? record.status),
      }
    }),
  }
}

const deploymentPage = (rows: unknown[], total?: number, hasMore = false) =>
  ok({
    deployments: rows,
    page: 0,
    offset: 0,
    limit: 20,
    has_more: hasMore,
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
    if (url.pathname.endsWith('/deployments'))
      return deploymentPage(deployments)
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
    const rank = (item: { status: string }) =>
      url.searchParams.get('sort') !== 'attention'
        ? 0
        : ['in-progress', 'retrying'].includes(item.status)
          ? 0
          : item.status === 'approval-awaiting'
            ? 1
            : item.status === 'failed-pending-retry'
              ? 2
              : 3
    rows.sort(
      (left, right) =>
        rank(left) - rank(right) ||
        right.created_at.localeCompare(left.created_at)
    )
    const limit = Number(url.searchParams.get('limit') ?? rows.length)
    return deploymentPage(
      rows.slice(0, limit),
      lifecycle === 'active' ? rows.length : undefined,
      rows.length > limit
    )
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
  details: JSON.stringify(
    kind === 'Deployment'
      ? {
          spec: {
            replicas: 3,
            selector: { matchLabels: { app: name } },
            strategy: { type: 'RollingUpdate' },
          },
          status: {
            replicas: 3,
            readyReplicas:
              health === 'unhealthy' ? 0 : health === 'degraded' ? 2 : 3,
            availableReplicas:
              health === 'unhealthy' ? 0 : health === 'degraded' ? 2 : 3,
            updatedReplicas: 3,
            conditions: [
              {
                type: 'Available',
                status: health === 'healthy' ? 'True' : 'False',
                reason:
                  health === 'healthy'
                    ? 'MinimumReplicasAvailable'
                    : 'MinimumReplicasUnavailable',
                message:
                  health === 'healthy'
                    ? 'Deployment has minimum availability.'
                    : 'Deployment does not have minimum availability.',
              },
            ],
          },
        }
      : kind === 'Service'
        ? {
            spec: {
              type: 'ClusterIP',
              clusterIP: '10.0.0.42',
              selector: { app: name },
              ports: [
                { name: 'http', port: 80, targetPort: 8080, protocol: 'TCP' },
              ],
            },
          }
        : kind === 'Ingress'
          ? {
              spec: {
                ingressClassName: 'nginx',
                rules: [
                  {
                    host: 'api.example.com',
                    http: {
                      paths: [
                        {
                          path: '/',
                          backend: {
                            service: { name: 'api', port: { number: 80 } },
                          },
                        },
                      ],
                    },
                  },
                ],
              },
            }
          : { status: { conditions: [{ type: 'Ready', status: 'True' }] } }
  ),
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
  withChrome(
    viewInstall({
      health_cluster_error:
        state === 'access-error'
          ? 'The runner cannot list Pods in the acme namespace.'
          : undefined,
    }),
    (url) => {
      const path = url.pathname
      if (path.endsWith('/status')) {
        const status =
          state === 'healthy' || state === 'empty'
            ? 'healthy'
            : state === 'degraded' || state === 'unhealthy'
              ? state
              : 'unknown'
        return ok({
          deployments: {
            status: 'success',
            status_human_description: 'Up to date',
          },
          resources: { status, status_human_description: status },
          health_checks: {
            status: ['healthy', 'degraded', 'unhealthy'].includes(state)
              ? 'active'
              : 'unknown',
            status_human_description: [
              'healthy',
              'degraded',
              'unhealthy',
            ].includes(state)
              ? 'Passing'
              : 'No observations',
          },
        })
      }
      const resourcesPath = path.endsWith('/resources')
      const componentsPath = path.endsWith('/components')
      if (!resourcesPath && !componentsPath) return undefined
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
        const sample = healthPreviewFixture(state)
        const rows = ['healthy', 'degraded', 'unhealthy'].includes(state)
          ? [
              ...sample.groups.flatMap((group) =>
                group.items.map((item) => item.resource)
              ),
              ...sample.checks,
              healthResource(
                'cert-manager',
                'Deployment',
                'healthy',
                'sandbox'
              ),
              {
                ...healthResource(
                  'cert-manager',
                  'Certificate',
                  'unknown',
                  'sandbox'
                ),
                api_group: 'cert-manager.io',
              },
              {
                name: 'primary',
                kind: 'aws_db_instance',
                provider: 'aws',
                source: 'component',
                install_component_id: 'instcmp-ledger',
                health: 'not-applicable',
                observed_at: new Date(Date.now() - 86_400_000).toISOString(),
              },
            ]
          : []
        const health = url.searchParams.get('health')
        return ok(health ? rows.filter((row) => row.health === health) : rows)
      }
    }
  )

const componentCatalog = [
  ['cmp-api', 'api', 'helm_chart', 'ghcr.io/acme/api:1.3.0'],
  ['cmp-worker', 'worker', 'helm_chart', 'ghcr.io/acme/worker:2.4.0'],
  ['cmp-ledger', 'ledger', 'terraform_module', ''],
  ['cmp-webhooks', 'webhooks', 'kubernetes_manifest', ''],
  ['cmp-scheduler', 'scheduler', 'pulumi', ''],
  ['cmp-migrate', 'migrate', 'job', ''],
  ['cmp-api-image', 'api-image', 'external_image', 'ghcr.io/acme/api:1.3.0'],
  [
    'cmp-worker-image',
    'worker-image',
    'external_image',
    'ghcr.io/acme/worker:2.4.0',
  ],
  [
    'cmp-ledger-image',
    'ledger-image',
    'external_image',
    'ghcr.io/acme/ledger:0.9.1',
  ],
  [
    'cmp-webhooks-image',
    'webhooks-image',
    'external_image',
    'ghcr.io/acme/webhooks:1.1.0',
  ],
] as const

const degradedIds = new Set([
  'cmp-api',
  'cmp-ledger',
  'cmp-api-image',
  'cmp-ledger-image',
])

const installComponent = (
  id: string,
  name: string,
  type: string,
  status: string,
  image: string
) => ({
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
      : componentCatalog.map(([id, name, type, image]) => ({
          component_id: id,
          component_name: name,
          type,
          external_image:
            type === 'external_image' ? { image_url: image } : undefined,
          references:
            (
              {
                'cmp-api': ['api-image'],
                'cmp-worker': ['worker-image'],
                'cmp-ledger': ['api-image', 'ledger-image'],
                'cmp-webhooks': ['api-image', 'webhooks-image'],
              } as Record<string, string[]>
            )[id]?.map(
              (name) =>
                `component|${name}|image.repository|nuon.components.${name}.outputs.image.repository`
            ) ?? [],
          component_dependency_ids:
            id === 'cmp-scheduler' ? ['cmp-api-image'] : [],
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

export type TResourcesAlignment =
  | 'in-sync'
  | 'inputs-unverified'
  | 'drifted'
  | 'ongoing'
  | 'image-sync-cancelled'
  | 'input-updates'
  | 'torn-down'
  | 'component-history'
  | 'paged'
  | 'unavailable'
  | 'error'

const driftedAlignment = (inputUpdatesOnly = false) => {
  const component = (id: string, name: string, type: string) => ({
    id: `ic-${id}`,
    name,
    component_id: id,
    component_type: type,
    expected_config_id: APP_CONFIG_ID,
  })
  const resources = [
    alignmentResource({
      id: 'stack-1',
      kind: 'stack',
      name: 'stack',
      component_id: undefined,
      component_type: undefined,
    }),
    alignmentResource({
      id: 'sbx-1',
      kind: 'sandbox',
      name: 'sandbox',
      component_id: undefined,
      component_type: undefined,
    }),
    unappliedResource(component('cmp-api', 'api', 'helm_chart')),
    openDeploymentResource({
      ...component('cmp-worker', 'worker', 'helm_chart'),
      open_deployment: {
        id: 'wf-open-ic-cmp-worker',
        status: 'approval-awaiting',
        type: 'input_update',
        created_at: new Date(Date.now() - 1_800_000).toISOString(),
      },
      explanation: {
        summary: 'These changes are waiting for approval.',
        status: 'approval-awaiting',
        workflow_id: 'wf-open-ic-cmp-worker',
      },
    }),
    stateDriftResource(component('cmp-ledger', 'ledger', 'terraform_module')),
    inputUpdateResource({
      ...component('cmp-webhooks', 'webhooks', 'kubernetes_manifest'),
      expected_build_id: 'bld-cmp-webhooks',
      applied_build_id: 'bld-cmp-webhooks',
      expected_config_id: APP_CONFIG_ID,
      applied_config_id: APP_CONFIG_ID,
    }),
    inputUpdateResource({
      ...component('cmp-search', 'search', 'helm_chart'),
      status: 'inactive',
      reason: 'This component is enabled, but hasn’t been deployed.',
      expected_build_id: 'bld-cmp-search',
      expected_config_id: APP_CONFIG_ID,
      applied_build_id: undefined,
      applied_config_id: undefined,
      last_attempt: {
        id: 'wf-input-enable-search',
        workflow_id: 'wf-input-enable-search',
        type: 'component_enabled',
        status: 'cancelled',
        created_at: new Date(Date.now() - 3_600_000).toISOString(),
      },
      explanation: {
        summary:
          'Enabling this component was cancelled. It hasn’t been deployed.',
        status: 'cancelled',
        workflow_id: 'wf-input-enable-search',
      },
    }),
    alignmentResource(component('cmp-scheduler', 'scheduler', 'pulumi')),
    openDeploymentResource({
      ...component('cmp-migrate', 'migrate', 'job'),
      open_deployment: {
        id: 'wf-open-ic-cmp-migrate',
        status: 'in-progress',
        created_at: new Date(Date.now() - 900_000).toISOString(),
      },
      explanation: {
        summary:
          'The deployment is in progress. These changes haven’t taken effect yet.',
        status: 'in-progress',
        workflow_id: 'wf-open-ic-cmp-migrate',
      },
    }),
    alignmentResource({
      ...component('cmp-reports', 'reports', 'helm_chart'),
      category: 'disabled',
      status: 'disabled',
    }),
    alignmentResource({
      ...component('cmp-retired-worker', 'retired-worker', 'helm_chart'),
      category: 'torn-down',
      status: 'inactive',
      applied_build_id: undefined,
      applied_config_id: undefined,
      plan: undefined,
    }),
    unappliedResource({
      ...component('cmp-api-image', 'api-image', 'external_image'),
      kind: 'image',
      reason: 'This image version hasn’t been synced.',
      plan: undefined,
      explanation: {
        summary:
          'This install’s configuration specifies an image version that hasn’t been copied to its registry. The latest image sync failed.',
        status: 'error',
        workflow_id: 'wf-attempt-ic-cmp-api-image',
      },
    }),
    alignmentResource({
      ...component('cmp-worker-image', 'worker-image', 'external_image'),
      kind: 'image',
    }),
    alignmentResource({
      ...component('cmp-ledger-image', 'ledger-image', 'external_image'),
      kind: 'image',
      category: 'unknown',
      reason: 'The last successful image sync could not be determined.',
    }),
    alignmentResource({
      ...component('cmp-webhooks-image', 'webhooks-image', 'external_image'),
      kind: 'image',
    }),
  ]
  return installOverview(
    inputUpdatesOnly
      ? resources.filter((resource) =>
          [
            'cmp-search',
            'cmp-webhooks',
            'cmp-scheduler',
            'cmp-reports',
            'cmp-retired-worker',
          ].includes(resource.component_id ?? '')
        )
      : resources,
    inputUpdatesOnly
      ? { status: 'current', commits_behind: 0, pending_commits: [] }
      : {}
  )
}

const inSyncAlignment = () =>
  installOverview(
    [
      alignmentResource({
        id: 'stack-1',
        kind: 'stack',
        name: 'stack',
        component_id: undefined,
        component_type: undefined,
      }),
      alignmentResource({
        id: 'sbx-1',
        kind: 'sandbox',
        name: 'sandbox',
        component_id: undefined,
        component_type: undefined,
      }),
      ...componentCatalog.map(([id, name, type]) =>
        alignmentResource({
          id: `ic-${id}`,
          name,
          component_id: id,
          component_type: type,
          kind: type === 'external_image' ? 'image' : 'component',
          expected_config_id: APP_CONFIG_ID,
        })
      ),
    ],
    { commits_behind: 0, pending_commits: [] }
  )

export type TResolveDriftScenario =
  | 'matching-plan'
  | 'changed-plan'
  | 'auto-approved'
  | 'policy-denied'

type TDriftResolution = {
  id: string
  resource: TInstallAlignmentResource
  requestId: string
  startedAt: number
  approvedAt?: number
  approvedPlanId?: string
  response?: { type: string; note: string }
}

const resourcePlanFixture = (
  resource: TInstallAlignmentResource,
  changed = false
) => {
  if (resource.component_type === 'helm_chart') {
    const contents = structuredClone(mixedHelmPlan)
    if (changed)
      contents.helm_content_diff[0].after =
        contents.helm_content_diff[0].after.replace(
          'replicas: 3',
          'replicas: 4'
        )
    return { type: 'helm_approval' as const, contents, deletes: 1 }
  }
  if (resource.component_type === 'kubernetes_manifest') {
    const contents = structuredClone(kubernetesPlan)
    if (changed)
      contents.k8s_content_diff[0].entries[0].applied =
        'postgres://review-required:5432/db'
    return {
      type: 'kubernetes_manifest_approval' as const,
      contents,
      deletes: 0,
    }
  }
  const contents = structuredClone(securityGroupUpdatePlan)
  if (changed)
    contents.resource_changes[0].change.after.description =
      'Web access restricted to the private network'
  return { type: 'terraform_plan' as const, contents, deletes: 0 }
}

const resolveDriftFixture = (
  overview: TInstallOverview,
  components: ReturnType<typeof resourceComponents>,
  scenario: TResolveDriftScenario
): TFixture => {
  const resolutions = new Map<string, TDriftResolution>()

  const snapshot = (resolution: TDriftResolution) => {
    const { id, resource, startedAt, approvedAt, response } = resolution
    const isImage = resource.kind === 'image'
    const planning = Date.now() - startedAt < 2000
    const denied =
      !planning &&
      (scenario === 'policy-denied' ||
        (!!response && response.type !== 'approve'))
    const finished =
      !planning && (denied || (!!approvedAt && Date.now() - approvedAt >= 2000))
    const status = planning
      ? 'in-progress'
      : denied
        ? 'error'
        : finished
          ? 'success'
          : approvedAt
            ? 'in-progress'
            : 'approval-awaiting'
    const description = planning
      ? isImage
        ? 'Preparing the registry sync'
        : 'Planning against the saved inputs'
      : denied
        ? scenario === 'policy-denied'
          ? 'Policy denied this deployment. The plan was not applied.'
          : 'Plan denied. The changes were not applied.'
        : finished
          ? isImage
            ? 'Image synced to the registry'
            : 'Drift resolved'
          : approvedAt
            ? isImage
              ? 'Syncing the image to the registry'
              : 'Applying the fresh plan'
            : 'Review the fresh plan before applying it'
    const planStatus = planning
      ? 'in-progress'
      : denied
        ? 'error'
        : approvedAt
          ? 'success'
          : 'approval-awaiting'
    const planFixture = resourcePlanFixture(resource)
    const plan = {
      ...planStep(
        `${id}-plan`,
        `Plan ${resource.name}`,
        1,
        planStatus,
        planFixture.type,
        resource.name!,
        {
          create: 1,
          update: 1,
          delete: planFixture.deletes,
          replace: 0,
          noop: 0,
        }
      ),
      install_workflow_id: id,
      owner_id: VIEW_INSTALL_ID,
      step_target_type: 'install_deploys',
      step_target_id: `${id}-deploy`,
      status: {
        status: planStatus,
        status_human_description: denied ? description : undefined,
      },
    }
    const workflow = {
      id,
      name: `Resolve drift for ${resource.name}`,
      type: 'manual_deploy',
      owner_id: VIEW_INSTALL_ID,
      created_at: new Date(startedAt).toISOString(),
      finished,
      approval_option: scenario === 'auto-approved' ? 'approve-all' : 'prompt',
      metadata: { resolve_drift: 'true', install_id: VIEW_INSTALL_ID },
      status: {
        status,
        status_human_description: description,
        metadata: { all_steps_loaded: true },
      },
      steps: [
        {
          ...syncStep,
          id: `${id}-prepare`,
          name: 'Prepare saved inputs',
          install_workflow_id: id,
        },
        ...(!isImage
          ? [
              {
                ...plan,
                approval: planning
                  ? undefined
                  : {
                      ...plan.approval,
                      response: denied
                        ? (response ?? { type: 'deny', note: description })
                        : approvedAt
                          ? {
                              type: 'approve',
                              note: resolution.approvedPlanId
                                ? 'Reused approval for an identical fresh plan'
                                : (response?.note ??
                                  'Approved by the install policy'),
                            }
                          : response,
                    },
              },
            ]
          : []),
        {
          id: `${id}-apply`,
          name: `${resource.kind === 'image' ? 'Sync' : 'Apply'} ${resource.name}`,
          group_idx: 1,
          execution_type: 'system',
          install_workflow_id: id,
          finished,
          status: {
            status: denied
              ? 'not-attempted'
              : finished
                ? 'success'
                : !planning && approvedAt
                  ? 'in-progress'
                  : 'pending',
          },
          metadata: { component_name: resource.name },
        },
      ],
    }
    const record = {
      id,
      title: workflow.name,
      type: 'manual_deploy',
      status,
      created_at: workflow.created_at,
      summary: description,
      workflow: { id, name: workflow.name, type: workflow.type },
      affected_resources: {
        components: resource.kind === 'image' ? [] : [resource.name!],
        images: resource.kind === 'image' ? [resource.name!] : [],
      },
      change_groups: [],
    }
    resource.open_deployment = finished
      ? undefined
      : { id, status, created_at: workflow.created_at }
    if (finished) {
      resource.last_attempt = {
        id: `${id}-deploy`,
        workflow_id: id,
        status,
        created_at: workflow.created_at,
      }
    }
    if (finished && !denied) {
      resource.category = 'in-sync'
      resource.status = 'active'
      resource.reason = undefined
      resource.explanation = undefined
      resource.plan = undefined
      const component = components.find(
        (row) => row.component_id === resource.component_id
      )
      if (component) {
        component.status = 'active'
        component.status_v2.status = 'active'
        const deploy = {
          ...component.install_deploys[0],
          id: `${id}-deploy`,
          created_at: workflow.created_at,
          updated_at: new Date(approvedAt! + 2000).toISOString(),
          status: 'active',
          status_v2: { status: 'active' },
        }
        component.install_deploys = [deploy]
      }
    }
    return { workflow, record }
  }

  return (url, init) => {
    const path = url.pathname
    const snapshots = Array.from(resolutions.values(), snapshot)
    if (
      path === `/v1/installs/${VIEW_INSTALL_ID}/resolve-drift` &&
      init?.method === 'POST'
    ) {
      const body = JSON.parse(String(init.body))
      const existing = Array.from(resolutions.values()).find(
        (entry) =>
          entry.requestId === body.request_id ||
          (entry.resource.id === body.resource_id &&
            !snapshot(entry).workflow.finished)
      )
      if (existing) return ok({ workflow_id: existing.id })
      const resource = overview.alignment?.resources?.find(
        (item) => item.id === body.resource_id
      )
      if (!resource)
        return { status: 404, body: { error: 'Resource not found' } }
      const startedAt = Date.now()
      const id = `wf-resolve-${resource.id}-${resolutions.size + 1}`
      resolutions.set(id, {
        id,
        resource,
        requestId: body.request_id,
        startedAt,
        approvedPlanId: body.approved_plan_id,
        approvedAt:
          resource.kind === 'image' ||
          scenario === 'auto-approved' ||
          (body.approved_plan_id && scenario !== 'changed-plan')
            ? startedAt + 2000
            : undefined,
      })
      snapshot(resolutions.get(id)!)
      return { status: 201, body: { workflow_id: id }, delayMs: 500 }
    }
    const match = path.match(/\/workflows\/(wf-resolve-[^/]+)(?:\/|$)/)
    const resolution = match ? resolutions.get(match[1]) : undefined
    if (resolution) {
      const { workflow } = snapshot(resolution)
      if (path.endsWith('/response') && init?.method === 'POST') {
        const body = JSON.parse(String(init.body))
        resolution.response = { type: body.response_type, note: body.note }
        if (body.response_type === 'approve') {
          resolution.approvedAt = Date.now()
          resolution.approvedPlanId = undefined
        }
        return { status: 201, body: resolution.response }
      }
      if (path.endsWith('/contents')) {
        return {
          body: resourcePlanFixture(
            resolution.resource,
            scenario === 'changed-plan'
          ).contents,
          delayMs: 300,
        }
      }
      if (path.endsWith('/sse')) return { body: workflow, sseEvent: 'workflow' }
      const stepId = path.match(/\/steps\/([^/]+)$/)?.[1]
      if (stepId) return ok(workflow.steps.find((step) => step.id === stepId))
      if (path.endsWith(resolution.id)) return ok(workflow)
    }
    const record = snapshots.find((item) =>
      path.endsWith(`/deployments/${item.record.id}`)
    )?.record
    if (record) return ok(record)
    const deploy = Array.from(resolutions.values()).find((item) =>
      path.endsWith(`/deploys/${item.id}-deploy`)
    )
    if (deploy) {
      return ok({
        id: `${deploy.id}-deploy`,
        component_id: deploy.resource.component_id,
        component_name: deploy.resource.name,
        created_at: new Date(deploy.startedAt).toISOString(),
        status_v2: snapshot(deploy).workflow.status,
      })
    }
    if (resolutions.size && path.endsWith('/deployments'))
      return deploymentPage(snapshots.map((item) => item.record))
    return undefined
  }
}

export const resourcesFixture = (
  state: 'empty' | 'provisioning' | 'active' | 'degraded',
  alignment: TResourcesAlignment = 'in-sync',
  managedByConfig = false,
  resolutionScenario: TResolveDriftScenario = 'matching-plan'
): TFixture => {
  const hasDrift =
    alignment === 'drifted' ||
    alignment === 'input-updates' ||
    alignment === 'image-sync-cancelled' ||
    alignment === 'ongoing'
  const overview = hasDrift
    ? driftedAlignment(alignment === 'input-updates')
    : inSyncAlignment()
  for (const resource of overview.alignment?.resources ?? []) {
    if (resource.plan) {
      const plan = resourcePlanFixture(resource)
      resource.plan.type = plan.type
      resource.plan.changes_delete = plan.deletes
    }
  }
  if (alignment === 'unavailable') {
    overview.alignment = {
      error: 'Resource alignment unavailable',
      resources: [],
    }
    overview.branch_tracking!.runs_behind = 2
    overview.branch_tracking!.commits_behind = 2
  }
  if (alignment === 'inputs-unverified') {
    const component = overview.alignment?.resources?.find(
      (resource) => resource.id === 'ic-cmp-api'
    )
    if (component) {
      component.expected_build_id = 'bld-cmp-api'
      component.applied_build_id = 'bld-cmp-api'
      component.expected_config_id = APP_CONFIG_ID
      component.applied_config_id = APP_CONFIG_ID
      component.reason =
        'Inputs not verified. Applied input values weren’t recorded for this deployment.'
    }
  }
  if (alignment === 'image-sync-cancelled') {
    const image = overview.alignment?.resources?.find(
      (resource) => resource.id === 'ic-cmp-api-image'
    )
    if (image) {
      image.status = 'cancelled'
      image.last_attempt = { ...image.last_attempt!, status: 'cancelled' }
      image.explanation = {
        summary:
          'The image sync was cancelled before this version was copied to the install’s registry.',
        status: 'cancelled',
        workflow_id: image.last_attempt.workflow_id,
      }
    }
  }
  if (alignment === 'ongoing') {
    for (const resource of overview.alignment?.resources ?? []) {
      if (
        ['ic-cmp-scheduler', 'ic-cmp-worker-image'].includes(resource.id ?? '')
      ) {
        resource.open_deployment = {
          id: `wf-open-${resource.id}`,
          status: 'queued',
          created_at: new Date(Date.now() - 600_000).toISOString(),
        }
        resource.explanation = {
          summary:
            resource.kind === 'image'
              ? 'The image sync is queued and hasn’t started yet.'
              : 'The deployment is queued and hasn’t started yet.',
          workflow_id: `wf-open-${resource.id}`,
          status: 'queued',
        }
      }
    }
  }
  const components = resourceComponents(
    state === 'empty' ? 'empty' : state === 'provisioning' ? 'active' : state
  )
  const config = resourceConfig(state === 'empty' ? 'empty' : 'active')
  if (alignment === 'component-history') {
    config.component_config_connections =
      config.component_config_connections.map((connection) => ({
        ...connection,
        id: `ccc-${connection.component_id}`,
        app_config_id: APP_CONFIG_ID,
        version: 12,
      }))
    const resources = [
      alignmentResource({
        id: 'ic-cmp-api',
        component_id: 'cmp-api',
        name: 'api',
        status: 'inactive',
        category: 'changes-not-applied',
        reason: 'This component is enabled, but hasn’t been deployed.',
        expected_build_id: 'bld-cmp-api',
        expected_config_id: APP_CONFIG_ID,
        explanation: {
          summary: 'No deployment has started for this component.',
        },
      }),
      alignmentResource({
        id: 'ic-cmp-worker',
        component_id: 'cmp-worker',
        name: 'worker',
        status: 'disabled',
        category: 'disabled',
        reason: 'component is disabled on this install',
        expected_build_id: 'bld-cmp-worker',
        expected_config_id: APP_CONFIG_ID,
      }),
      alignmentResource({
        id: 'ic-cmp-ledger',
        component_id: 'cmp-ledger',
        name: 'ledger',
        component_type: 'terraform_module',
        category: 'unknown',
        reason: 'the applied deploy for this component could not be determined',
        expected_build_id: 'bld-cmp-ledger',
        expected_config_id: APP_CONFIG_ID,
      }),
    ]
    overview.alignment!.resources = [
      ...overview.alignment!.resources!.filter(
        (resource) => resource.kind !== 'component'
      ),
      ...resources,
    ]
    components.splice(
      0,
      components.length,
      ...components.filter(
        (component) => component.component.type === 'external_image'
      ),
      ...resources.map((resource) => ({
        ...installComponent(
          resource.component_id!,
          resource.name!,
          resource.component_type!,
          resource.status!,
          ''
        ),
        enabled: resource.category !== 'disabled',
        install_deploys: [],
      }))
    )
  }
  if (hasDrift) {
    components.push({
      ...installComponent(
        'cmp-retired-worker',
        'retired-worker',
        'helm_chart',
        'inactive',
        ''
      ),
      enabled: true,
      install_deploys: [],
    })
    components.push({
      ...installComponent(
        'cmp-reports',
        'reports',
        'helm_chart',
        'disabled',
        ''
      ),
      enabled: false,
      install_deploys: [],
    })
    components.push({
      ...installComponent('cmp-search', 'search', 'helm_chart', 'inactive', ''),
      enabled: true,
      install_deploys: [],
    })
  }
  if (alignment === 'paged') {
    const extras = Array.from({ length: 12 }, (_, index) => {
      const id = `cmp-service-${String(index + 1).padStart(2, '0')}`
      return [id, `service-${String(index + 1).padStart(2, '0')}`] as const
    })
    overview.alignment!.resources!.push(
      ...extras.map(([id, name]) =>
        alignmentResource({
          id: `ic-${id}`,
          name,
          component_id: id,
          component_type: 'helm_chart',
          expected_config_id: APP_CONFIG_ID,
        })
      ),
      // Still installed, but no longer in the selected app config: the list
      // endpoint never returns it, so it must not count toward pages.
      alignmentResource({
        id: 'ic-cmp-legacy',
        name: 'legacy',
        component_id: 'cmp-legacy',
        category: 'unknown',
        reason: 'component is not part of the install’s selected app config',
      })
    )
    components.push(
      ...extras.map(([id, name]) =>
        installComponent(id, name, 'helm_chart', 'active', '')
      )
    )
    // The overview lists install components oldest first; the list endpoint
    // returns them newest first.
    components.reverse()
  }
  if (alignment === 'torn-down') {
    for (const resource of overview.alignment?.resources ?? []) {
      if (resource.component_id === 'cmp-ledger-image') {
        resource.category = 'torn-down'
        resource.status = 'inactive'
        resource.reason = 'component was explicitly torn down on this install'
      }
      if (resource.component_id === 'cmp-scheduler') {
        resource.category = 'torn-down'
        resource.status = 'inactive'
      }
    }
    for (const component of components) {
      if (component.component_id === 'cmp-scheduler') {
        component.status = 'inactive'
        component.status_v2.status = 'inactive'
        component.install_deploys = []
      }
    }
  }
  const install = {
    ...viewInstall(
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
    ),
    ...(managedByConfig
      ? { metadata: { managed_by: 'nuon/cli/install-config' } }
      : {}),
  }

  const resolution = resolveDriftFixture(
    overview,
    components,
    resolutionScenario
  )
  return withChrome(install, (url, init) => {
    const resolved = resolution(url, init)
    if (resolved) return resolved
    const path = url.pathname
    if (path === `/v1/installs/${VIEW_INSTALL_ID}/overview`) {
      if (alignment === 'error') return fail('unable to get install overview')
      return ok(overview)
    }
    if (
      alignment === 'component-history' &&
      /\/components\/cmp-(api|worker|ledger)\/deploys$/.test(path)
    ) {
      return ok([], true)
    }
    const inputAttemptId = path.match(
      /\/(?:workflows|deployments)\/(wf-input-(?:ic-cmp-webhooks|enable-search))$/
    )?.[1]
    if (inputAttemptId) {
      const enabling = inputAttemptId === 'wf-input-enable-search'
      const title = enabling ? 'Enable search' : 'Update webhook inputs'
      const type = enabling ? 'component_enabled' : 'input_update'
      const summary = enabling
        ? 'Enablement was cancelled before search was deployed.'
        : 'Input changes were saved but the deployment was cancelled.'
      if (path.includes('/deployments/')) {
        return ok({
          id: inputAttemptId,
          type,
          title,
          summary,
          status: 'cancelled',
          workflow: { id: inputAttemptId, name: title, type },
          affected_resources: {
            components: [enabling ? 'search' : 'webhooks'],
            images: [],
          },
          change_groups: [],
        })
      }
      return ok({
        id: inputAttemptId,
        name: title,
        type,
        status: {
          status: 'cancelled',
          status_human_description: summary,
        },
        finished: true,
        created_at: new Date(Date.now() - 3_600_000).toISOString(),
        steps: [],
      })
    }
    if (
      path === `/v1/installs/${VIEW_INSTALL_ID}/components/cmp-reports/toggle`
    ) {
      return { status: 201, body: { workflow_id: 'wf-open-enable-reports' } }
    }
    if (
      path === `/v1/installs/${VIEW_INSTALL_ID}/components/cmp-reports/deploys`
    ) {
      return ok([], true)
    }
    const openWorkflowId = path.match(
      /\/(?:deployments|workflows)\/(wf-open-[^/]+)$/
    )?.[1]
    if (openWorkflowId) {
      if (path.startsWith('/v1/workflows/')) {
        return ok({ ...deploymentWorkflow('in-progress'), id: openWorkflowId })
      }
      return ok({ ...deploymentRecord('in-progress'), id: openWorkflowId })
    }
    if (path.includes('/approvals/') && path.endsWith('/contents')) {
      const resource = overview.alignment?.resources?.find((item) =>
        path.includes(`/approvals/${item.plan?.approval_id}/`)
      )
      if (resource)
        return { body: resourcePlanFixture(resource).contents, delayMs: 800 }
    }
    if (path === `/v1/apps/${VIEW_APP_ID}/branches/br-1`) return ok(branch)
    if (path === `/v1/apps/${VIEW_APP_ID}/configs`) return ok([config], true)
    if (path.includes('/configs/') && path.includes('/apps/')) {
      if (state === 'provisioning') return pendingReply()
      if (state === 'degraded') {
        return fail('The stack config could not be read from the app branch.')
      }
      return ok({
        ...config,
        input: {
          inputs: hasDrift
            ? [
                {
                  id: 'input-reports-enabled',
                  name: overrideName('enabled', 'reports'),
                  default: 'false',
                  index: 0,
                },
              ]
            : [],
        },
        component_config_connections: [
          ...config.component_config_connections,
          ...(hasDrift
            ? [
                {
                  id: 'ccc-reports',
                  version: 1,
                  component_id: 'cmp-reports',
                  type: 'helm_chart',
                  toggleable: true,
                  default_enabled: false,
                  helm: {
                    chart_name: 'reports',
                    chart_version: '1.0.0',
                    release_name: 'reports',
                    namespace: 'payments',
                  },
                },
                {
                  id: 'ccc-search',
                  version: 1,
                  component_id: 'cmp-search',
                  type: 'helm_chart',
                  toggleable: true,
                  default_enabled: false,
                  helm: {
                    chart_name: 'search',
                    chart_version: '1.0.0',
                    release_name: 'search',
                    namespace: 'payments',
                  },
                },
              ]
            : []),
        ],
      })
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
      const id = path.split('/components/')[1]
      if (id) return ok(components.find((row) => row.component_id === id) ?? {})
      const types = url.searchParams.get('types') ?? ''
      const query = (url.searchParams.get('q') ?? '').toLowerCase()
      const ids = url.searchParams.get('ids')?.split(',')
      const filtered = components.filter((row) => {
        if (ids && !ids.includes(row.id)) return false
        if (types && !types.split(',').includes(row.component.type))
          return false
        if (query && !row.component.name.includes(query)) return false
        return true
      })
      const offset = Number(url.searchParams.get('offset') ?? 0)
      const limit = Number(url.searchParams.get('limit') ?? 10)
      return {
        body: filtered.slice(offset, offset + limit),
        paginated: { hasNext: filtered.length > offset + limit, offset, limit },
      }
    }
    if (path === '/v1/builds') {
      const id = url.searchParams.get('component_id')
      if (alignment === 'component-history') {
        const resource = overview.alignment?.resources?.find(
          (resource) => resource.component_id === id
        )
        return ok(
          resource?.expected_build_id
            ? [
                {
                  id: resource.expected_build_id,
                  component_id: id,
                  status: 'active',
                  status_v2: { status: 'active' },
                  component_config_connection_id: `ccc-${id}`,
                  component_config_connection: {
                    id: `ccc-${id}`,
                    component_id: id,
                    app_config_id: resource.expected_config_id,
                  },
                  created_at: '2026-10-04T12:00:00Z',
                },
              ]
            : [],
          true
        )
      }
      const row = resourceComponents(
        state === 'degraded' ? 'degraded' : 'active'
      ).find((item) => item.component_id === id)
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
      ? {
          app_branch: undefined,
          app_branch_id: undefined,
          app_config_id: undefined,
        }
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
  counts: {
    create: number
    update: number
    delete: number
    replace: number
    noop: number
  },
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

const deploymentRecord = (status: string, inputUpdate = false) => ({
  id: 'wf-deploy-1',
  type: 'app_branch_update',
  status,
  created_at: '2026-10-02T16:00:00Z',
  title: inputUpdate ? 'Updated input' : 'Update production install',
  summary: inputUpdate
    ? 'Updated install input values.'
    : 'Apply the latest app template to this install.',
  workflow: {
    id: 'wf-deploy-1',
    name: inputUpdate ? 'Updated input' : 'Update production install',
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

const deploymentWorkflow = (
  state: 'in-progress' | 'awaiting' | 'succeeded' | 'failed' | 'input-update'
) => {
  if (state === 'input-update') {
    return {
      id: 'wf-deploy-1',
      name: 'Updated input',
      type: 'app_branch_config_update',
      created_at: '2026-10-02T16:00:00Z',
      finished: true,
      status: {
        status: 'success',
        status_human_description: 'Install inputs updated',
      },
      steps: [syncStep],
      metadata: {
        changed_input_values: JSON.stringify({
          api_replicas: { old: '2', new: '4' },
          region: { old: 'us-east-1', new: 'us-west-2' },
        }),
      },
    }
  }

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

const templateComparison = {
  head_sha: 'abc123def456',
  base_sha: 'fedcba654321',
  head_run: {
    id: 'run-main-42',
    created_at: '2026-10-02T16:00:00Z',
    app_config_id: 'cfg-12',
    vcs_connection_commit: {
      sha: 'abc123def456',
      message: 'Bump the api chart and add a worker queue',
      author_name: 'Avery Chen',
    },
  },
  base_run: {
    id: 'run-main-41',
    app_config_id: 'cfg-11',
    vcs_connection_commit: { sha: 'fedcba654321' },
  },
  config_diff_content: {
    additions: 1,
    removals: 0,
    changed: 2,
    sections: [
      {
        name: 'Stack',
        additions: 0,
        removals: 0,
        changed: 1,
        entries: [
          {
            op: 'change',
            name: 'region',
            description: "'us-east-1' -> 'us-west-2'",
          },
        ],
      },
      {
        name: 'Components',
        additions: 1,
        removals: 0,
        changed: 1,
        entries: [
          { op: 'change', name: 'api', description: 'Chart updated' },
          { op: 'add', name: 'worker', description: 'Queue worker added' },
        ],
      },
    ],
  },
}

const templateConfigDiff = {
  config_id: 'cfg-12',
  old_config_id: 'cfg-11',
  summary: {
    has_changed: true,
    added: 1,
    removed: 0,
    changed: 2,
    unchanged: 4,
  },
  changed: 'stack, components',
  diff: {
    key: 'app_config',
    children: [
      {
        key: 'stack',
        diff: {
          op: 'change',
          diff: 'stack.toml',
          before: 'name = "payments"\nregion = "us-east-1"\n',
          after: 'name = "payments"\nregion = "us-west-2"\n',
        },
        children: [
          {
            key: 'region',
            diff: { op: 'change', diff: "'us-east-1' -> 'us-west-2'" },
          },
        ],
      },
      {
        key: 'components',
        children: [
          {
            key: 'component.api',
            children: [
              {
                key: 'type',
                diff: { op: 'noop', diff: "'helm_chart' -> 'helm_chart'" },
              },
              {
                key: 'chart_version',
                diff: { op: 'change', diff: "'1.2.0' -> '1.4.0'" },
              },
              {
                key: 'values.yaml',
                diff: {
                  op: 'change',
                  diff: 'values.yaml',
                  before: 'replicaCount: 2\nimage:\n  tag: "1.2.0"\n',
                  after: 'replicaCount: 3\nimage:\n  tag: "1.4.0"\n',
                },
              },
            ],
          },
          {
            key: 'component.worker',
            children: [
              {
                key: 'type',
                diff: {
                  op: 'add',
                  diff: "'' -> 'terraform_module'",
                },
              },
              {
                key: 'worker.tf',
                diff: {
                  op: 'add',
                  diff: 'worker.tf',
                  after:
                    'resource "aws_sqs_queue" "jobs" {\n  name = "payments-jobs"\n}\n',
                },
              },
            ],
          },
        ],
      },
    ],
  },
}

export const deploymentDetailFixture = (
  state: 'in-progress' | 'awaiting' | 'succeeded' | 'failed' | 'input-update'
): TFixture =>
  withChrome(viewInstall(), (url) => {
    const path = url.pathname
    const record = deploymentRecord(
      state === 'awaiting'
        ? 'pending'
        : state === 'succeeded' || state === 'input-update'
          ? 'success'
          : state === 'failed'
            ? 'error'
            : 'in-progress',
      state === 'input-update'
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
    if (path.includes('/comparison')) return ok(templateComparison)
    if (path.includes('/configs/') && path.endsWith('/diff'))
      return ok(templateConfigDiff)
    if (path.endsWith('/configs/cfg-12'))
      return ok({ id: 'cfg-12', version: 12 })
    if (path.endsWith('/configs/cfg-11'))
      return ok({ id: 'cfg-11', version: 11 })
    if (path.endsWith('/runs/run-main-42/builds'))
      return ok([
        {
          id: 'bld-api',
          component_id: 'cmp-api',
          component_name: 'api',
          status: 'success',
          change_reason: 'config_changed',
        },
      ])
    if (path.endsWith('/runs/run-main-42'))
      return ok({
        id: 'run-main-42',
        steps: [
          {
            name: 'Build components',
            status: {
              status: 'success',
              metadata: {
                builds: [
                  {
                    component_id: 'cmp-api',
                    component_name: 'api',
                    component_type: 'helm_chart',
                    status: 'success',
                    change_reason: 'config_changed',
                  },
                ],
              },
            },
          },
        ],
      })
    return undefined
  })
