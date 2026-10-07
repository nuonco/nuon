import { expect, test } from 'bun:test'
import type {
  TInstallDeploymentRecord,
  TInstallDeploymentSummary,
  TWorkflow,
  TWorkflowStep,
} from '@/types'
import {
  deploymentChangeDescription,
  deploymentOutcomes,
  deploymentStepContext,
  deploymentSteps,
  deploymentTabOrder,
  isAwaitingDeploymentApproval,
  recoveredDeploymentResources,
  stepAffectedResources,
  summaryDeploymentEvidence,
} from './deployment-progress'

const deployment: TInstallDeploymentRecord = {
  id: 'wf-example',
  type: 'provision',
  title: 'Roll out acme',
  summary: '',
  status: 'in-progress',
  created_at: '2026-10-05T12:00:00Z',
  affected_resources: {
    stack: true,
    sandbox: true,
    components: ['api', 'worker'],
    images: [],
  },
  change_groups: [],
}
const step = (
  name: string,
  status: NonNullable<TWorkflowStep['status']>['status'],
  group: number,
  props: Partial<TWorkflowStep> = {}
): TWorkflowStep => ({
  id: `${name}-${status}-${group}`,
  name,
  group_idx: group,
  execution_type: 'system',
  status: { status },
  ...props,
})
const workflow = (
  steps: TWorkflowStep[],
  props: Partial<TWorkflow> = {}
): TWorkflow => ({ id: deployment.id, steps, ...props })
const isApi = ({ name }: { name: string }) => name === 'api'
const component = (
  name: string,
  status: NonNullable<TWorkflowStep['status']>['status'],
  group: number,
  componentName: string,
  props: Partial<TWorkflowStep> = {}
) =>
  step(name, status, group, {
    metadata: { component_name: componentName },
    ...props,
  })

test('a completed plan is not a completed resource', () => {
  const outcomes = deploymentOutcomes(
    deployment,
    workflow([
      step('await install stack', 'success', 1),
      step('provision sandbox apply plan', 'success', 2),
      component('plan api', 'success', 3, 'api', {
        execution_type: 'approval',
      }),
      component('apply api', 'not-attempted', 3, 'api'),
      component('apply worker', 'not-attempted', 4, 'worker'),
    ])
  )
  expect(outcomes.map(({ status }) => status)).toEqual([
    'success',
    'success',
    'not-started',
    'not-started',
  ])
  expect(outcomes[2].detail).toBe('Planned; not applied')
})

test('completed infrastructure does not imply pending components completed', () => {
  const outcomes = deploymentOutcomes(
    deployment,
    workflow([
      step('await install stack', 'success', 1),
      step('provision sandbox apply plan', 'success', 2),
      component('apply api', 'pending', 3, 'api'),
      component('apply worker', 'pending', 4, 'worker'),
    ])
  )
  expect(outcomes.map(({ status }) => status)).toEqual([
    'success',
    'success',
    'not-started',
    'not-started',
  ])
})

test('planned components remain visible before deploy records exist', () => {
  const outcomes = deploymentOutcomes(
    {
      ...deployment,
      affected_resources: { ...deployment.affected_resources, components: [] },
    },
    workflow([
      step('generate install stack', 'error', 1),
      component('sync and plan api', 'not-attempted', 3, 'api', {
        execution_type: 'approval',
      }),
      component('apply api', 'not-attempted', 3, 'api'),
    ])
  )
  expect(outcomes.find(({ name }) => name === 'api')).toMatchObject({
    category: 'Components',
    status: 'not-started',
    detail: 'Not started',
  })
})

test('configuration sync and skipped applies are not applied resources', () => {
  for (const apply of [
    component('sync and plan api', 'success', 3, 'api'),
    component('apply api', 'success', 3, 'api', { execution_type: 'skipped' }),
  ]) {
    expect(
      deploymentOutcomes(deployment, workflow([apply])).find(isApi)?.status
    ).toBe('unknown')
  }
})

test('image sync outcomes use the deployment component association', () => {
  const imageDeployment = {
    ...deployment,
    component_name: 'api',
    image: { repository: 'example.com/acme/api', next_tag: 'v2' },
    affected_resources: {
      components: ['api'],
      images: ['example.com/acme/api'],
    },
  }
  const outcomes = deploymentOutcomes(
    imageDeployment,
    workflow([component('sync api', 'success', 3, 'api')])
  )
  expect(outcomes.map(({ status }) => status)).toEqual(['success', 'success'])
  expect(outcomes[0]).toMatchObject({ detail: 'Image synced' })
  expect(outcomes[0].applied).toBeFalsy()
})

