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
  lines: IScriptLine[]
}

const info = (text: string): IScriptLine => ({ level: 'info', text })
const warn = (text: string): IScriptLine => ({ level: 'warn', text })

const helmPlan = (
  name: string,
  changes: NonNullable<IScriptStep['changes']>,
  prompt = false
): IScriptStep => ({
  id: `run-${name}-plan`,
  name: `plan ${name}`,
  componentName: name,
  approval: 'helm_approval',
  prompt,
  changes,
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

const helmApply = (name: string): IScriptStep => ({
  id: `run-${name}-apply`,
  name: `apply ${name}`,
  componentName: name,
  lines: [
    info(`Starting helm upgrade for ${name}`),
    info('Waiting for rollout to start'),
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
    lines: [
      info('Initializing terraform'),
      info('Refreshing state'),
      info('aws_security_group.ingress: Refreshing state'),
      warn('Deprecated attribute "acl" on aws_s3_bucket.assets'),
      info('Plan: 2 to add, 3 to change, 0 to destroy'),
    ],
  },
  {
    id: 'run-stack-apply',
    name: 'Apply stack',
    targetType: 'install_stack_versions',
    lines: [
      info('Applying terraform plan'),
      info('aws_security_group.ingress: Modifying'),
      info('aws_iam_role.worker: Modifying'),
      info('aws_s3_bucket.assets: Creating'),
      info('aws_s3_bucket.assets: Creation complete'),
      info('Apply complete: 2 added, 3 changed, 0 destroyed'),
    ],
  },
  helmPlan('acme-api', { create: 0, update: 2, delete: 0, noop: 5 }),
  helmApply('acme-api'),
  helmPlan('acme-worker', { create: 0, update: 1, delete: 0, noop: 4 }, true),
  helmApply('acme-worker'),
  helmPlan('acme-web', { create: 1, update: 0, delete: 0, noop: 3 }),
  helmApply('acme-web'),
]
