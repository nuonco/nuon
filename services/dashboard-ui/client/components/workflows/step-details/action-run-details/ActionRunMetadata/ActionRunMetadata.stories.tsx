export default {
  title: 'Workflows/StepDetails/ActionRunMetadata',
}

import { ActionRunMetadata, ActionRunMetadataSkeleton } from './ActionRunMetadata'

const mockActionRun = {
  id: 'run-1',
  trigger_type: 'workflow',
  triggered_by_type: 'deploy',
  status_v2: { status: 'success' },
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
} as any

export const Default = () => (
  <div className="p-4">
    <ActionRunMetadata
      actionRun={mockActionRun}
      orgId="org-1"
    />
  </div>
)

export const Loading = () => (
  <div className="p-4">
    <ActionRunMetadataSkeleton />
  </div>
)