test('an image sync is not an apply while a later plan awaits approval', () => {
  const outcomes = deploymentOutcomes(
    { ...deployment, affected_resources: { components: [], images: [] } },
    workflow([
      component('sync api', 'success', 3, 'api'),
      component('sync and plan api', 'approval-awaiting', 3, 'api', {
        execution_type: 'approval',
      }),
    ])
  )
  expect(outcomes[0]).toMatchObject({
    status: 'not-started',
    detail: 'Not started — awaiting plan approval',
  })
})

test('a failed readiness action retains the preceding successful apply', () => {
  const outcomes = deploymentOutcomes(
    deployment,
    workflow([
      component('apply api', 'success', 3, 'api'),
      component('apply worker', 'success', 4, 'worker'),
      step('worker readiness check', 'error', 4, {
        step_target_type: 'install_action_workflow_runs',
      }),
    ])
  )
  expect(outcomes.find(({ name }) => name === 'api')?.status).toBe('success')
  expect(outcomes.find(({ name }) => name === 'worker')).toMatchObject({
    status: 'error',
    detail: 'Applied; worker readiness check failed',
  })
})

test('an approval wait is distinct from an applied change', () => {
  const awaiting = component('plan api', 'approval-awaiting', 3, 'api', {
    execution_type: 'approval',
    approval: { type: 'helm_approval' },
  })
  expect(isAwaitingDeploymentApproval(awaiting)).toBe(true)
  expect(
    deploymentOutcomes(deployment, workflow([awaiting])).find(isApi)
  ).toMatchObject({
    status: 'not-started',
    detail: 'Not started — awaiting plan approval',
  })
  expect(
    isAwaitingDeploymentApproval({
      ...awaiting,
      approval: { type: 'helm_approval', response: { type: 'approve' } },
    })
  ).toBe(false)
})

test('a failed retry followed by success reports recovery', () => {
  const steps = [
    component('apply api', 'error', 3, 'api', { idx: 100, retry_index: 0 }),
    component('apply api', 'success', 3, 'api', { idx: 101, retry_index: 1 }),
  ]
  expect(deploymentSteps(workflow(steps))).toHaveLength(1)
  expect(
    deploymentOutcomes(deployment, workflow(steps)).find(isApi)?.status
  ).toBe('success')
})

test('a failed retry round does not erase an earlier successful apply', () => {
  const run = workflow(
    [
      component('apply api', 'success', 3, 'api', {
        group_retry_idx: 0,
        retried: true,
      }),
      component('api readiness check', 'error', 3, 'api', {
        group_retry_idx: 0,
        retried: true,
      }),
      component('plan api', 'error', 3, 'api', {
        execution_type: 'approval',
        group_retry_idx: 1,
      }),
    ],
    { finished: true, status: { status: 'error' } }
  )
  const outcomes = deploymentOutcomes(deployment, run)
  expect(outcomes.find(isApi)).toMatchObject({
    status: 'error',
    applied: true,
    detail: 'Applied; plan api failed',
  })
  expect(deploymentChangeDescription(run, outcomes)).toStartWith(
    'Some changes were applied.'
  )
})

test('finished summaries distinguish unapplied, partially applied and unknown changes', () => {
  const failed = workflow(
    [
      step('generate install stack', 'error', 1),
      step('provision sandbox plan', 'not-attempted', 2),
      component('apply api', 'not-attempted', 3, 'api'),
      component('apply worker', 'not-attempted', 4, 'worker'),
    ],
    { finished: true, status: { status: 'error' } }
  )
  expect(
    deploymentChangeDescription(failed, deploymentOutcomes(deployment, failed))
  ).toStartWith('No completed resource changes were confirmed.')
  expect(
    deploymentChangeDescription(
      failed,
      deploymentOutcomes(deployment, workflow([]))
    )
  ).toContain('Check resource outcomes')
  expect(
    deploymentChangeDescription({ ...failed, plan_only: true }, [])
  ).toContain('plan-only deployment')
})

