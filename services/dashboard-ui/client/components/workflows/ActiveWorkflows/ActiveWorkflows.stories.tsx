export default {
  title: 'Workflows/ActiveWorkflows',
}

import { ActiveWorkflows } from './ActiveWorkflows'

const mockWorkflows = [
  {
    id: 'workflow-1',
    type: 'deploy_components',
    name: 'Deploy components',
    owner_id: 'install-1',
    status: { status: 'in-progress' },
    metadata: { owner_name: 'My Install' },
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
] as any

const mockMultipleWorkflows = [
  {
    id: 'workflow-1',
    type: 'deploy_components',
    name: 'Deploy components',
    owner_id: 'install-1',
    status: { status: 'in-progress' },
    metadata: { owner_name: 'My Install' },
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'workflow-2',
    type: 'provision_sandbox',
    name: 'Provision sandbox',
    owner_id: 'install-2',
    status: { status: 'in-progress' },
    metadata: { owner_name: 'Staging' },
    created_at: new Date(Date.now() - 300000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'workflow-3',
    type: 'action_workflow_run',
    name: 'Run migrations',
    owner_id: 'install-1',
    status: { status: 'in-progress' },
    metadata: { owner_name: 'My Install' },
    created_at: new Date(Date.now() - 600000).toISOString(),
    updated_at: new Date().toISOString(),
  },
] as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <ActiveWorkflows workflows={mockWorkflows} />
  </div>
)

export const MultipleWorkflows = () => (
  <div className="max-w-2xl p-4">
    <ActiveWorkflows workflows={mockMultipleWorkflows} />
  </div>
)

export const Empty = () => (
  <div className="max-w-2xl p-4">
    <ActiveWorkflows workflows={[]} />
  </div>
)
