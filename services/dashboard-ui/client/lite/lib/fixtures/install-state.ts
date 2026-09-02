// Shaped from a real production install state, matching where its bulk
// actually sits: ~9k lines of components (one of them alone accounting for
// most of that, through deeply nested outputs), ~3.5k lines of action
// workflows, 57 inputs, and an embedded stack template that lands as a single
// line of roughly 60k characters. Around 13k lines pretty printed.
// All names and identifiers here are invented.

const id = (prefix: string, seed: number) =>
  `${prefix}_01${seed.toString(36).padStart(6, '0')}k2m4p6q8r0s2t4`

const COMPONENTS = [
  'app_nodepools',
  'certificate_wildcard_public',
  'analytics_cluster',
  'analytics_nodepools',
  'database_network',
  'database_primary',
  'database_workflows',
  'crd_analytics_operator',
  'api',
  'api_init_db',
  'api_workload_identity',
  'dashboard',
  'docs_site',
  'gateway',
  'gateway_certificates',
  'ingress_controller',
  'internal_dns',
  'log_shipper',
  'metrics_agent',
  'object_store',
  'object_store_iam',
  'queue_broker',
  'redis_cache',
  'runner_nodepools',
  'secrets_operator',
  'service_mesh',
  'ssl_policy',
  'vpc',
  'vpc_peering',
  'workflows_engine',
]

const STATUSES = ['active', 'active', 'active', 'degraded', 'provisioning']

const OUTPUT_KEYS = [
  'endpoint',
  'service_account_email',
  'bucket_name',
  'connection_name',
  'private_ip',
  'secret_name',
  'zone',
]

// The dominant component's outputs: a rendered config tree, which is where
// most of a real install state's lines come from.
const nested = (breadth: number, depth: number, seed: number): unknown => {
  if (depth === 0) {
    return seed % 3 === 0 ? seed * 17 : `value-${seed}`
  }
  return Object.fromEntries(
    Array.from({ length: breadth }, (_, index) => [
      `${['spec', 'metadata', 'config', 'limits', 'env'][index % 5]}_${index}`,
      nested(breadth, depth - 1, seed * 31 + index),
    ])
  )
}

const template = () => {
  const resource = (index: number) => ({
    Type: 'Custom::PlatformResource',
    Properties: {
      ServiceToken: `arn:aws:lambda:us-west-2:000000000000:function:provisioner-${index}`,
      ResourceName: `platform-resource-${index}`,
      Tags: [
        { Key: 'managed-by', Value: 'nuon' },
        { Key: 'environment', Value: 'production' },
        { Key: 'index', Value: String(index) },
      ],
      Policy: {
        Version: '2012-10-17',
        Statement: [
          {
            Effect: 'Allow',
            Action: ['storage.objects.get', 'storage.objects.list'],
            Resource: [`projects/example/buckets/bucket-${index}/*`],
          },
        ],
      },
    },
  })

  return JSON.stringify({
    AWSTemplateFormatVersion: '2010-09-09',
    Description: 'Install stack for an example org',
    Resources: Object.fromEntries(
      Array.from({ length: 120 }, (_, index) => [`Resource${index}`, resource(index)])
    ),
  })
}