test('a failed apply does not imply that no resources changed', () => {
  const failed = workflow(
    [
      step('await install stack', 'error', 1),
      step('provision sandbox apply plan', 'not-attempted', 2),
      component('apply api', 'not-attempted', 3, 'api'),
      component('apply worker', 'not-attempted', 4, 'worker'),
    ],
    { finished: true, status: { status: 'error' } }
  )
  const description = deploymentChangeDescription(
    failed,
    deploymentOutcomes(deployment, failed)
  )
  expect(description).toContain('partial changes')
  expect(description).not.toContain('No changes applied')
})

test('only the current group retry round contributes outcomes and progress', () => {
  const steps = [
    component('plan api', 'success', 3, 'api', {
      execution_type: 'approval',
      group_retry_idx: 0,
    }),
    component('apply api', 'error', 3, 'api', { group_retry_idx: 0 }),
    component('plan api', 'success', 3, 'api', {
      execution_type: 'approval',
      group_retry_idx: 1,
    }),
    component('apply api', 'pending', 3, 'api', { group_retry_idx: 1 }),
  ]
  expect(deploymentSteps(workflow(steps))).toHaveLength(2)
  expect(
    deploymentOutcomes(deployment, workflow(steps)).find(isApi)
  ).toMatchObject({
    status: 'not-started',
    detail: 'Planned; not applied',
  })
})

test('duplicate unstarted steps are not collapsed as retries', () => {
  expect(
    deploymentSteps(
      workflow([step('apply', 'pending', 1), step('apply', 'pending', 1)])
    )
  ).toHaveLength(2)
})

test('hidden and superseded steps do not contribute progress', () => {
  expect(
    deploymentSteps(
      workflow([
        step('old apply', 'error', 1, { retried: true }),
        step('internal', 'success', 2, { execution_type: 'hidden' }),
        step('apply', 'success', 3),
      ])
    )
  ).toHaveLength(1)
})

test('plan-only workflows never report a resource as applied', () => {
  for (const status of ['success', 'approved'] as const) {
    expect(
      deploymentOutcomes(
        deployment,
        workflow(
          [
            component('plan api', status, 3, 'api', {
              execution_type: 'approval',
            }),
            component('apply api', 'success', 3, 'api', {
              execution_type: 'skipped',
            }),
          ],
          { plan_only: true }
        )
      ).find(isApi)
    ).toMatchObject({ status: 'not-started', detail: 'Planned; not applied' })
  }
})

test('missing step evidence stays unknown even when the workflow succeeded', () => {
  expect(
    deploymentOutcomes(
      deployment,
      workflow([], { finished: true, status: { status: 'success' } })
    ).every(({ status }) => status === 'unknown')
  ).toBe(true)
  expect(deploymentOutcomes(deployment, undefined)).toEqual([])
})

test('recovery notes follow the latest earlier outcome for the same resource', () => {
  const stackDeployment = {
    ...deployment,
    affected_resources: { stack: true, components: [], images: [] },
  }
  const completed = deploymentOutcomes(
    stackDeployment,
    workflow([step('await install stack', 'success', 1)])
  )
  const failed = deploymentOutcomes(
    stackDeployment,
    workflow([step('generate install stack', 'error', 1)])
  )
  expect(recoveredDeploymentResources(completed, [failed])).toEqual(['Stack'])
  expect(recoveredDeploymentResources(completed, [completed, failed])).toEqual(
    []
  )
  expect(recoveredDeploymentResources(failed, [failed])).toEqual([])
  expect(recoveredDeploymentResources(completed, [[]])).toEqual([])
})

test('completed runs lead with change summary and other runs lead with template updates', () => {
  expect(deploymentTabOrder('success')).toEqual([
    'changes',
    'workflow',
    'template',
  ])
  for (const status of ['in-progress', 'error', 'approval-awaiting'])
    expect(deploymentTabOrder(status)).toEqual([
      'template',
      'workflow',
      'changes',
    ])
})

test('failed deployment context selects the failed step and only later unstarted work', () => {
  for (const status of [
    'error',
    'failed-pending-retry',
    'approval-denied',
    'approval-expired',
  ] as const) {
    const failed = step('Apply stack policy', status, 2)
    const next = step('Verify install health', 'not-attempted', 3)
    expect(
      deploymentStepContext('error', [
        step('Earlier skipped work', 'pending', 0),
        step('Sync configuration', 'success', 1),
        failed,
        next,
      ])
    ).toEqual({ current: failed, next })
  }
})

