import { DateTime } from 'luxon'
import type { TDeploymentBranchMeta } from '@/components/installs/DeploymentDetail/DeploymentMeta'
import type {
  TInstallDeploymentRecord,
  TWorkflow,
  TWorkflowStep,
} from '@/types'

export type TDeploymentScenario = 'active' | 'finished' | 'failed'

type TStepStatus = 'success' | 'in-progress' | 'pending' | 'error'

const SECOND = 1_000_000_000

const component = (name: string) => ({
  name,
  plan: `plan ${name}`,
  apply: `apply ${name}`,
})

const COMPONENTS = ['acme-api', 'acme-worker', 'acme-web'].map(component)

const counts = {
  stack: { create: 2, update: 3, delete: 0, replace: 0, noop: 8 },
  'acme-api': { create: 0, update: 2, delete: 0, replace: 0, noop: 5 },
  'acme-worker': { create: 0, update: 1, delete: 0, replace: 0, noop: 4 },
  'acme-web': { create: 1, update: 0, delete: 0, replace: 0, noop: 3 },
}

const planStep = (
  id: string,
  name: string,
  groupIdx: number,
  status: TStepStatus,
  componentName: string | undefined,
  planType: 'terraform_plan' | 'helm_approval',
  change: (typeof counts)['stack']
) =>
  ({
    id,
    name,
    group_idx: groupIdx,
    execution_type: 'approval',
    finished: status === 'success',
    execution_time: status === 'success' ? 6 * SECOND : undefined,
    status: { status },
    metadata: componentName ? { component_name: componentName } : {},
    approval:
      status === 'pending'
        ? undefined
        : {
            id: `approval-${id}`,
            type: planType,
            changes_state: 'ok',
            changes_create: change.create,
            changes_update: change.update,
            changes_delete: change.delete,
            changes_replace: change.replace,
            changes_noop: change.noop,
            response:
              status === 'success' ? { id: `response-${id}` } : undefined,
          },
  }) as TWorkflowStep

const applyStep = (
  id: string,
  name: string,
  groupIdx: number,
  status: TStepStatus,
  componentName: string,
  description?: string
) =>
  ({
    id,
    name,
    group_idx: groupIdx,
    execution_type: 'system',
    finished: status === 'success',
    execution_time: status === 'success' ? 14 * SECOND : undefined,
    status: { status, status_human_description: description },
    metadata: { component_name: componentName },
  }) as TWorkflowStep

const syncStep = {
  id: 'step-sync',
  name: 'Sync install configuration',
  group_idx: 0,
  execution_type: 'system',
  finished: true,
  execution_time: 2 * SECOND,
  status: { status: 'success' },
} as TWorkflowStep

const stackStep = planStep(
  'step-stack-plan',
  'Plan stack',
  1,
  'success',
  undefined,
  'terraform_plan',
  counts.stack
)

const componentSteps = (
  statuses: Record<string, [TStepStatus, TStepStatus]>,
  failure?: { component: string; description: string }
) =>
  COMPONENTS.flatMap(({ name, plan, apply }, index) => {
    const [planStatus, applyStatus] = statuses[name]
    const group = index + 2
    return [
      planStep(
        `step-${name}-plan`,
        plan,
        group,
        planStatus,
        name,
        'helm_approval',
        counts[name as keyof typeof counts]
      ),
      applyStep(
        `step-${name}-apply`,
        apply,
        group,
        applyStatus,
        name,
        failure?.component === name ? failure.description : undefined
      ),
    ]
  })

const stepsFor = (scenario: TDeploymentScenario): TWorkflowStep[] => {
  if (scenario === 'finished')
    return [
      syncStep,
      stackStep,
      ...componentSteps({
        'acme-api': ['success', 'success'],
        'acme-worker': ['success', 'success'],
        'acme-web': ['success', 'success'],
      }),
    ]
  if (scenario === 'failed')
    return [
      syncStep,
      stackStep,
      ...componentSteps(
        {
          'acme-api': ['success', 'success'],
          'acme-worker': ['success', 'error'],
          'acme-web': ['pending', 'pending'],
        },
        {
          component: 'acme-worker',
          description: 'Helm upgrade failed for acme-worker',
        }
      ),
    ]
  return [
    syncStep,
    stackStep,
    ...componentSteps({
      'acme-api': ['success', 'success'],
      'acme-worker': ['success', 'in-progress'],
      'acme-web': ['pending', 'pending'],
    }),
  ]
}

const WORKFLOW_STATUS: Record<
  TDeploymentScenario,
  { status: string; description: string }
> = {
  active: { status: 'in-progress', description: 'Deploying components' },
  finished: { status: 'success', description: 'Deployment finished' },
  failed: { status: 'error', description: 'Helm upgrade failed for acme-worker' },
}

export const DEPLOYMENT_PATH = '/deployments/wf-deploy-v2'

export const deploymentScenario = (
  scenario: TDeploymentScenario
): {
  workflow: TWorkflow
  deployment: TInstallDeploymentRecord
  branch: TDeploymentBranchMeta
} => {
  const { status, description } = WORKFLOW_STATUS[scenario]
  const createdAt = DateTime.now().minus({ minutes: 14 }).toISO() ?? ''
  const workflow = {
    id: 'wf-deploy-v2',
    name: 'Update production install',
    type: 'app_branch_config_update',
    created_at: createdAt,
    started_at: createdAt,
    finished: scenario !== 'active',
    execution_time: scenario === 'active' ? undefined : 11 * 60 * SECOND,
    status: { status, status_human_description: description },
    steps: stepsFor(scenario),
  } as TWorkflow

  const deployment = {
    id: 'wf-deploy-v2',
    type: 'app_branch_update',
    status,
    created_at: createdAt,
    title: 'Update production install',
    summary: 'Apply the latest app template to this install.',
    workflow: { id: workflow.id, name: workflow.name, type: workflow.type },
    app_branch: {
      id: 'branch-main',
      name: 'main',
      run_id: 'run-main-42',
      sha: 'abc123def456',
    },
    affected_resources: {
      stack: true,
      components: COMPONENTS.map(({ name }) => name),
      images: [],
    },
    change_groups: [],
  } as TInstallDeploymentRecord

  const branch: TDeploymentBranchMeta = {
    name: 'main',
    href: '/branches/branch-main',
    runHref: '/branches/branch-main/runs/run-main-42',
    sha: 'abc123def456',
    message: 'Apply the latest template',
    author: 'Ada',
    createdAt: DateTime.now().minus({ minutes: 16 }).toISO() ?? '',
  }

  return { workflow, deployment, branch }
}
