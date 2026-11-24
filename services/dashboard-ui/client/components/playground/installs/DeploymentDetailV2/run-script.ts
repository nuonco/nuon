export type TScriptLevel = 'info' | 'warn' | 'error'

export interface IScriptLine {
  level: TScriptLevel
  text: string
}

export interface IScriptStep {
  id: string
  name: string
  componentName?: string
  targetType?: string
  approval?: 'terraform_plan' | 'helm_approval'
  prompt?: boolean
  changes?: { create: number; update: number; delete: number; noop: number }
  policy?: { warnings?: string[]; skipped?: string[] }
  failAttempts?: number
  awaitAction?: boolean
  failureLines?: IScriptLine[]
  failureDescription?: string
  batch?: number
  lines: IScriptLine[]
}

const info = (text: string): IScriptLine => ({ level: 'info', text })
const warn = (text: string): IScriptLine => ({ level: 'warn', text })
const error = (text: string): IScriptLine => ({ level: 'error', text })

const BULK_BATCH = 15
const BULK_WARN_EVERY = 41

const bulk = (label: string, count: number): IScriptLine[] =>
  Array.from({ length: count }, (_, index) =>
    (index + 1) % BULK_WARN_EVERY === 0
      ? warn(`${label} ${index + 1}: transient API error, retrying`)
      : info(`${label} ${index + 1} of ${count}`)
  )

const helmPlan = (
  name: string,
  changes: NonNullable<IScriptStep['changes']>,
  prompt = false,
  policy?: IScriptStep['policy']
): IScriptStep => ({
  id: `run-${name}-plan`,
  name: `plan ${name}`,
  componentName: name,
  approval: 'helm_approval',
  prompt,
  changes,
  policy,
  lines: [
    info(`Rendering chart values for ${name}`),
    info('Running helm diff against the live release'),
    info(`${name}: deployment image tag changed`),
    info(`${name}: configmap updated`),
    info(
      `Plan complete: ${changes.update} to update, ${changes.noop} unchanged`
    ),
  ],
})

const helmApply = (
  name: string,
  failure?: Pick<
    IScriptStep,
    'failAttempts' | 'awaitAction' | 'failureLines' | 'failureDescription'
  >,
  bulkCount = 0
): IScriptStep => ({
  id: `run-${name}-apply`,
  name: `apply ${name}`,
  componentName: name,
  batch: bulkCount ? BULK_BATCH : undefined,
  ...failure,
  lines: [
    info(`Starting helm upgrade for ${name}`),
    info('Waiting for rollout to start'),
    ...bulk(`${name}: reconciling object`, bulkCount),
    info(`${name}: 1 of 3 replicas updated`),
    info(`${name}: 2 of 3 replicas updated`),
    info(`${name}: 3 of 3 replicas updated`),
    info('Release upgraded'),
  ],
})

export const RUN_SCRIPT: IScriptStep[] = [
  {
    id: 'run-sync',
    name: 'Sync install configuration',
    lines: [
      info('Fetching install configuration'),
      info('Comparing against the last synced version'),
      info('Configuration synced'),
    ],
  },
  {
    id: 'run-stack-plan',
    name: 'Plan stack',
    targetType: 'install_stack_versions',
    approval: 'terraform_plan',
    prompt: true,
    changes: { create: 2, update: 3, delete: 0, noop: 8 },
    policy: {
      warnings: ['S3 bucket should enable versioning'],
    },
    batch: BULK_BATCH,
    lines: [
      info('Initializing terraform'),
      info('Refreshing state'),
      ...bulk('aws_resource: refreshing state', 120),
      warn('Deprecated attribute "acl" on aws_s3_bucket.assets'),
      info('Plan: 2 to add, 3 to change, 0 to destroy'),
    ],
  },
  {
    id: 'run-stack-apply',
    name: 'Apply stack',
    targetType: 'install_stack_versions',
    batch: BULK_BATCH,
    lines: [
      info('Applying terraform plan'),
      ...bulk('aws_resource: modifying', 250),
      info('aws_s3_bucket.assets: Creating'),
      info('aws_s3_bucket.assets: Creation complete'),
      info('Apply complete: 2 added, 3 changed, 0 destroyed'),
    ],
  },
  helmPlan('acme-api', { create: 0, update: 2, delete: 0, noop: 5 }, false, {
    warnings: [
      'Container image should not use the latest tag',
      'Deployment should define CPU limits',
    ],
    skipped: ['Pod security policy check was skipped'],
  }),
  helmApply('acme-api', undefined, 300),
  helmPlan('acme-worker', { create: 0, update: 1, delete: 0, noop: 4 }, true),
  helmApply(
    'acme-worker',
    {
      failAttempts: 2,
      failureDescription: 'Helm upgrade failed for acme-worker',
      failureLines: [
        warn('acme-worker: 0 of 3 replicas ready'),
        error('UPGRADE FAILED: timed out waiting for the condition'),
      ],
    },
    200
  ),
  helmPlan('acme-web', { create: 1, update: 0, delete: 0, noop: 3 }),
  helmApply(
    'acme-web',
    {
      failAttempts: 3,
      awaitAction: true,
      failureDescription: 'Helm upgrade failed for acme-web',
      failureLines: [
        warn('acme-web: 0 of 3 replicas ready'),
        error('UPGRADE FAILED: timed out waiting for the condition'),
      ],
    },
    150
  ),
  {
    id: 'run-verify',
    name: 'Verify install health',
    lines: [
      info('Checking component health'),
      info('acme-api: healthy'),
      info('acme-worker: healthy'),
      info('Install is healthy'),
    ],
  },
]

const LARGE_COMPONENTS = Array.from(
  { length: 22 },
  (_, index) => `acme-svc-${String(index + 1).padStart(2, '0')}`
)

export const LARGE_RUN_SCRIPT: IScriptStep[] = [
  RUN_SCRIPT[0],
  ...LARGE_COMPONENTS.flatMap((name) => [
    helmPlan(name, { create: 0, update: 1, delete: 0, noop: 4 }),
    helmApply(name, undefined, 60),
  ]),
  RUN_SCRIPT[RUN_SCRIPT.length - 1],
]