test('running and approval contexts retain their current and next steps', () => {
  const running = step('Apply stack policy', 'in-progress', 2)
  const approval = step('Approve component plan', 'approval-awaiting', 2)
  const next = step('Verify install health', 'pending', 3)
  expect(deploymentStepContext('in-progress', [running, next])).toEqual({
    current: running,
    next,
  })
  expect(
    deploymentStepContext('in-progress', [running, approval, next])
  ).toEqual({ current: approval, next })
})

test('terminal context preserves the final step without inventing upcoming work', () => {
  const completed = step('Verify install health', 'success', 2)
  const failed = step('Verify install health', 'error', 2)
  expect(
    deploymentStepContext('success', [
      step('Sync configuration', 'success', 1),
      completed,
    ])
  ).toEqual({ current: completed, next: undefined })
  expect(deploymentStepContext('error', [failed])).toEqual({
    current: failed,
    next: undefined,
  })
})

test('missing step evidence does not misidentify successful work as the failure', () => {
  expect(deploymentStepContext('error', [])).toEqual({
    current: undefined,
    next: undefined,
  })
  expect(
    deploymentStepContext('error', [step('Sync configuration', 'success', 1)])
  ).toEqual({ current: undefined, next: undefined })
})

test('overview step summaries produce the same outcomes as full workflow steps', () => {
  const summary: TInstallDeploymentSummary = {
    id: deployment.id,
    type: 'provision',
    title: deployment.title,
    status: 'error',
    created_at: deployment.created_at,
    finished: true,
    steps: [
      {
        id: 'stack',
        name: 'await install stack',
        status: 'success',
        group_idx: 1,
        execution_type: 'system',
        step_target_type: 'install_stack_versions',
      },
      {
        id: 'sandbox',
        name: 'provision sandbox apply plan',
        status: 'success',
        group_idx: 2,
        execution_type: 'system',
      },
      {
        id: 'plan-api',
        name: 'plan api',
        status: 'approval-awaiting',
        group_idx: 3,
        execution_type: 'approval',
        component_name: 'api',
        approval_response_id: 'response-1',
      },
      {
        id: 'apply-worker',
        name: 'apply worker',
        status: 'error',
        group_idx: 4,
        execution_type: 'system',
        component_name: 'worker',
      },
    ],
  }
  const evidence = summaryDeploymentEvidence(summary)
  const resources = stepAffectedResources(evidence.steps)
  expect(resources).toEqual({
    stack: true,
    sandbox: true,
    inputs: false,
    secrets: false,
    components: ['api', 'worker'],
    images: [],
  })
  const full = workflow(
    [
      step('await install stack', 'success', 1, {
        id: 'stack',
        step_target_type: 'install_stack_versions',
      }),
      step('provision sandbox apply plan', 'success', 2, { id: 'sandbox' }),
      component('plan api', 'approval-awaiting', 3, 'api', {
        id: 'plan-api',
        execution_type: 'approval',
        approval: { response: { id: 'response-1' } },
      }),
      component('apply worker', 'error', 4, 'worker', { id: 'apply-worker' }),
    ],
    { finished: true }
  )
  expect(
    deploymentOutcomes({ affected_resources: resources }, evidence)
  ).toEqual(deploymentOutcomes({ affected_resources: resources }, full))
  expect(
    deploymentStepContext('error', deploymentSteps(evidence)).current?.id
  ).toBe('apply-worker')
})

test('auto-approved plans count as completed steps', () => {
  const outcomes = deploymentOutcomes(
    deployment,
    workflow(
      [
        step('await install stack', 'success', 1),
        step('provision sandbox apply plan', 'success', 2),
        component('plan api', 'approved', 3, 'api', {
          execution_type: 'approval',
        }),
        component('apply api', 'success', 3, 'api'),
        component('plan worker', 'approved', 4, 'worker', {
          execution_type: 'approval',
        }),
        component('apply worker', 'success', 4, 'worker'),
      ],
      { finished: true }
    )
  )
  expect(outcomes.map(({ status }) => status)).toEqual([
    'success',
    'success',
    'success',
    'success',
  ])
})

