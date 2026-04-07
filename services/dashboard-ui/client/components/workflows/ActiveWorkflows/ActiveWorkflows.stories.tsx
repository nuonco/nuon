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

export const Default = () => (
  <div className="max-w-2xl p-4">
    <ActiveWorkflows
      orgId="org-1"
      workflows={mockWorkflows}
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-2xl p-4">
    <ActiveWorkflows orgId="org-1" workflows={[]} />
  </div>
)
