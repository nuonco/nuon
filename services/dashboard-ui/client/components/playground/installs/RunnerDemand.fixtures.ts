import type { TFixture } from '@/views/install/install-fixture'
import { deploymentDetailFixture } from '@/views/install/page-fixtures'

export type TRunnerDemandState =
  | 'waiting'
  | 'running'
  | 'completed'
  | 'cancelled'

export const runnerDemandFixture = (
  state: () => TRunnerDemandState,
  onCancel: () => void
): TFixture => {
  const base = deploymentDetailFixture('in-progress')
  const queuedAt = new Date(Date.now() - 2 * 86400_000).toISOString()
  let pickedUpAt: string | undefined

  return (url, init) => {
    const scenario = state()
    const waiting = scenario === 'waiting'
    const finished = scenario === 'completed' || scenario === 'cancelled'
    if (waiting || scenario === 'cancelled') pickedUpAt = undefined
    else pickedUpAt ??= new Date().toISOString()
    const status = waiting
      ? 'waiting-for-runner'
      : scenario === 'running'
        ? 'in-progress'
        : scenario === 'completed'
          ? 'success'
          : 'cancelled'
    const description = waiting
      ? 'Waiting for the runner to pick up the api deployment'
      : scenario === 'running'
        ? 'Deploying api — runner picked up the job'
        : scenario === 'completed'
          ? 'Deployment finished'
          : 'Cancelled before runner pickup'
    const common = {
      owner_id: 'inst-1',
      install_workflow_id: 'wf-deploy-1',
      execution_type: 'system',
      created_at: queuedAt,
      created_by: { email: 'ada@example.com' },
    }
    const apply = {
      ...common,
      id: 'step-api-apply',
      name: 'apply api plan',
      group_idx: 2,
      idx: 0,
      step_target_type: 'install_deploys',
      step_target_id: 'dplrunnerdemandapiexample0001',
      metadata: { component_name: 'api' },
      finished,
      started_at: pickedUpAt,
      execution_time: scenario === 'completed' ? 64_000_000_000 : 0,
      status: { status, status_human_description: description, history: [] },
    }
    const steps = [
      {
        ...common,
        id: 'step-config',
        name: 'sync install configuration',
        group_idx: 0,
        finished: true,
        execution_time: 1_800_000_000,
        status: { status: 'success' },
      },
      {
        ...common,
        id: 'step-api-plan',
        name: 'plan api',
        group_idx: 1,
        metadata: { component_name: 'api' },
        finished: true,
        execution_time: 8_200_000_000,
        status: { status: 'success' },
      },
      apply,
      {
        ...common,
        id: 'step-api-health',
        name: 'verify api health',
        group_idx: 3,
        metadata: { component_name: 'api' },
        finished: scenario === 'completed',
        execution_time: scenario === 'completed' ? 12_000_000_000 : 0,
        status: { status: scenario === 'completed' ? 'success' : 'pending' },
      },
    ]
    const workflow = {
      id: 'wf-deploy-1',
      name: 'Update production install',
      owner_id: 'inst-1',
      owner_type: 'installs',
      type: 'app_branch_config_update',
      created_at: queuedAt,
      finished,
      status: { status, status_human_description: description },
      steps,
    }
    const logStream = { id: 'logs-api', open: !finished }
    const logs = (
      waiting || scenario === 'cancelled'
        ? []
        : [
            'Runner picked up the api deployment',
            'Applying Helm release api in namespace acme',
            'Waiting for api pods to become ready',
            ...(scenario === 'completed'
              ? ['Helm release api upgraded successfully']
              : []),
          ]
    ).map((body, idx) => ({
      id: `log-${idx}`,
      timestamp: pickedUpAt,
      timestamp_date: pickedUpAt?.slice(0, 10),
      timestamp_time: pickedUpAt?.slice(11, 23),
      severity_number: 9,
      severity_text: 'Info',
      body,
      service_name: 'runner',
      scope_name: 'oteljob',
      log_stream_id: logStream.id,
      runner_job_id: 'job-api',
      log_attributes: { 'runner_job.id': 'job-api' },
      resource_attributes: {},
      scope_attributes: {},
    }))
    const path = url.pathname

    if (
      path === '/v1/workflows/wf-deploy-1/cancel' &&
      init?.method === 'POST'
    ) {
      onCancel()
      return { body: true }
    }
    if (path === '/v1/workflows/wf-deploy-1') return { body: workflow }
    if (path.includes('/workflows/wf-deploy-1/steps/')) {
      return { body: steps.find((step) => step.id === path.split('/').at(-1)) }
    }
    if (path.endsWith('/workflows/wf-deploy-1/sse')) {
      return { body: null, events: [{ name: 'workflow', data: workflow }] }
    }
    if (path === '/v1/installs/inst-1/deploys/dplrunnerdemandapiexample0001') {
      return {
        body: {
          id: apply.step_target_id,
          component_id: 'component-api',
          component_name: 'api',
          created_at: queuedAt,
          status_v2: apply.status,
          log_stream: logStream,
          runner_jobs: [
            {
              id: 'job-api',
              status: waiting
                ? 'available'
                : scenario === 'completed'
                  ? 'finished'
                  : status,
            },
          ],
        },
      }
    }
    if (path.endsWith('/log-streams/logs-api/logs/sse')) {
      return {
        body: null,
        events: [
          { name: 'message', data: logs.slice(0, 2) },
          { name: 'message', data: logs.slice(2), delay: 1500 },
          ...(finished
            ? [{ name: 'status', data: 'complete', delay: 1600 }]
            : []),
        ],
      }
    }
    if (path === '/v1/log-streams/logs-api/logs') return { body: logs }
    if (path.startsWith('/v1/log-streams/logs-api')) return { body: [] }
    if (path === '/v1/installs/inst-1') {
      const reply = base(url, init)
      return {
        ...reply,
        body: {
          ...(reply?.body as object),
          runner_status: waiting ? 'offline' : 'active',
        },
      }
    }
    const reply = base(url, init)
    if (path.endsWith('/deployments/wf-deploy-1')) {
      return {
        ...reply,
        body: {
          ...(reply?.body as object),
          status,
          created_at: queuedAt,
          affected_resources: { components: ['api'], images: [] },
        },
      }
    }
    return reply
  }
}