test('sandbox reprovisions and component teardowns count as applied', () => {
  const reprovision = deploymentOutcomes(
    {
      ...deployment,
      affected_resources: { sandbox: true, components: [], images: [] },
    },
    workflow(
      [
        step('reprovision sandbox plan', 'approved', 1, {
          execution_type: 'approval',
        }),
        step('reprovision sandbox apply', 'success', 1),
      ],
      { finished: true }
    )
  )
  expect(reprovision[0]).toMatchObject({
    category: 'Sandbox',
    status: 'success',
  })
  const teardown = deploymentOutcomes(
    { ...deployment, affected_resources: { components: [], images: [] } },
    workflow(
      [
        component('teardown plan api', 'approved', 1, 'api', {
          execution_type: 'approval',
        }),
        component('teardown apply plan api', 'success', 1, 'api'),
      ],
      { finished: true }
    )
  )
  expect(teardown[0]).toMatchObject({ name: 'api', status: 'success' })
})

test('a group retry keeps later groups as upcoming work', () => {
  const steps = deploymentSteps(
    workflow([
      step('apply api', 'error', 1, { idx: 100, group_retry_idx: 0 }),
      step('apply api', 'in-progress', 1, { idx: 300, group_retry_idx: 1 }),
      step('apply worker', 'pending', 2, { idx: 200 }),
    ])
  )
  expect(steps.map(({ name }) => name)).toEqual(['apply api', 'apply worker'])
  expect(deploymentStepContext('in-progress', steps).next?.name).toBe(
    'apply worker'
  )
})

test('a workflow awaiting a manual retry is not described as running', () => {
  const parked = workflow(
    [
      step('await install stack', 'success', 1),
      component('apply api', 'failed-pending-retry', 3, 'api'),
    ],
    {
      finished: false,
      status: { status: 'failed-pending-retry' },
    }
  )
  expect(
    deploymentChangeDescription(parked, deploymentOutcomes(deployment, parked))
  ).not.toContain('still running')
})

test('bulk component teardowns count as applied', () => {
  const outcomes = deploymentOutcomes(
    { ...deployment, affected_resources: { components: [], images: [] } },
    workflow(
      [
        component('teardown plan api', 'approved', 1, 'api', {
          execution_type: 'approval',
        }),
        component('teardown api', 'success', 1, 'api'),
      ],
      { finished: true }
    )
  )
  expect(outcomes[0]).toMatchObject({ name: 'api', status: 'success' })
})

test('step evidence adds resources the detail record omits and replaces a missing record', () => {
  const run = workflow(
    [
      step('reprovision sandbox plan', 'approved', 1, {
        execution_type: 'approval',
      }),
      step('reprovision sandbox apply', 'success', 1),
    ],
    { finished: true }
  )
  const withRecord = deploymentOutcomes(
    {
      ...deployment,
      affected_resources: { components: [], images: [] },
    },
    run
  )
  expect(withRecord.map(({ category }) => category)).toEqual(['Sandbox'])
  expect(deploymentOutcomes(undefined, run)).toEqual(withRecord)
})

test('a cancelled resource is cancelled, not unknown', () => {
  const run = workflow(
    [
      step('await install stack', 'success', 1),
      step('reprovision sandbox plan', 'cancelled', 2, {
        execution_type: 'approval',
      }),
      step('reprovision sandbox apply plan', 'not-attempted', 2),
    ],
    { finished: true, status: { status: 'cancelled' } }
  )
  expect(deploymentOutcomes(deployment, run).slice(0, 2)).toMatchObject([
    { category: 'Stack', status: 'success' },
    { category: 'Sandbox', status: 'cancelled', detail: 'Cancelled' },
  ])
  expect(
    deploymentStepContext('cancelled', deploymentSteps(run))
  ).toMatchObject({
    current: { name: 'reprovision sandbox plan' },
    next: { name: 'reprovision sandbox apply plan' },
  })
})

test('a no-op plan reports no changes', () => {
  const outcomes = deploymentOutcomes(
    { ...deployment, affected_resources: { components: [], images: [] } },
    workflow(
      [
        component('sync and plan api', 'auto-skipped', 1, 'api', {
          execution_type: 'approval',
        }),
        component('apply api', 'auto-skipped', 1, 'api'),
      ],
      { finished: true }
    )
  )
  expect(outcomes[0]).toMatchObject({ status: 'success', detail: 'No changes' })
})

