export default {
  title: 'Orgs/PendingApprovals',
}

import { PendingApprovals } from './PendingApprovals'

const mockApprovals = [
  {
    id: 'approval-1',
    type: 'manual',
    workflow_step: {
      name: 'apply_changes',
      owner_id: 'install-1',
      install_workflow_id: 'workflow-1',
    },
  },
] as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <PendingApprovals
      orgId="org-1"
      approvals={mockApprovals}
      activeWorkflows={[{ owner_id: 'install-1', metadata: { owner_name: 'My Install' } }] as any}
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-2xl p-4">
    <PendingApprovals orgId="org-1" approvals={[]} activeWorkflows={[]} />
  </div>
)
