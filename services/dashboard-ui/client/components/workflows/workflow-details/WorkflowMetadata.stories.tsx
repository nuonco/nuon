export default {
  title: 'Workflows/WorkflowMetadata',
}

import { WorkflowMetadata } from './WorkflowMetadata'
import type { TWorkflow } from '@/types'

const mockWorkflow = {
  id: 'inwj31srgsthldkueeo',
  type: 'provision',
  created_at: '2024-01-01T12:00:00Z',
  created_by: { email: 'nat@nuon.co' },
  status: {
    status: 'in-progress',
    created_at_ts: 1704067320,
    history: [
      { status: 'pending', created_at_ts: 1704067200 },
      { status: 'queued', created_at_ts: 1704067210 },
      {
        status: 'in-progress',
        created_at_ts: 1704067260,
        status_human_description: 'Executing step update install stack outputs',
      },
    ],
  },
} as TWorkflow

const mockFinishedWorkflow = {
  id: 'inwj31srgsthldkueeo',
  type: 'deploy_components',
  created_at: '2024-01-01T12:00:00Z',
  created_by: { email: 'nat@nuon.co' },
  finished: true,
  status: {
    status: 'success',
    created_at_ts: 1704067400,
    history: [
      { status: 'pending', created_at_ts: 1704067200 },
      { status: 'queued', created_at_ts: 1704067210 },
      { status: 'in-progress', created_at_ts: 1704067260 },
      {
        status: 'failed',
        created_at_ts: 1704067320,
        status_human_description: 'Step deploy_component failed: terraform apply exit code 1',
      },
      { status: 'in-progress', created_at_ts: 1704067350, status_human_description: 'Retrying step deploy_component' },
    ],
  },
} as TWorkflow

export const InProgress = () => <WorkflowMetadata workflow={mockWorkflow} />
export const Finished = () => <WorkflowMetadata workflow={mockFinishedWorkflow} />
