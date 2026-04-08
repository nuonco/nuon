export default {
  title: 'Workflows/StepDetails/ActionRunHeader',
}

import { ActionRunHeader, ActionRunHeaderSkeleton } from './ActionRunHeader'

const mockActionRun = {
  id: 'run-1',
  config: { action_workflow_id: 'action-1' },
} as any

const mockStep = {
  owner_id: 'install-1',
} as any

export const Default = () => (
  <div className="p-4">
    <ActionRunHeader
      actionRun={mockActionRun}
      isAdhoc={false}
      step={mockStep}
      orgId="org-1"
    />
  </div>
)

export const Adhoc = () => (
  <div className="p-4">
    <ActionRunHeader
      actionRun={mockActionRun}
      isAdhoc
      orgId="org-1"
    />
  </div>
)

export const Loading = () => (
  <div className="p-4">
    <ActionRunHeaderSkeleton />
  </div>
)
