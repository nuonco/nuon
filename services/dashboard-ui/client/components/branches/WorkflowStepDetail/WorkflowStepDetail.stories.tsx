export default {
  title: 'Branches/WorkflowStepDetail',
}

import { WorkflowStepDetail } from './WorkflowStepDetail'

const noop = () => {}

const mockStep = {
  id: 'step-abc123',
  name: 'Deploy to staging',
  status: { status: 'in-progress', status_human_description: 'Waiting for pods to be ready' },
  group_idx: 1,
  idx: 1,
  execution_type: 'deploy',
  retryable: true,
  started_at: '2024-06-15T10:30:00Z',
  install_workflow_id: 'wf-xyz789',
} as any

export const Default = () => (
  <WorkflowStepDetail step={mockStep} onClose={noop} />
)

export const Completed = () => (
  <WorkflowStepDetail
    step={{
      ...mockStep,
      status: { status: 'success', status_human_description: 'Deployment completed' },
      finished_at: '2024-06-15T10:35:00Z',
      execution_time: 300000000000,
    }}
    onClose={noop}
  />
)

export const Failed = () => (
  <WorkflowStepDetail
    step={{
      ...mockStep,
      status: { status: 'error', status_human_description: 'Pod CrashLoopBackOff' },
      finished_at: '2024-06-15T10:32:00Z',
      execution_time: 120000000000,
      retryable: false,
    }}
    onClose={noop}
  />
)

export const CommitStepShell = () => (
  <WorkflowStepDetail
    step={{
      id: 'step-commit',
      name: 'Commit',
      group_idx: 0,
      execution_type: 'system',
      status: {
        status: 'success',
        metadata: {
          commit_sha: 'a1b2c3d4e5f6',
          commit_message: 'feat: add deployment plan editor',
          author_name: 'Ada Lovelace',
          branch: 'feature/deploy-plans',
          base_branch: 'main',
          files_changed: 2,
          additions: 120,
          deletions: 8,
        },
      },
    } as any}
    onClose={noop}
  />
)

export const BuildStepShell = () => (
  <WorkflowStepDetail
    step={{
      id: 'step-build',
      name: 'Build components',
      group_idx: 1,
      execution_type: 'system',
      status: {
        status: 'success',
        metadata: {
          builds: [
            { component_id: 'c1', component_name: 'api', status: 'success', cache_status: 'cache hit', duration: 2.4 },
            { component_id: 'c2', component_name: 'web', status: 'success', cache_status: 'no cache', duration: 41.7 },
          ],
        },
      },
    } as any}
    onClose={noop}
  />
)

export const Minimal = () => (
  <WorkflowStepDetail
    step={{ id: 'step-min', name: 'Basic step', status: { status: 'pending' } } as any}
    onClose={noop}
  />
)