export const installState = () => ({
  id: id('inst', 1),
  name: 'production',
  org: {
    status: 'active',
    populated: true,
    id: id('org', 2),
    name: 'example-org',
  },
  app: {
    populated: true,
    status: 'active',
    id: id('app', 3),
    name: 'platform',
    variables: {},
  },
  sandbox: {
    populated: true,
    status: 'active',
    type: 'gcp-gke',
    version: 'v0.42.1',
    outputs: Object.fromEntries(
      OUTPUT_KEYS.map((key, index) => [key, `sandbox-${key.replace(/_/g, '-')}-${index}`])
    ),
    recent_runs: [],
  },
  inputs: {
    populated: true,
    inputs: Object.fromEntries(
      Array.from({ length: 57 }, (_, index) => [
        `input_${index.toString().padStart(2, '0')}`,
        index % 4 === 0 ? String(index * 7) : `value-${index}`,
      ])
    ),
  },
  actions: {
    populated: true,
    workflows: Object.fromEntries(
      Array.from({ length: 130 }, (_, index) => [
        `workflow_${index.toString().padStart(3, '0')}`,
        {
          populated: true,
          status: STATUSES[index % STATUSES.length],
          id: id('wkf', index),
          outputs: {
            duration_ms: index * 137,
            attempts: (index % 3) + 1,
            succeeded: index % 7 !== 0,
            steps: Object.fromEntries(
              Array.from({ length: 4 }, (_, step) => [
                `step_${step}`,
                {
                  status: STATUSES[(index + step) % STATUSES.length],
                  started_at: `2026-08-${((index % 28) + 1)
                    .toString()
                    .padStart(2, '0')}T0${step}:00:00Z`,
                  duration_ms: (index + step) * 211,
                },
              ])
            ),
          },
        },
      ])
    ),
  },
  runner: {
    populated: true,
    id: id('rnr', 4),
    runner_group_id: id('rgp', 5),
    status: 'active',
  },
  components: Object.fromEntries(
    COMPONENTS.map((name, index) => [
      name,
      {
        build_id: id('bld', index + 100),
        component_id: id('cmp', index + 200),
        install_component_id: id('icp', index + 300),
        name,
        // One component carries a rendered config tree, the way a real install
        // state has a single component responsible for most of its size.
        outputs:
          index === 8
            ? (nested(9, 4, 7) as Record<string, unknown>)
            : Object.fromEntries(
                OUTPUT_KEYS.slice(0, (index % 5) + 2).map((key) => [
                  key,
                  `${name.replace(/_/g, '-')}-${key.replace(/_/g, '-')}`,
                ])
              ),
        populated: true,
        status: STATUSES[index % STATUSES.length],
      },
    ])
  ),
  domain: {
    populated: true,
    public_domain: 'example-org.nuon.run',
    internal_domain: 'example-org.internal',
  },
  cloud_account: {
    aws: null,
    azure: null,
    gcp: { project_id: 'example-project', region: 'us-west1' },
  },
  secrets: {
    populated: true,
    secrets: Object.fromEntries(
      Array.from({ length: 18 }, (_, index) => [
        `secret_${index.toString().padStart(2, '0')}`,
        { name: `platform-secret-${index}`, version: (index % 4) + 1 },
      ])
    ),
  },
  labels: { environment: 'production', tier: 'platform' },
  install: {
    populated: true,
    id: id('inst', 1),
    status: 'active',
    created_at: '2026-01-14T09:12:44Z',
    updated_at: '2026-08-30T17:03:01Z',
    history: Array.from({ length: 20 }, (_, index) => ({
      at: `2026-0${(index % 8) + 1}-1${index % 9}T12:00:00Z`,
      event: ['deploy', 'reprovision', 'sync'][index % 3],
      actor: 'service-account',
    })),
  },
  stale_at: null,
  install_stack: {
    populated: true,
    quick_link_url: 'https://console.example.com/stacks/new',
    template_url: 'https://storage.example.com/stacks/install.json',
    // One enormous single line, the way the real payload embeds its template.
    template_json: template(),
    checksum: 'sha256:0000000000000000000000000000000000000000000000000000000000000000',
    status: 'active',
    outputs: Object.fromEntries(
      Array.from({ length: 33 }, (_, index) => [
        `stack_output_${index.toString().padStart(2, '0')}`,
        index % 5 === 0
          ? { name: `nested-${index}`, value: `nested-value-${index}` }
          : `output-value-${index}`,
      ])
    ),
  },
})

export const installStateJSON = () => JSON.stringify(installState(), null, 2)