test('an operator-skipped apply is reported as skipped', () => {
  const outcomes = deploymentOutcomes(
    { ...deployment, affected_resources: { components: [], images: [] } },
    workflow(
      [
        component('sync and plan api', 'approved', 1, 'api', {
          execution_type: 'approval',
        }),
        component('apply api', 'user-skipped', 1, 'api'),
      ],
      { finished: true }
    )
  )
  expect(outcomes[0]).toMatchObject({
    status: 'user-skipped',
    detail: 'Skipped by operator',
  })
})

test('an answered approval is in progress until the step advances', () => {
  const answered = component(
    'sync and plan api',
    'approval-awaiting',
    1,
    'api',
    {
      execution_type: 'approval',
      approval: { response: { type: 'approve' } },
    }
  )
  const steps = [answered, component('apply api', 'pending', 1, 'api')]
  expect(
    deploymentOutcomes(
      { ...deployment, affected_resources: { components: [], images: [] } },
      workflow(steps)
    )[0]
  ).toMatchObject({ status: 'in-progress' })
  expect(deploymentStepContext('in-progress', steps).current).toBe(answered)
})

test('a helm release recovery reports the release as recovered', () => {
  const outcomes = deploymentOutcomes(
    { ...deployment, affected_resources: { components: [], images: [] } },
    workflow([component('recover helm release api', 'success', 1, 'api')], {
      finished: true,
    })
  )
  expect(outcomes[0]).toMatchObject({
    status: 'success',
    detail: 'Helm release recovered',
  })
})

test('record resources without matching steps are ignored once steps exist', () => {
  const stackOnly = workflow(
    [
      step('generate install stack', 'success', 1, {
        step_target_type: 'install_stack_versions',
      }),
      step('await install stack', 'success', 1),
      step('runner healthy', 'success', 2),
    ],
    { finished: true }
  )
  expect(
    deploymentOutcomes(
      {
        ...deployment,
        affected_resources: {
          stack: true,
          sandbox: true,
          components: [],
          images: [],
        },
      },
      stackOnly
    ).map(({ category }) => category)
  ).toEqual(['Stack'])
  expect(
    deploymentOutcomes(
      {
        ...deployment,
        affected_resources: {
          stack: true,
          sandbox: true,
          components: [],
          images: [],
        },
      },
      workflow([])
    ).map(({ category }) => category)
  ).toEqual(['Stack', 'Sandbox'])
})

test('a denied plan says denied', () => {
  expect(
    deploymentOutcomes(
      { ...deployment, affected_resources: { components: [], images: [] } },
      workflow(
        [
          step('reprovision sandbox plan', 'approval-denied', 1, {
            execution_type: 'approval',
          }),
        ],
        { finished: true }
      )
    )[0]
  ).toMatchObject({
    category: 'Sandbox',
    status: 'error',
    detail: 'reprovision sandbox plan denied',
  })
})

test('queued, checking-plan and approval-retry steps are active', () => {
  for (const [status, detail] of [
    ['queued', 'apply api queued'],
    ['checking-plan', 'apply api checking plan'],
    ['approval-retry', 'apply api re-planning'],
  ] as const) {
    const active = component('apply api', status as never, 1, 'api')
    const steps = [active, component('verify health api', 'pending', 1, 'api')]
    expect(
      deploymentOutcomes(
        { ...deployment, affected_resources: { components: [], images: [] } },
        workflow(steps)
      )[0]
    ).toMatchObject({ status: 'in-progress', detail })
    expect(deploymentStepContext('in-progress', steps)).toMatchObject({
      current: active,
      next: { name: 'verify health api' },
    })
  }
})

test('inputs and secrets steps produce their own outcomes', () => {
  const run = workflow(
    [
      step('runner healthy', 'success', 1),
      step('update install state inputs', 'success', 2),
      step('rotate Action Run (pre-secrets-sync)', 'success', 3),
      step('sync secrets', 'success', 4),
      step('restart_pods Action Run (post-secrets-sync)', 'error', 5),
    ],
    { finished: true }
  )
  expect(
    deploymentOutcomes(undefined, run).map(({ category, status, detail }) => ({
      category,
      status,
      detail,
    }))
  ).toEqual([
    { category: 'Inputs', status: 'success', detail: 'Completed' },
    {
      category: 'Secrets',
      status: 'error',
      detail: 'Applied; restart_pods Action Run (post-secrets-sync) failed',
    },
  ])
})
